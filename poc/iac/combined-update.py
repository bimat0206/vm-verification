import boto3
from botocore.exceptions import ClientError
from concurrent.futures import ThreadPoolExecutor, as_completed
import logging
import threading
import sys

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# ---------------------------
# User-defined configuration
# ---------------------------
API_ID = "hpux2uegnd"      # Your API Gateway REST API ID
STAGE_NAME = "v1"          # The stage name to deploy to
AWS_REGION = "us-east-1"   # The AWS region for your API Gateway
PROFILE = "default"        # AWS CLI profile (set to None for default)
IMAGE_TAG = "latest"       # Docker image tag to update Lambda with
FORCE_UPDATE = True        # Always apply changes
MAX_WORKERS = 10           # Maximum number of concurrent operations

class SimpleProgressTracker:
    def __init__(self, total, description="Processing"):
        self.total = total
        self.current = 0
        self.description = description
        self.lock = threading.Lock()
        self._last_printed = -1
        print(f"\n{description}: 0/{total} (0%)")
    
    def update(self, n=1):
        with self.lock:
            self.current += n
            percent = int((self.current / self.total) * 100)
            # Only print if percentage changed to avoid spam
            if percent != self._last_printed:
                print(f"{self.description}: {self.current}/{self.total} ({percent}%)")
                sys.stdout.flush()
                self._last_printed = percent
    
    def close(self):
        print(f"{self.description}: Complete ✓\n")

class AWSDeploymentManager:
    def __init__(self, region, profile=None):
        self.region = region
        self.session = boto3.Session(profile_name=profile) if profile else boto3.Session()
        self.lambda_client = self.session.client('lambda', region_name=region)
        self.ecr_client = self.session.client('ecr', region_name=region)
        self.apigateway_client = self.session.client('apigateway', region_name=region)
        self.sts_client = self.session.client('sts')

    def get_account_id(self):
        try:
            response = self.sts_client.get_caller_identity()
            return response['Account']
        except ClientError as e:
            logger.error(f"Failed to get account ID: {e}")
            raise

    def get_ecr_repositories_with_images(self):
        logger.info("Fetching ECR repositories and their image tags...")
        try:
            paginator = self.ecr_client.get_paginator('describe_repositories')
            repositories = []
            for page in paginator.paginate():
                repositories.extend(page['repositories'])
            if not repositories:
                logger.info("No ECR repositories found")
                return [], {}, set()
            repo_tags = {}
            repos_with_images = set()
            progress = SimpleProgressTracker(len(repositories), "Fetching ECR images")
            def get_repo_images(repo):
                repo_name = repo['repositoryName']
                try:
                    paginator = self.ecr_client.get_paginator('describe_images')
                    images = []
                    for page in paginator.paginate(repositoryName=repo_name):
                        images.extend(page['imageDetails'])
                    if images:
                        repos_with_images.add(repo_name)
                        all_tags = []
                        for image in images:
                            if 'imageTags' in image:
                                all_tags.extend(image['imageTags'])
                        if all_tags:
                            progress.update()
                            return repo_name, all_tags
                    progress.update()
                    return repo_name, []
                except ClientError as e:
                    logger.warning(f"Failed to get images for repository {repo_name}: {e}")
                    progress.update()
                    return repo_name, []
            with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
                future_to_repo = {executor.submit(get_repo_images, repo): repo for repo in repositories}
                for future in as_completed(future_to_repo):
                    repo_name, tags = future.result()
                    if tags:
                        repo_tags[repo_name] = tags
                        repos_with_images.add(repo_name)
            progress.close()
            logger.info(f"Found {len(repos_with_images)} repositories with images")
            return repositories, repo_tags, repos_with_images
        except ClientError as e:
            logger.error(f"Failed to fetch ECR repositories: {e}")
            raise

    def get_container_lambda_functions(self):
        logger.info("Fetching container-based Lambda functions...")
        try:
            paginator = self.lambda_client.get_paginator('list_functions')
            all_functions = []
            for page in paginator.paginate():
                all_functions.extend(page['Functions'])
            container_functions = []
            for func in all_functions:
                if func.get('PackageType') == 'Image':
                    container_functions.append(func)
                    continue
                if func.get('ImageUri'):
                    container_functions.append(func)
                    continue
                if 'Code' in func and func['Code'].get('ImageUri'):
                    func['ImageUri'] = func['Code']['ImageUri']
                    container_functions.append(func)
            logger.info(f"Found {len(container_functions)} container-based Lambda functions out of {len(all_functions)} total")
            return container_functions
        except ClientError as e:
            logger.error(f"Failed to fetch Lambda functions: {e}")
            raise

    def extract_key(self, value, prefix):
        if f"{prefix}-" in value:
            v = value.split(f"{prefix}-", 1)[1]
            v = v.rsplit('-', 1)[0]
            return v
        return value

    def find_matching_repository(self, function_name, repos_with_images):
        fn_key = self.extract_key(function_name, 'lambda')
        # Try exact match first
        for repo in repos_with_images:
            repo_key = self.extract_key(repo, 'ecr')
            if repo_key == fn_key:
                return repo
        # Try partial matches
        best_match = None
        best_score = 0
        for repo in repos_with_images:
            repo_key = self.extract_key(repo, 'ecr')
            score = 0
            if fn_key in repo_key or repo_key in fn_key:
                common_parts = set(fn_key.split('-')) & set(repo_key.split('-'))
                score = len(common_parts) * 10
                if score > best_score:
                    best_score = score
                    best_match = repo
        return best_match

    def update_container_function(self, function_name, image_uri, current_uri, force):
        if current_uri == image_uri and current_uri:
            return False, f"Function {function_name} already uses target image"
        try:
            self.lambda_client.update_function_code(
                FunctionName=function_name,
                ImageUri=image_uri
            )
            waiter = self.lambda_client.get_waiter('function_updated')
            waiter.wait(
                FunctionName=function_name,
                WaiterConfig={'Delay': 1, 'MaxAttempts': 60}
            )
            return True, f"Successfully updated {function_name} to {image_uri}"
        except ClientError as e:
            return False, f"Failed to update {function_name}: {str(e)}"

    def update_lambda_image_uris(self, image_tag, force):
        logger.info("Starting Lambda function updates...")
        repositories, repo_tags, repos_with_images = self.get_ecr_repositories_with_images()
        functions = self.get_container_lambda_functions()
        if not repos_with_images:
            logger.warning("No ECR repositories with images found. Nothing to update.")
            return 0, len(functions) if functions else 0
        if not functions:
            logger.warning("No container-based Lambda functions found. Nothing to update.")
            return 0, 0
        repo_uri_map = {repo['repositoryName']: repo['repositoryUri'] for repo in repositories}
        updated = 0
        skipped = 0
        progress = SimpleProgressTracker(len(functions), "Updating Lambda functions")
        def process_function(function):
            name = function['FunctionName']
            is_container = function.get('PackageType') == 'Image'
            current_uri = function.get('ImageUri', '')
            if not is_container:
                progress.update()
                return False, f"Function {name} is not a container Lambda"
            match_repo = self.find_matching_repository(name, repos_with_images)
            if not match_repo:
                progress.update()
                return False, f"No matching ECR repository found for {name}"
            available_tags = repo_tags.get(match_repo, [])
            if not available_tags:
                progress.update()
                return False, f"No image tags found for repository {match_repo}"
            use_tag = image_tag if image_tag in available_tags else available_tags[0]
            repo_uri = repo_uri_map[match_repo]
            new_uri = f"{repo_uri}:{use_tag}"
            # Always update if current_uri is empty or different
            result = self.update_container_function(name, new_uri, current_uri, force)
            progress.update()
            return result
        with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
            future_to_function = {executor.submit(process_function, func): func for func in functions}
            for future in as_completed(future_to_function):
                try:
                    was_updated, message = future.result()
                    if was_updated:
                        updated += 1
                    else:
                        skipped += 1
                    if "Successfully updated" in message:
                        logger.info(message)
                    elif "already uses" in message:
                        logger.debug(message)
                    else:
                        logger.warning(message)
                except Exception as e:
                    logger.error(f"Error processing function: {e}")
                    skipped += 1
        progress.close()
        logger.info(f"Lambda functions processed. Updated: {updated}, Skipped: {skipped}")
        return updated, skipped

    def deploy_api_gateway(self, api_id, stage_name):
        logger.info(f"Deploying API Gateway REST API {api_id} to stage {stage_name}...")
        try:
            deployment = self.apigateway_client.create_deployment(
                restApiId=api_id,
                stageName=stage_name,
                description='Auto deployment by enhanced script'
            )
            deployment_id = deployment['id']
            self.apigateway_client.update_stage(
                restApiId=api_id,
                stageName=stage_name,
                patchOperations=[
                    {'op': 'replace', 'path': '/deploymentId', 'value': deployment_id}
                ]
            )
            logger.info(f"API deployed successfully. Stage '{stage_name}' updated with deployment ID {deployment_id}")
            return deployment_id
        except ClientError as e:
            logger.error(f"Failed to deploy API Gateway: {e}")
            raise

    def show_summary(self):
        logger.info("=" * 60)
        logger.info("CONTAINER LAMBDA DEPLOYMENT SUMMARY")
        logger.info("=" * 60)
        try:
            repositories, repo_tags, repos_with_images = self.get_ecr_repositories_with_images()
            functions = self.get_container_lambda_functions()
            print(f"\n📦 ECR Repositories: {len(repositories)} total, {len(repos_with_images)} with images")
            for repo_name in sorted(repos_with_images):
                tags = repo_tags.get(repo_name, [])
                tag_str = f"Tags: {', '.join(tags[:3])}" + ("..." if len(tags) > 3 else "")
                print(f"  • {repo_name} ({tag_str})")
            print(f"\n⚡ Container Lambda Functions: {len(functions)} total")
            if functions:
                print(f"   Container functions found:")
                for func in functions[:5]:  # Show first 5
                    name = func['FunctionName']
                    image_uri = func.get('ImageUri', 'No ImageUri')
                    package_type = func.get('PackageType', 'Not set')
                    print(f"     • {name} (PackageType: {package_type})")
                    if len(image_uri) > 80:
                        print(f"       ImageUri: {image_uri[:80]}...")
                    else:
                        print(f"       ImageUri: {image_uri}")
                if len(functions) > 5:
                    print(f"       ... and {len(functions) - 5} more")
            matchable_functions = []
            repo_uri_map = {repo['repositoryName']: repo['repositoryUri'] for repo in repositories}
            for func in functions:
                name = func['FunctionName']
                current_uri = func.get('ImageUri', '')
                if not current_uri:
                    continue
                match_repo = self.find_matching_repository(name, repos_with_images)
                if match_repo:
                    tags = repo_tags.get(match_repo, [])
                    use_tag = IMAGE_TAG if IMAGE_TAG in tags else (tags[0] if tags else "no-tags")
                    new_uri = f"{repo_uri_map[match_repo]}:{use_tag}"
                    needs_update = current_uri != new_uri
                    matchable_functions.append((name, match_repo, use_tag, current_uri, new_uri, needs_update))
            print(f"  • With matching ECR repos: {len(matchable_functions)}")
            if matchable_functions:
                print(f"\n🔗 Function → Repository Matches:")
                updates_needed = 0
                for func_name, repo_name, use_tag, current_uri, new_uri, needs_update in sorted(matchable_functions):
                    status = "UPDATE NEEDED" if needs_update else "UP TO DATE"
                    print(f"  • {func_name} → {repo_name}:{use_tag} [{status}]")
                    if needs_update:
                        updates_needed += 1
                print(f"\n🚀 Functions needing updates: {updates_needed}/{len(matchable_functions)}")
            else:
                print(f"\n⚠️  No container functions found with matching ECR repositories")
            logger.info("=" * 60)
        except Exception as e:
            logger.error(f"Failed to generate summary: {e}")

def main():
    logger.info("Starting AWS Container Lambda deployment process...")
    try:
        manager = AWSDeploymentManager(AWS_REGION, PROFILE if PROFILE != "default" else None)
        manager.show_summary()
        # Step 1: Update Lambda functions
        updated, _ = manager.update_lambda_image_uris(IMAGE_TAG, FORCE_UPDATE)
        # Step 2: Deploy API Gateway (only if some Lambda functions were updated)
        if FORCE_UPDATE and updated > 0:
            print()
            deployment_id = manager.deploy_api_gateway(API_ID, STAGE_NAME)
            logger.info(f"Deployment completed successfully. Deployment ID: {deployment_id}")
        else:
            logger.info("No Lambda functions were updated - skipping API Gateway deployment")
        logger.info("All operations completed successfully!")
    except Exception as e:
        logger.error(f"Deployment failed: {e}")
        raise

if __name__ == "__main__":
    main()
