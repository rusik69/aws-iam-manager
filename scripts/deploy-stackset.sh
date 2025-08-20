#!/bin/bash

# AWS IAM Manager StackSet Deployment Script
# This script deploys the IAM Manager role to all accounts in the organization

set -e

# Configuration - can be overridden by environment variables or command line arguments
STACK_SET_NAME="${STACK_SET_NAME:-IAMManagerRoleStackSet}"
TEMPLATE_FILE="${TEMPLATE_FILE:-../cloudformation/iam-manager-role.yaml}"
MASTER_ACCOUNT_ID="${MASTER_ACCOUNT_ID:-257801542589}"
MASTER_USER_NAME="${MASTER_USER_NAME:-iam-manager}"
ROLE_NAME="${ROLE_NAME:-OrganizationAccountAccessRole}"
REGIONS="${REGIONS:-us-east-1}"  # Add more regions if needed, comma-separated

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to show usage information
show_usage() {
    echo "AWS IAM Manager StackSet Deployment Script"
    echo
    echo "Usage: $0 [OPTIONS] [COMMAND]"
    echo
    echo "Environment Variables (can also be set via command line):"
    echo "  STACK_SET_NAME      Name of the StackSet (default: IAMManagerRoleStackSet)"
    echo "  TEMPLATE_FILE       Path to CloudFormation template (default: ../cloudformation/iam-manager-role.yaml)"
    echo "  MASTER_ACCOUNT_ID   AWS Account ID of the master account (default: 257801542589)"
    echo "  MASTER_USER_NAME    Name of the IAM user in master account (default: iam-manager)"
    echo "  ROLE_NAME          Name of the IAM role to create (default: OrganizationAccountAccessRole)"
    echo "  REGIONS            Target regions, comma-separated (default: us-east-1)"
    echo
    echo "Options:"
    echo "  --stack-set-name NAME     Override STACK_SET_NAME"
    echo "  --template-file PATH      Override TEMPLATE_FILE"
    echo "  --master-account-id ID    Override MASTER_ACCOUNT_ID"
    echo "  --master-user-name NAME   Override MASTER_USER_NAME"
    echo "  --role-name NAME          Override ROLE_NAME"
    echo "  --regions LIST            Override REGIONS"
    echo "  -h, --help               Show this help message"
    echo
    echo "Commands:"
    echo "  (none)    Deploy or update the StackSet"
    echo "  status    Show current StackSet status"
    echo "  delete    Delete the StackSet and all instances"
    echo "  help      Show this help message"
    echo
    echo "Examples:"
    echo "  $0                                    # Deploy with default settings"
    echo "  $0 --role-name MyCustomRole           # Deploy with custom role name"
    echo "  ROLE_NAME=MyCustomRole $0             # Deploy with environment variable"
    echo "  $0 status                            # Check StackSet status"
    echo
}

# Function to parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --stack-set-name)
                STACK_SET_NAME="$2"
                shift 2
                ;;
            --template-file)
                TEMPLATE_FILE="$2"
                shift 2
                ;;
            --master-account-id)
                MASTER_ACCOUNT_ID="$2"
                shift 2
                ;;
            --master-user-name)
                MASTER_USER_NAME="$2"
                shift 2
                ;;
            --role-name)
                ROLE_NAME="$2"
                shift 2
                ;;
            --regions)
                REGIONS="$2"
                shift 2
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            status|delete|help)
                COMMAND="$1"
                shift
                ;;
            *)
                if [ -z "$COMMAND" ]; then
                    COMMAND="$1"
                fi
                shift
                ;;
        esac
    done
}

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if AWS CLI is configured
check_aws_cli() {
    print_status "Checking AWS CLI configuration..."
    
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed. Please install it first."
        exit 1
    fi
    
    if ! aws sts get-caller-identity &> /dev/null; then
        print_error "AWS CLI is not configured or credentials are invalid."
        exit 1
    fi
    
    local caller_identity=$(aws sts get-caller-identity)
    local account_id=$(echo $caller_identity | jq -r '.Account')
    local user_arn=$(echo $caller_identity | jq -r '.Arn')
    
    print_success "AWS CLI configured successfully"
    print_status "Account ID: $account_id"
    print_status "User ARN: $user_arn"
    
    if [ "$account_id" != "$MASTER_ACCOUNT_ID" ]; then
        print_warning "Current account ($account_id) doesn't match expected master account ($MASTER_ACCOUNT_ID)"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_error "Deployment cancelled by user."
            exit 1
        fi
    fi
}

# Function to check if StackSet already exists
check_stackset_exists() {
    print_status "Checking if StackSet already exists..."
    
    if aws cloudformation describe-stack-set --stack-set-name "$STACK_SET_NAME" &> /dev/null; then
        print_warning "StackSet '$STACK_SET_NAME' already exists."
        read -p "Do you want to update it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            return 0  # Update existing
        else
            print_error "Deployment cancelled by user."
            exit 1
        fi
    else
        return 1  # Create new
    fi
}

# Function to get all organization accounts
get_organization_accounts() {
    print_status "Fetching organization accounts..."
    
    local accounts=$(aws organizations list-accounts --query 'Accounts[?Status==`ACTIVE`].Id' --output text)
    
    if [ -z "$accounts" ]; then
        print_error "No accounts found in organization or insufficient permissions."
        exit 1
    fi
    
    echo $accounts
}

# Function to create StackSet
create_stackset() {
    print_status "Creating StackSet '$STACK_SET_NAME'..."
    
    aws cloudformation create-stack-set \
        --stack-set-name "$STACK_SET_NAME" \
        --template-body file://"$TEMPLATE_FILE" \
        --parameters \
            ParameterKey=MasterAccountId,ParameterValue="$MASTER_ACCOUNT_ID" \
            ParameterKey=RoleName,ParameterValue="$ROLE_NAME" \
            ParameterKey=MasterUserName,ParameterValue="$MASTER_USER_NAME" \
        --capabilities CAPABILITY_NAMED_IAM \
        --description "IAM Manager Role for cross-account access" \
        --tags \
            Key=Purpose,Value="AWS IAM Manager" \
            Key=CreatedBy,Value="StackSet Deployment Script" \
        --permission-model SELF_MANAGED
        
    if [ $? -eq 0 ]; then
        print_success "StackSet created successfully."
    else
        print_error "Failed to create StackSet."
        exit 1
    fi
}

# Function to update StackSet
update_stackset() {
    print_status "Updating StackSet '$STACK_SET_NAME'..."
    
    aws cloudformation update-stack-set \
        --stack-set-name "$STACK_SET_NAME" \
        --template-body file://"$TEMPLATE_FILE" \
        --parameters \
            ParameterKey=MasterAccountId,ParameterValue="$MASTER_ACCOUNT_ID" \
            ParameterKey=RoleName,ParameterValue="$ROLE_NAME" \
            ParameterKey=MasterUserName,ParameterValue="$MASTER_USER_NAME" \
        --capabilities CAPABILITY_NAMED_IAM \
        --description "IAM Manager Role for cross-account access - Updated"
        
    if [ $? -eq 0 ]; then
        print_success "StackSet updated successfully."
    else
        print_error "Failed to update StackSet."
        exit 1
    fi
}

# Function to deploy StackSet to accounts
deploy_to_accounts() {
    local accounts="$1"
    print_status "Deploying StackSet to accounts: $accounts"
    
    # Convert space-separated accounts to comma-separated for AWS CLI
    local account_list=$(echo $accounts | tr ' ' ',')
    
    local operation_id=$(aws cloudformation create-stack-instances \
        --stack-set-name "$STACK_SET_NAME" \
        --accounts $account_list \
        --regions $REGIONS \
        --query 'OperationId' \
        --output text)
    
    if [ $? -eq 0 ]; then
        print_success "Deployment initiated. Operation ID: $operation_id"
        print_status "Monitoring deployment progress..."
        monitor_operation "$operation_id"
    else
        print_error "Failed to deploy StackSet to accounts."
        exit 1
    fi
}

# Function to monitor StackSet operation
monitor_operation() {
    local operation_id="$1"
    local max_attempts=60  # 30 minutes with 30-second intervals
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        local status=$(aws cloudformation describe-stack-set-operation \
            --stack-set-name "$STACK_SET_NAME" \
            --operation-id "$operation_id" \
            --query 'StackSetOperation.Status' \
            --output text)
        
        case $status in
            "RUNNING")
                print_status "Operation still running... (attempt $((attempt + 1))/$max_attempts)"
                ;;
            "SUCCEEDED")
                print_success "StackSet operation completed successfully!"
                return 0
                ;;
            "FAILED"|"STOPPED")
                print_error "StackSet operation failed with status: $status"
                print_error "Check the CloudFormation console for detailed error information."
                return 1
                ;;
            *)
                print_warning "Unknown operation status: $status"
                ;;
        esac
        
        sleep 30
        ((attempt++))
    done
    
    print_warning "Operation monitoring timed out. Check the CloudFormation console for current status."
    return 1
}

# Function to list StackSet instances
list_instances() {
    print_status "Listing StackSet instances..."
    
    aws cloudformation list-stack-instances \
        --stack-set-name "$STACK_SET_NAME" \
        --query 'Summaries[*].[Account,Region,Status]' \
        --output table
}

# Main execution
main() {
    print_status "Starting AWS IAM Manager StackSet deployment..."
    print_status "Configuration:"
    print_status "  StackSet Name: $STACK_SET_NAME"
    print_status "  Template: $TEMPLATE_FILE"
    print_status "  Master Account: $MASTER_ACCOUNT_ID"
    print_status "  Master User: $MASTER_USER_NAME"
    print_status "  Role Name: $ROLE_NAME"
    print_status "  Regions: $REGIONS"
    echo
    
    # Check prerequisites
    check_aws_cli
    
    # Check if template file exists
    if [ ! -f "$TEMPLATE_FILE" ]; then
        print_error "Template file not found: $TEMPLATE_FILE"
        exit 1
    fi
    
    # Get organization accounts
    local accounts=$(get_organization_accounts)
    print_success "Found accounts: $accounts"
    echo
    
    # Check if StackSet exists and create/update accordingly
    if check_stackset_exists; then
        update_stackset
    else
        create_stackset
    fi
    
    # Deploy to accounts
    deploy_to_accounts "$accounts"
    
    # List final status
    echo
    list_instances
    
    print_success "StackSet deployment completed!"
    print_status "The IAM Manager role has been deployed to all organization accounts."
    print_status "You can now use the AWS IAM Manager application to manage users across accounts."
}

# Parse command line arguments first
parse_arguments "$@"

# Handle script commands
case "${COMMAND:-}" in
    "status")
        print_status "Checking StackSet status..."
        if aws cloudformation describe-stack-set --stack-set-name "$STACK_SET_NAME" &> /dev/null; then
            list_instances
        else
            print_warning "StackSet '$STACK_SET_NAME' does not exist."
        fi
        exit 0
        ;;
    "delete")
        print_warning "This will delete the StackSet and all instances."
        print_status "StackSet Name: $STACK_SET_NAME"
        print_status "Role Name: $ROLE_NAME"
        echo
        read -p "Are you sure? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_status "Deleting all StackSet instances..."
            aws cloudformation delete-stack-instances \
                --stack-set-name "$STACK_SET_NAME" \
                --retain-stacks false \
                --regions $REGIONS \
                --deployment-targets OrganizationalUnitIds=r-*
            
            print_status "Waiting for instances to be deleted..."
            sleep 60
            
            print_status "Deleting StackSet..."
            aws cloudformation delete-stack-set --stack-set-name "$STACK_SET_NAME"
            print_success "StackSet deletion initiated."
        fi
        exit 0
        ;;
    "help")
        show_usage
        exit 0
        ;;
    "")
        # Default: run main deployment
        main
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        print_status "Use '$0 --help' for usage information."
        exit 1
        ;;
esac