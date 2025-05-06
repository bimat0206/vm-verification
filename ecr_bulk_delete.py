import boto3

def delete_all_images(ecr, repo_name):
    try:
        paginator = ecr.get_paginator('list_images')
        for page in paginator.paginate(repositoryName=repo_name):
            image_ids = page.get('imageIds', [])
            if image_ids:
                try:
                    ecr.batch_delete_image(repositoryName=repo_name, imageIds=image_ids)
                except Exception as e:
                    print(f"Error deleting images in {repo_name}: {e}")
    except Exception as e:
        print(f"Error listing images for {repo_name}: {e}")

def delete_repository(ecr, repo_name):
    delete_all_images(ecr, repo_name)
    try:
        ecr.delete_repository(repositoryName=repo_name, force=True)
        print(f"Repository {repo_name} deleted successfully.")
    except Exception as e:
        print(f"Failed to delete repository {repo_name}: {e}")

def print_table(repos):
    # Determine column widths
    idx_width = len(str(len(repos))) + 2
    name_width = max(len("Repository Name"), *(len(r['repositoryName']) for r in repos)) + 2
    uri_width = max(len("Repository URI"), *(len(r['repositoryUri']) for r in repos)) + 2

    # Header
    header = f"{'No.':<{idx_width}}{'Repository Name':<{name_width}}{'Repository URI':<{uri_width}}"
    print(header)
    print('-' * (idx_width + name_width + uri_width))

    # Rows
    for idx, repo in enumerate(repos, 1):
        print(f"{idx:<{idx_width}}{repo['repositoryName']:<{name_width}}{repo['repositoryUri']:<{uri_width}}")

def main():
    ecr = boto3.client('ecr')
    prefix = input("Enter the prefix of the ECR repository names to search: ").strip()

    repos = []
    paginator = ecr.get_paginator('describe_repositories')
    for page in paginator.paginate():
        for repo in page['repositories']:
            if repo['repositoryName'].startswith(prefix):
                repos.append(repo)

    if not repos:
        print("No repositories found with prefix:", prefix)
        return

    print("\nThe following ECR repositories will be deleted:\n")
    print_table(repos)

    print("\n*Note: AWS ECR does not store the 'created by' information in repository metadata. "
          "To find out who created a repository, you must query CloudTrail logs if logging was enabled at the time of creation.*\n")

    confirm = input("Are you sure you want to delete these repositories? (yes/no): ")
    if confirm.lower() != 'yes':
        print("Aborted.")
        return

    for repo in repos:
        print(f"Deleting repository: {repo['repositoryName']}")
        delete_repository(ecr, repo['repositoryName'])

if __name__ == "__main__":
    main()
