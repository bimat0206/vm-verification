#!/bin/bash
# terraform-validator.sh
# Script to check AWS account and validate Terraform code
# Usage: ./terraform-validator.sh [directory] [output_format] [-p|--profile PROFILE] [-c|--config CONFIG_FILE]
# Output formats: text (default), json

# Don't exit immediately on error to allow proper error handling
set +e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
DEFAULT_CONFIG_FILE=".terraform-validator-config"
MAX_PLANS_TO_KEEP=5          # Changed from 10 to 5
MAX_STATE_BACKUPS_TO_KEEP=5  # Added for state backups
MAX_TFVARS_BACKUPS_TO_KEEP=5 # Added for tfvars backups
PLAN_COMPRESSION_DAYS=7
LOG_DIR="logs"
MAX_LOGS_TO_KEEP=30
LOG_EXTENSION=".log"

# New default directories for backups and plans
TFSTATE_BACKUP_DIR="validator/backup-state"
TFVARS_BACKUP_DIR="validator/backup-value"
PLAN_DIR="validator/plan"

# Reset color on exit
trap 'echo -e "${NC}"' EXIT

# Parse command line arguments
PROFILE="default"
DIR="."
OUTPUT_FORMAT="text"
CONFIG_FILE=""
LOG_FILE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -p|--profile)
      PROFILE="$2"
      shift 2
      ;;
    --profile=*)
      PROFILE="${1#*=}"
      shift
      ;;
    -p=*)
      PROFILE="${1#*=}"
      shift
      ;;
    -c|--config)
      CONFIG_FILE="$2"
      shift 2
      ;;
    --config=*)
      CONFIG_FILE="${1#*=}"
      shift
      ;;
    -*)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
    *)
      # First non-option argument is the directory
      if [[ "$DIR" == "." ]]; then
        DIR="$1"
      # Second non-option argument is the output format
      elif [[ "$OUTPUT_FORMAT" == "text" ]]; then
        OUTPUT_FORMAT="$1"
      fi
      shift
      ;;
  esac
done

# Load configuration file if exists
load_config() {
    local config_file="$1"
    if [ -f "$config_file" ]; then
        source "$config_file"
        output "info" "Loaded configuration from $config_file"
    fi
}

# Function to setup logging
setup_logging() {
    # Create logs directory if it doesn't exist
    mkdir -p "$LOG_DIR"
    
    # Generate timestamp for log file
    local timestamp=$(date +%Y%m%d_%H%M%S)
    LOG_FILE="$LOG_DIR/terraform_validator_${timestamp}${LOG_EXTENSION}"
    
    # Create log file with header
    {
        echo "=== Terraform Validator Run ==="
        echo "Start Time: $(date '+%Y-%m-%d %H:%M:%S')"
        echo "Directory: $DIR"
        echo "AWS Profile: $PROFILE"
        echo "Output Format: $OUTPUT_FORMAT"
        echo "=============================="
        echo ""
    } > "$LOG_FILE"
    
    # Manage old log files
    local log_count=$(ls -t "$LOG_DIR"/terraform_validator_*"$LOG_EXTENSION" 2>/dev/null | wc -l)
    if [ "$log_count" -gt "$MAX_LOGS_TO_KEEP" ]; then
        ls -t "$LOG_DIR"/terraform_validator_*"$LOG_EXTENSION" | tail -n +$((MAX_LOGS_TO_KEEP + 1)) | xargs -r rm
        output "info" "Cleaned up old log files, keeping only the $MAX_LOGS_TO_KEEP most recent"
    fi
}

# Function to log messages
log_message() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    # Log to file
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
    
    # Also log to console if not in JSON mode
    if [ "$OUTPUT_FORMAT" != "json" ]; then
        case $level in
            info)    echo -e "${BLUE}[INFO]${NC} $message" ;;
            success) echo -e "${GREEN}[SUCCESS]${NC} $message" ;;
            warning) echo -e "${YELLOW}[WARNING]${NC} $message" ;;
            error)   echo -e "${RED}[ERROR]${NC} $message" ;;
            *)       echo -e "$message" ;;
        esac
    fi
}

# Function to output in the specified format
output() {
    local level=$1
    local message=$2
    
    # Log the message
    log_message "$level" "$message"
    
    if [ "$OUTPUT_FORMAT" = "json" ]; then
        # JSON output
        echo "{\"level\":\"$level\", \"message\":\"$message\", \"timestamp\":\"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"}"
    fi
}

# Function to log command output
log_command_output() {
    local command="$1"
    local output="$2"
    
    {
        echo "=== Command: $command ==="
        echo "$output"
        echo "========================="
        echo ""
    } >> "$LOG_FILE"
}

# Function to manage plan files
manage_plan_files() {
    local plans_dir="$1"
    local max_plans="$2"
    local compression_days="$3"
    
    # Create plans directory if it doesn't exist
    mkdir -p "$plans_dir"
    
    # Compress old plans
    find "$plans_dir" -name "plan_*.tfplan" -mtime +$compression_days -exec gzip {} \;
    
    # Keep only the most recent plans
    local plan_count=$(ls -t "$plans_dir"/plan_*.tfplan* 2>/dev/null | wc -l)
    if [ "$plan_count" -gt "$max_plans" ]; then
        ls -t "$plans_dir"/plan_*.tfplan* | tail -n +$((max_plans + 1)) | xargs -r rm
        output "info" "Cleaned up old plan files, keeping only the $max_plans most recent"
    fi
}

# Function to manage state backup files
manage_state_backups() {
    local backup_dir="$1"
    local max_backups="$2"
    
    # Create backup directory if it doesn't exist
    mkdir -p "$backup_dir"
    
    # Keep only the most recent backups
    local backup_count=$(ls -t "$backup_dir"/terraform.tfstate.backup.*.md 2>/dev/null | wc -l)
    if [ "$backup_count" -gt "$max_backups" ]; then
        ls -t "$backup_dir"/terraform.tfstate.backup.*.md | tail -n +$((max_backups + 1)) | xargs -r rm
        output "info" "Cleaned up old state backups, keeping only the $max_backups most recent"
    fi
}

# Function to manage tfvars backup files
manage_tfvars_backups() {
    local backup_dir="$1"
    local max_backups="$2"
    
    # Create backup directory if it doesn't exist
    mkdir -p "$backup_dir"
    
    # Keep only the most recent backups
    local backup_count=$(ls -t "$backup_dir"/*.tfvars.backup.*.md 2>/dev/null | wc -l)
    if [ "$backup_count" -gt "$max_backups" ]; then
        ls -t "$backup_dir"/*.tfvars.backup.*.md | tail -n +$((max_backups + 1)) | xargs -r rm
        output "info" "Cleaned up old tfvars backups, keeping only the $max_backups most recent"
    fi
}

# Function to backup state file
backup_state() {
    local dir="$1"
    local backup_dir="$2"
    
    # Create backup directory if it doesn't exist
    mkdir -p "$backup_dir"
    
    if [ -f "$dir/terraform.tfstate" ]; then
        local timestamp=$(date +%Y%m%d_%H%M%S)
        local backup_file="$backup_dir/terraform.tfstate.backup.${timestamp}.md"
        
        # Create the initial part of the markdown file
        {
            echo "# Terraform State Backup"
            echo ""
            echo "**Date:** $(date '+%Y-%m-%d %H:%M:%S')"
            echo "**Directory:** $dir"
            echo "**File:** terraform.tfstate"
            echo ""
            echo "## State Content"
            echo '```json'
        } > "$backup_file"
        
        # Append the file content directly
        cat "$dir/terraform.tfstate" >> "$backup_file"
        
        # Close the code block
        echo '```' >> "$backup_file"
        
        output "info" "Created state file backup: $backup_file"
        
        # Verify the backup contains content
        if grep -q "State Content" "$backup_file" && [ $(wc -l < "$backup_file") -gt 8 ]; then
            output "success" "Verified state backup contains content"
        else
            output "warning" "State backup may not contain all content, please check $backup_file"
        fi
        
        # Manage state backup files to keep only the most recent ones
        manage_state_backups "$backup_dir" "$MAX_STATE_BACKUPS_TO_KEEP"
    else
        output "warning" "No terraform.tfstate file found in $dir"
    fi
}

# Function to backup tfvars files
backup_tfvars() {
    local dir="$1"
    local backup_dir="$2"
    
    # Create backup directory if it doesn't exist
    mkdir -p "$backup_dir"
    
    # Find all tfvars files in the directory
    local tfvars_files=($(find "$dir" -maxdepth 1 -name "*.tfvars" -o -name "*.auto.tfvars"))
    
    if [ ${#tfvars_files[@]} -eq 0 ]; then
        output "warning" "No .tfvars files found in $dir"
        return
    fi
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    for tfvars_file in "${tfvars_files[@]}"; do
        local filename=$(basename "$tfvars_file")
        local backup_file="$backup_dir/${filename}.backup.${timestamp}.md"
        
        # Create the initial part of the markdown file
        {
            echo "# Terraform Variables Backup"
            echo ""
            echo "**Date:** $(date '+%Y-%m-%d %H:%M:%S')"
            echo "**Directory:** $dir"
            echo "**File:** $filename"
            echo ""
            echo "## Variables Content"
            echo '```hcl'
        } > "$backup_file"
        
        # Append the file content directly
        cat "$tfvars_file" >> "$backup_file"
        
        # Close the code block
        echo '```' >> "$backup_file"
        
        output "info" "Created tfvars backup: $backup_file"
        
        # Verify the backup contains content
        if grep -q "Variables Content" "$backup_file" && [ $(wc -l < "$backup_file") -gt 8 ]; then
            output "success" "Verified backup contains content"
        else
            output "warning" "Backup may not contain all content, please check $backup_file"
        fi
    done
    
    # Manage tfvars backup files to keep only the most recent ones
    manage_tfvars_backups "$backup_dir" "$MAX_TFVARS_BACKUPS_TO_KEEP"
}

# Function to list available workspaces
list_workspaces() {
    local dir="$1"
    pushd "$dir" > /dev/null || return 1
    terraform workspace list
    popd > /dev/null || true
}

# Function to check AWS CLI installation
check_aws_cli() {
    output "info" "Checking AWS CLI installation..."
    if ! command -v aws &> /dev/null; then
        output "error" "AWS CLI is not installed. Please install it first."
        exit 1
    fi
    
    AWS_VERSION=$(aws --version 2>&1 | awk '{print $1}' | cut -d'/' -f2)
    output "success" "AWS CLI version $AWS_VERSION is installed"
    log_command_output "aws --version" "$AWS_VERSION"
}

# Function to check Terraform installation
check_terraform() {
    output "info" "Checking Terraform installation..."
    if ! command -v terraform &> /dev/null; then
        output "error" "Terraform is not installed. Please install it first."
        exit 1
    fi
    
    # Handle potential errors in terraform version command
    TF_VERSION_OUTPUT=$(terraform version -json 2>/dev/null)
    if [ $? -ne 0 ]; then
        # Fallback to non-json version
        TF_VERSION=$(terraform version | head -n1 | cut -d'v' -f2)
        output "success" "Terraform version $TF_VERSION is installed"
    else
        # Parse json output if available
        if command -v jq &> /dev/null; then
            TF_VERSION=$(echo "$TF_VERSION_OUTPUT" | jq -r '.terraform_version' 2>/dev/null)
            if [ $? -ne 0 ]; then
                TF_VERSION=$(terraform version | head -n1 | cut -d'v' -f2)
            fi
        else
            # If jq is not available
            TF_VERSION=$(terraform version | head -n1 | cut -d'v' -f2)
        fi
        output "success" "Terraform version $TF_VERSION is installed"
    fi
    
    log_command_output "terraform version" "$(terraform version)"
    
    # Check if Terraform version is at least 1.0.0
    if [[ $(echo "$TF_VERSION" | cut -d'.' -f1) -lt 1 ]]; then
        output "warning" "Terraform version is older than 1.0.0. Consider upgrading."
    fi
}

# Function to check AWS account
check_aws_account() {
    output "info" "Checking AWS account..."
    
    # Set profile argument if profile is specified
    local PROFILE_ARG=""
    if [[ -n "$PROFILE" && "$PROFILE" != "default" ]]; then
        PROFILE_ARG="--profile $PROFILE"
    fi
    
    # Check if AWS credentials are configured
    aws $PROFILE_ARG sts get-caller-identity &> /dev/null
    if [ $? -ne 0 ]; then
        output "error" "AWS credentials are not configured or are invalid"
        if [[ -n "$PROFILE" ]]; then
            output "error" "Profile '$PROFILE' may not exist or has invalid credentials"
        else
            output "error" "Please configure your default AWS profile or specify a profile using the -p or --profile option"
        fi
        exit 1
    fi
    
    # Get account ID
    AWS_ACCOUNT_ID=$(aws $PROFILE_ARG sts get-caller-identity --query "Account" --output text)
    output "success" "AWS Account ID: $AWS_ACCOUNT_ID"
    
    # Get user/role name
    AWS_USER=$(aws $PROFILE_ARG sts get-caller-identity --query "Arn" --output text)
    output "info" "AWS Identity: $AWS_USER"
    
    # Check AWS region
    if [[ -n "$PROFILE" ]]; then
        AWS_REGION=$(aws configure get region --profile $PROFILE)
    else
        AWS_REGION=$(aws configure get region)
    fi
    
    if [ -z "$AWS_REGION" ]; then
        output "warning" "AWS region is not set"
    else
        output "info" "AWS Region: $AWS_REGION"
    fi
    
    # Log AWS account information
    {
        echo "=== AWS Account Information ==="
        echo "Account ID: $AWS_ACCOUNT_ID"
        echo "Identity: $AWS_USER"
        echo "Region: $AWS_REGION"
        echo "=============================="
        echo ""
    } >> "$LOG_FILE"
}

# Function to validate Terraform
validate_terraform() {
    local dir=$1
    output "info" "Validating Terraform in $dir..."
    
    if [ ! -d "$dir" ]; then
        output "error" "Directory not found: $dir"
        exit 1
    fi
    
    # Check if directory contains Terraform files
    find "$dir" -maxdepth 1 -name "*.tf" | grep -q .
    if [ $? -ne 0 ]; then
        output "warning" "No Terraform files found in $dir"
        return
    fi
    
    # Navigate to the directory
    pushd "$dir" > /dev/null || {
        output "error" "Cannot change to directory: $dir"
        exit 1
    }
    
    # Initialize Terraform without backend
    output "info" "Initializing Terraform..."
    terraform init -backend=false -input=false > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        output "error" "Terraform initialization failed"
        # Show the actual error
        terraform init -backend=false -input=false
        popd > /dev/null || true
        exit 1
    fi
    
    # Validate Terraform configuration
    output "info" "Running terraform validate..."
    terraform validate > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        output "success" "Terraform validation successful"
    else
        output "error" "Terraform validation failed"
        # Show the actual errors
        terraform validate
        popd > /dev/null || true
        exit 1
    fi
    
    # Backup tfvars files
    backup_tfvars "$dir" "$TFVARS_BACKUP_DIR"
    
    # Create plan directory path
    mkdir -p "$PLAN_DIR"
    
    # Manage plan files
    manage_plan_files "$PLAN_DIR" "$MAX_PLANS_TO_KEEP" "$PLAN_COMPRESSION_DAYS"
    
    # Generate timestamp for the plan file
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    PLAN_FILE="$PLAN_DIR/plan_${TIMESTAMP}.tfplan"
    
    # Run terraform plan and save to file
    output "info" "Running terraform plan and saving to ${PLAN_FILE}..."
    terraform plan -detailed-exitcode -input=false -out="${PLAN_FILE}" > /dev/null 2>&1
    PLAN_EXIT_CODE=$?
    
    if [ $PLAN_EXIT_CODE -eq 0 ]; then
        output "success" "Terraform plan shows no changes"
    elif [ $PLAN_EXIT_CODE -eq 2 ]; then
        output "warning" "Terraform plan shows changes"
        output "info" "Plan saved to ${PLAN_FILE}"
    else
        output "error" "Terraform plan failed"
        # Run plan again to show errors
        terraform plan
        popd > /dev/null || true
        exit 1
    fi
    
    # Check for best practices
    output "info" "Checking Terraform best practices..."
    
    # Check for hardcoded credentials
    grep -r --include="*.tf" -E "(access_key|secret_key|password|token)[[:space:]]*=[[:space:]]*\"[^\"]+\"" . > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        output "warning" "Potential hardcoded credentials found in Terraform files"
        # Show the files with hardcoded credentials
        grep -r --include="*.tf" -l -E "(access_key|secret_key|password|token)[[:space:]]*=[[:space:]]*\"[^\"]+\"" .
    fi
    
    # Check for default tags
    grep -r --include="*.tf" -q "tags" .
    if [ $? -ne 0 ]; then
        output "warning" "No tags found in Terraform files. Consider adding tags for better resource management."
    fi
    
    # Return to original directory
    popd > /dev/null || true
}

# Function to review plan file
review_plan() {
    local plan_file="$1"
    if [ -f "$plan_file" ]; then
        output "info" "Reviewing plan file: $plan_file"
        terraform show -json "$plan_file" | jq '.' | less
    else
        output "error" "Plan file not found: $plan_file"
    fi
}

# Function to prompt user for next action
prompt_next_action() {
    local dir=$1
    
    # Skip interactive prompt if using JSON output
    if [ "$OUTPUT_FORMAT" = "json" ]; then
        output "info" "Interactive mode not available in JSON output format"
        return
    fi
    
    echo ""
    echo -e "${YELLOW}==== Next Steps ====${NC}"
    echo -e "1) Run terraform plan (detailed)"
    echo -e "2) Run terraform plan and apply"
    echo -e "3) Apply latest saved plan"
    echo -e "4) Review latest saved plan"
    echo -e "5) List saved plans"
    echo -e "6) Manage workspaces"
    echo -e "7) Exit"
    echo ""
    
    read -p "Select an option (1-7): " choice
    
    case $choice in
        1)
            echo ""
            output "info" "Running detailed terraform plan..."
            pushd "$dir" > /dev/null || true
            terraform plan
            popd > /dev/null || true
            ;;
        2)
            echo ""
            output "info" "Running terraform plan and apply..."
            pushd "$dir" > /dev/null || true
            terraform plan
            
            echo ""
            read -p "Do you want to apply these changes? (y/n): " apply_choice
            if [[ $apply_choice == "y" || $apply_choice == "Y" ]]; then
                output "info" "Creating state backup..."
                backup_state "$dir" "$TFSTATE_BACKUP_DIR"
                output "info" "Applying changes..."
                terraform apply
                output "success" "Terraform apply completed"
            else
                output "info" "Terraform apply cancelled"
            fi
            popd > /dev/null || true
            ;;
        3)
            echo ""
            output "info" "Applying latest saved plan..."
            pushd "$dir" > /dev/null || true
            LATEST_PLAN=$(ls -t "$PLAN_DIR"/plan_*.tfplan 2>/dev/null | head -n1)
            if [ -n "$LATEST_PLAN" ]; then
                output "info" "Creating state backup..."
                backup_state "$dir" "$TFSTATE_BACKUP_DIR"
                output "info" "Applying plan from: $LATEST_PLAN"
                terraform apply "$LATEST_PLAN"
                output "success" "Plan application completed"
            else
                output "error" "No saved plans found"
            fi
            popd > /dev/null || true
            ;;
        4)
            echo ""
            output "info" "Reviewing latest saved plan..."
            pushd "$dir" > /dev/null || true
            LATEST_PLAN=$(ls -t "$PLAN_DIR"/plan_*.tfplan 2>/dev/null | head -n1)
            if [ -n "$LATEST_PLAN" ]; then
                review_plan "$LATEST_PLAN"
            else
                output "error" "No saved plans found"
            fi
            popd > /dev/null || true
            ;;
        5)
            echo ""
            output "info" "Listing saved plans..."
            pushd "$dir" > /dev/null || true
            ls -lh "$PLAN_DIR"/plan_*.tfplan* 2>/dev/null || output "info" "No saved plans found"
            popd > /dev/null || true
            ;;
        6)
            echo ""
            output "info" "Managing workspaces..."
            pushd "$dir" > /dev/null || true
            echo -e "${YELLOW}Current workspace:${NC}"
            terraform workspace show
            echo -e "\n${YELLOW}Available workspaces:${NC}"
            list_workspaces "$dir"
            echo -e "\n${YELLOW}Options:${NC}"
            echo "1) Select workspace"
            echo "2) Create new workspace"
            echo "3) Delete workspace"
            echo "4) Back to main menu"
            read -p "Select an option (1-4): " ws_choice
            
            case $ws_choice in
                1)
                    read -p "Enter workspace name: " ws_name
                    terraform workspace select "$ws_name"
                    ;;
                2)
                    read -p "Enter new workspace name: " ws_name
                    terraform workspace new "$ws_name"
                    ;;
                3)
                    read -p "Enter workspace name to delete: " ws_name
                    terraform workspace delete "$ws_name"
                    ;;
                4)
                    ;;
                *)
                    output "error" "Invalid option"
                    ;;
            esac
            popd > /dev/null || true
            ;;
        7)
            output "info" "Exiting..."
            ;;
        *)
            output "error" "Invalid option. Exiting..."
            exit 1
            ;;
    esac
}

# Main function
main() {
    # Setup logging first
    setup_logging
    
    # Load configuration if specified
    if [ -n "$CONFIG_FILE" ]; then
        load_config "$CONFIG_FILE"
    elif [ -f "$DEFAULT_CONFIG_FILE" ]; then
        load_config "$DEFAULT_CONFIG_FILE"
    fi
    
    output "info" "Starting Terraform validation script"
    output "info" "Output format: $OUTPUT_FORMAT"
    if [[ -n "$PROFILE" ]]; then
        output "info" "Using AWS profile: $PROFILE"
    fi
    
    check_aws_cli
    check_terraform
    check_aws_account
    validate_terraform "$DIR"
    
    output "success" "Validation completed successfully"
    
    # Log completion
    {
        echo "=== Run Completed ==="
        echo "End Time: $(date '+%Y-%m-%d %H:%M:%S')"
        echo "Status: Success"
        echo "====================="
    } >> "$LOG_FILE"
    
    # Prompt for next action if validation was successful
    prompt_next_action "$DIR"
}

# Execute main function
main