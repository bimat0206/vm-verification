#!/bin/bash
# update-lambda-image-uris.sh
# Script to update Lambda function Image URIs with matching ECR repository URLs

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to display usage information
usage() {
    echo -e "${BLUE}Usage:${NC} $0 [options]"
    echo
    echo "Options:"
    echo "  -p, --profile    AWS profile to use (default: default)"
    echo "  -r, --region     AWS region (default: from AWS configuration)"
    echo "  -t, --tag        Image tag to use (default: latest)"
    echo "  -f, --force      Apply changes without dry run"
    echo "  -h, --help       Show this help message"
    echo
    exit 1
}

# Default values
PROFILE="default"
TAG="latest"
DRY_RUN=true

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -p|--profile)
            PROFILE="$2"
            shift 2
            ;;
        -r|--region)
            REGION="$2"
            shift 2
            ;;
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -f|--force)
            DRY_RUN=false
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown option:${NC} $1"
            usage
            ;;
    esac
done

# If region not provided, try to get it from AWS configuration
if [ -z "$REGION" ]; then
    REGION=$(aws configure get region --profile $PROFILE)
    if [ -z "$REGION" ]; then
        echo -e "${RED}Error:${NC} AWS region not specified and not found in AWS configuration"
        exit 1
    fi
fi

echo -e "${BLUE}AWS Profile:${NC} $PROFILE"
echo -e "${BLUE}AWS Region:${NC} $REGION"
echo -e "${BLUE}Image Tag:${NC} $TAG"
echo -e "${BLUE}Dry Run:${NC} $DRY_RUN"
echo

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error:${NC} jq is required but not installed. Please install jq first."
    exit 1
fi

# Get AWS account ID
echo -e "${BLUE}Getting AWS account ID...${NC}"
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --profile $PROFILE --query "Account" --output text)
if [ -z "$AWS_ACCOUNT_ID" ]; then
    echo -e "${RED}Error:${NC} Failed to retrieve AWS account ID"
    exit 1
fi
echo -e "${GREEN}AWS Account ID:${NC} $AWS_ACCOUNT_ID"

# Get the ECR repositories directly from AWS
echo -e "${BLUE}Getting ECR repositories from AWS...${NC}"
ECR_REPOS=$(aws ecr describe-repositories --profile $PROFILE --region $REGION --query "repositories[*]" --output json)
if [ -z "$ECR_REPOS" ] || [ "$ECR_REPOS" == "[]" ]; then
    echo -e "${RED}Error:${NC} No ECR repositories found in AWS account"
    exit 1
fi
echo -e "${GREEN}Found $(echo $ECR_REPOS | jq length) ECR repositories${NC}"

# Get repositories with images and their available tags
echo -e "${BLUE}Getting repositories with images...${NC}"
REPOS_WITH_IMAGES=()
REPO_TAGS=()

NUM_REPOS=$(echo $ECR_REPOS | jq length)
for ((i=0; i<$NUM_REPOS; i++)); do
    REPO_NAME=$(echo $ECR_REPOS | jq -r ".[$i].repositoryName")
    # Check if the repository has any images
    IMAGES=$(aws ecr describe-images --profile $PROFILE --region $REGION --repository-name $REPO_NAME --query "imageDetails[*].imageTags" --output json 2>/dev/null)
    
    if [ -n "$IMAGES" ] && [ "$IMAGES" != "[]" ] && [ "$IMAGES" != "null" ]; then
        REPOS_WITH_IMAGES+=("$REPO_NAME")
        
        # Check if the repository has the specified tag
        HAS_TAG=false
        AVAILABLE_TAGS=""
        
        if echo "$IMAGES" | jq -e "flatten | contains([\"$TAG\"])" &>/dev/null; then
            HAS_TAG=true
            AVAILABLE_TAGS="$TAG"
        else
            # Get the first available tag
            FIRST_TAG=$(echo "$IMAGES" | jq -r "flatten | .[0]" 2>/dev/null)
            if [ -n "$FIRST_TAG" ] && [ "$FIRST_TAG" != "null" ]; then
                AVAILABLE_TAGS="$FIRST_TAG"
            fi
        fi
        
        REPO_TAGS+=("$REPO_NAME:$AVAILABLE_TAGS")
    fi
done

echo -e "${GREEN}Found ${#REPOS_WITH_IMAGES[@]} repositories with images${NC}"

# Get Lambda functions
echo -e "${BLUE}Getting Lambda functions...${NC}"
LAMBDA_FUNCTIONS=$(aws lambda list-functions --profile $PROFILE --region $REGION --query "Functions[*].{Name:FunctionName,ImageUri:ImageUri}" --output json)
if [ -z "$LAMBDA_FUNCTIONS" ]; then
    echo -e "${RED}Error:${NC} No Lambda functions found"
    exit 1
fi
echo -e "${GREEN}Found $(echo $LAMBDA_FUNCTIONS | jq length) Lambda functions${NC}"

# Process updates
echo -e "${BLUE}Processing Lambda functions and ECR repositories...${NC}"
echo

# Get number of Lambda functions
NUM_FUNCTIONS=$(echo $LAMBDA_FUNCTIONS | jq length)
UPDATED=0
SKIPPED=0

for ((i=0; i<$NUM_FUNCTIONS; i++)); do
    FUNCTION_NAME=$(echo $LAMBDA_FUNCTIONS | jq -r ".[$i].Name")
    CURRENT_IMAGE_URI=$(echo $LAMBDA_FUNCTIONS | jq -r ".[$i].ImageUri")
    
    # Extract the function key from the function name
    # For example: "kootoro-dev-lambda-prepare-turn1-f6d3xl" 
    # We want to extract "prepare-turn1" as the function key
    
    # Try to match the full function name pattern
    if [[ $FUNCTION_NAME =~ -lambda-([a-zA-Z0-9-]+)-[a-z0-9]{6}$ ]]; then
        FUNCTION_KEY="${BASH_REMATCH[1]}"
    elif [[ $FUNCTION_NAME =~ -lambda-([a-zA-Z0-9-]+)$ ]]; then
        FUNCTION_KEY="${BASH_REMATCH[1]}"
    else
        # Fallback to a simpler pattern
        FUNCTION_KEY=$(echo $FUNCTION_NAME | sed -E 's/.*-lambda-//g' | sed -E 's/-[a-z0-9]{6}$//g')
    fi
    
    echo -e "${YELLOW}Function:${NC} $FUNCTION_NAME"
    echo -e "${YELLOW}Function Key:${NC} $FUNCTION_KEY"
    
    # Find matching ECR repository with images
    BEST_MATCH=""
    BEST_MATCH_URL=""
    BEST_MATCH_TAG=""
    BEST_MATCH_SCORE=0
    
    for REPO_NAME in "${REPOS_WITH_IMAGES[@]}"; do
        # Extract the relevant part from the repository name
        if [[ $REPO_NAME =~ -ecr-([a-zA-Z0-9-]+)-[a-z0-9]{6}$ ]]; then
            REPO_KEY="${BASH_REMATCH[1]}"
        elif [[ $REPO_NAME =~ -ecr-([a-zA-Z0-9-]+)$ ]]; then
            REPO_KEY="${BASH_REMATCH[1]}"
        else
            # Fallback to a simpler pattern
            REPO_KEY=$(echo $REPO_NAME | sed -E 's/.*-ecr-//g' | sed -E 's/-[a-z0-9]{6}$//g')
        fi
        
        # Get the repository URI
        REPO_URI=$(echo $ECR_REPOS | jq -r ".[] | select(.repositoryName==\"$REPO_NAME\") | .repositoryUri")
        
        # Get available tag for this repository
        AVAIL_TAG=""
        for REPO_TAG_PAIR in "${REPO_TAGS[@]}"; do
            if [[ $REPO_TAG_PAIR == "$REPO_NAME:"* ]]; then
                AVAIL_TAG="${REPO_TAG_PAIR#*:}"
                break
            fi
        done
        
        # Skip if no available tag
        if [ -z "$AVAIL_TAG" ]; then
            continue
        fi
        
        # Check if this repository is a good match for our function
        # Perfect match: function_key exactly equals repo_key
        if [ "$FUNCTION_KEY" == "$REPO_KEY" ]; then
            BEST_MATCH=$REPO_NAME
            BEST_MATCH_URL=$REPO_URI
            BEST_MATCH_TAG=$AVAIL_TAG
            BEST_MATCH_SCORE=100
            break
        fi
        
        # Good match: function_key is a prefix of repo_key
        # For example: "prepare-turn1" matches "prepare-turn1-prompt"
        if [[ "$REPO_KEY" == "$FUNCTION_KEY"* ]] && [ $BEST_MATCH_SCORE -lt 80 ]; then
            BEST_MATCH=$REPO_NAME
            BEST_MATCH_URL=$REPO_URI
            BEST_MATCH_TAG=$AVAIL_TAG
            BEST_MATCH_SCORE=80
            continue
        fi
        
        # Decent match: repo_key is a prefix of function_key
        # For example: "prepare" matches "prepare-turn1"
        if [[ "$FUNCTION_KEY" == "$REPO_KEY"* ]] && [ $BEST_MATCH_SCORE -lt 60 ]; then
            BEST_MATCH=$REPO_NAME
            BEST_MATCH_URL=$REPO_URI
            BEST_MATCH_TAG=$AVAIL_TAG
            BEST_MATCH_SCORE=60
            continue
        fi
        
        # Weak match: partial string match
        if [[ "$REPO_KEY" == *"$FUNCTION_KEY"* || "$FUNCTION_KEY" == *"$REPO_KEY"* ]] && [ $BEST_MATCH_SCORE -lt 40 ]; then
            BEST_MATCH=$REPO_NAME
            BEST_MATCH_URL=$REPO_URI
            BEST_MATCH_TAG=$AVAIL_TAG
            BEST_MATCH_SCORE=40
        fi
    done
    
    if [ -n "$BEST_MATCH" ] && [ -n "$BEST_MATCH_TAG" ]; then
        echo -e "${GREEN}Matched ECR Repository:${NC} $BEST_MATCH"
        echo -e "${GREEN}Available Image Tag:${NC} $BEST_MATCH_TAG"
        
        NEW_IMAGE_URI="${BEST_MATCH_URL}:${BEST_MATCH_TAG}"
        
        # Check if the image URI needs to be updated
        if [ "$CURRENT_IMAGE_URI" != "$NEW_IMAGE_URI" ]; then
            echo -e "${YELLOW}Current Image URI:${NC} $CURRENT_IMAGE_URI"
            echo -e "${YELLOW}New Image URI:${NC} $NEW_IMAGE_URI"
            
            if [ "$DRY_RUN" = false ]; then
                echo -e "${BLUE}Updating function...${NC}"
                if aws lambda update-function-code \
                    --profile $PROFILE \
                    --region $REGION \
                    --function-name $FUNCTION_NAME \
                    --image-uri $NEW_IMAGE_URI &>/dev/null; then
                    
                    echo -e "${GREEN}Updated successfully${NC}"
                    UPDATED=$((UPDATED+1))
                else
                    echo -e "${RED}Failed to update function${NC}"
                    SKIPPED=$((SKIPPED+1))
                fi
            else
                echo -e "${BLUE}[DRY RUN] Would update function${NC}"
                UPDATED=$((UPDATED+1))
            fi
        else
            echo -e "${GREEN}Image URI already up to date${NC}"
            SKIPPED=$((SKIPPED+1))
        fi
    else
        echo -e "${YELLOW}No matching ECR repository found with images${NC}"
        SKIPPED=$((SKIPPED+1))
    fi
    echo
done

echo -e "${GREEN}Summary:${NC}"
echo -e "  - Total Lambda functions: $NUM_FUNCTIONS"
echo -e "  - Functions updated: $UPDATED"
echo -e "  - Functions skipped: $SKIPPED"

if [ "$DRY_RUN" = true ]; then
    echo -e "${BLUE}Note:${NC} This was a dry run. No changes were made."
    echo -e "To apply changes, run with the --force flag."
fi

echo
echo -e "${GREEN}Done!${NC}"