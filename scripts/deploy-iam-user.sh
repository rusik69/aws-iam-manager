#!/bin/bash

# AWS IAM Manager - User Deployment Script
# This script creates an IAM user with the necessary permissions for AWS IAM Manager
# including StackSet deployment capabilities.

set -e  # Exit on any error

# Configuration
USER_NAME="iam-manager"
POLICY_NAME="IAMManagerCustomPolicy"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_aws_cli() {
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed. Please install it first:"
        log_info "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
        exit 1
    fi
}

check_aws_credentials() {
    log_info "Checking AWS credentials..."
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS credentials not configured or invalid."
        log_info "Please run 'aws configure' to set up your credentials."
        exit 1
    fi
    
    CALLER_IDENTITY=$(aws sts get-caller-identity)
    ACCOUNT_ID=$(echo $CALLER_IDENTITY | jq -r '.Account')
    USER_ARN=$(echo $CALLER_IDENTITY | jq -r '.Arn')
    
    log_success "Connected as: $USER_ARN"
    log_success "Account ID: $ACCOUNT_ID"
}

check_permissions() {
    log_info "Checking IAM permissions..."
    
    # Test IAM permissions
    if ! aws iam get-user --user-name $(aws sts get-caller-identity --query User.UserName --output text) &> /dev/null; then
        if ! aws iam list-users --max-items 1 &> /dev/null; then
            log_error "Insufficient IAM permissions. You need IAM admin permissions to create users and policies."
            exit 1
        fi
    fi
    
    log_success "IAM permissions verified"
}

user_exists() {
    aws iam get-user --user-name "$USER_NAME" &> /dev/null
}

create_user() {
    log_info "Creating IAM user: $USER_NAME"
    
    if user_exists; then
        log_warning "User $USER_NAME already exists"
        read -p "Do you want to continue and update the user's policies? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Operation cancelled"
            exit 0
        fi
    else
        aws iam create-user \
            --user-name "$USER_NAME" \
            --path "/" \
            --tags Key=Purpose,Value=IAMManager Key=CreatedBy,Value=deploy-script
        
        log_success "Created user: $USER_NAME"
    fi
}

attach_managed_policies() {
    log_info "Attaching managed policies..."
    
    # Required managed policies for StackSet functionality
    MANAGED_POLICIES=(
        "arn:aws:iam::aws:policy/IAMFullAccess"
        "arn:aws:iam::aws:policy/CloudFormationFullAccess" 
        "arn:aws:iam::aws:policy/AWSOrganizationsReadOnlyAccess"
    )
    
    for policy in "${MANAGED_POLICIES[@]}"; do
        policy_name=$(basename "$policy")
        log_info "Attaching policy: $policy_name"
        
        aws iam attach-user-policy \
            --user-name "$USER_NAME" \
            --policy-arn "$policy"
        
        log_success "Attached: $policy_name"
    done
}

create_custom_policy() {
    log_info "Creating custom policy: $POLICY_NAME"
    
    # Check if policy already exists
    if aws iam get-policy --policy-arn "arn:aws:iam::$ACCOUNT_ID:policy/$POLICY_NAME" &> /dev/null; then
        log_warning "Policy $POLICY_NAME already exists"
        
        # Create a new version of the policy
        log_info "Creating new policy version..."
        aws iam create-policy-version \
            --policy-arn "arn:aws:iam::$ACCOUNT_ID:policy/$POLICY_NAME" \
            --policy-document file://<(cat << 'EOF'
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "STSPermissions",
            "Effect": "Allow",
            "Action": [
                "sts:AssumeRole",
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        },
        {
            "Sid": "StackSetServicePermissions",
            "Effect": "Allow",
            "Action": [
                "iam:CreateServiceLinkedRole",
                "iam:GetRole",
                "iam:PassRole"
            ],
            "Resource": [
                "arn:aws:iam::*:role/aws-service-role/stacksets.cloudformation.amazonaws.com/AWSServiceRoleForCloudFormationStackSets*",
                "arn:aws:iam::*:role/OrganizationAccountAccessRole*"
            ]
        }
    ]
}
EOF
            ) \
            --set-as-default
    else
        # Create new policy
        aws iam create-policy \
            --policy-name "$POLICY_NAME" \
            --policy-document file://<(cat << 'EOF'
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "STSPermissions", 
            "Effect": "Allow",
            "Action": [
                "sts:AssumeRole",
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        },
        {
            "Sid": "StackSetServicePermissions",
            "Effect": "Allow",
            "Action": [
                "iam:CreateServiceLinkedRole",
                "iam:GetRole",
                "iam:PassRole"
            ],
            "Resource": [
                "arn:aws:iam::*:role/aws-service-role/stacksets.cloudformation.amazonaws.com/AWSServiceRoleForCloudFormationStackSets*",
                "arn:aws:iam::*:role/OrganizationAccountAccessRole*"
            ]
        }
    ]
}
EOF
            ) \
            --description "Additional permissions for AWS IAM Manager StackSet operations"
    fi
    
    log_success "Custom policy ready: $POLICY_NAME"
}

attach_custom_policy() {
    log_info "Attaching custom policy..."
    
    aws iam attach-user-policy \
        --user-name "$USER_NAME" \
        --policy-arn "arn:aws:iam::$ACCOUNT_ID:policy/$POLICY_NAME"
    
    log_success "Attached custom policy: $POLICY_NAME"
}

create_access_key() {
    log_info "Creating access keys..."
    
    # Check if user already has 2 access keys (AWS limit)
    EXISTING_KEYS=$(aws iam list-access-keys --user-name "$USER_NAME" --query 'AccessKeyMetadata[].AccessKeyId' --output text)
    KEY_COUNT=$(echo "$EXISTING_KEYS" | wc -w)
    
    if [ "$KEY_COUNT" -ge 2 ]; then
        log_warning "User already has the maximum number of access keys (2)"
        log_info "Existing access keys:"
        aws iam list-access-keys --user-name "$USER_NAME" --query 'AccessKeyMetadata[].[AccessKeyId,Status,CreateDate]' --output table
        
        read -p "Do you want to delete the oldest key and create a new one? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            OLDEST_KEY=$(aws iam list-access-keys --user-name "$USER_NAME" --query 'AccessKeyMetadata | sort_by(@, &CreateDate) | [0].AccessKeyId' --output text)
            log_info "Deleting oldest access key: $OLDEST_KEY"
            aws iam delete-access-key --user-name "$USER_NAME" --access-key-id "$OLDEST_KEY"
            log_success "Deleted access key: $OLDEST_KEY"
        else
            log_info "Skipping access key creation"
            return
        fi
    fi
    
    # Create new access key
    ACCESS_KEY_OUTPUT=$(aws iam create-access-key --user-name "$USER_NAME")
    ACCESS_KEY_ID=$(echo "$ACCESS_KEY_OUTPUT" | jq -r '.AccessKey.AccessKeyId')
    SECRET_ACCESS_KEY=$(echo "$ACCESS_KEY_OUTPUT" | jq -r '.AccessKey.SecretAccessKey')
    
    log_success "Created new access key!"
    
    # Display the credentials prominently
    echo
    echo "======================================================================="
    echo -e "${GREEN}                    AWS CREDENTIALS CREATED                     ${NC}"
    echo "======================================================================="
    echo
    echo -e "${YELLOW}AWS_ACCESS_KEY_ID=${NC}$ACCESS_KEY_ID"
    echo -e "${YELLOW}AWS_SECRET_ACCESS_KEY=${NC}$SECRET_ACCESS_KEY"
    echo -e "${YELLOW}AWS_REGION=${NC}us-east-1"
    echo
    echo "======================================================================="
    echo -e "${RED}IMPORTANT: Save these credentials securely. They will not be shown again!${NC}"
    echo "======================================================================="
    echo
    
    # Create .env file template
    cat > /tmp/iam-manager.env << EOF
# AWS IAM Manager Environment Variables
# Copy these values to your .env file

AWS_ACCESS_KEY_ID=$ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY=$SECRET_ACCESS_KEY
AWS_REGION=us-east-1
PORT=8080
EOF
    
    log_info "Environment file template created at: /tmp/iam-manager.env"
    log_info "Copy this file to your project directory as .env"
}

enable_stacksets_trusted_access() {
    log_info "Checking StackSets trusted access..."
    
    # Check if trusted access is already enabled
    TRUSTED_SERVICES=$(aws organizations list-aws-service-access-for-organization --query 'EnabledServicePrincipals[].ServicePrincipal' --output text 2>/dev/null || echo "")
    
    if echo "$TRUSTED_SERVICES" | grep -q "stacksets.cloudformation.amazonaws.com"; then
        log_success "StackSets trusted access already enabled"
    else
        log_info "Enabling StackSets trusted access..."
        if aws organizations enable-aws-service-access --service-principal stacksets.cloudformation.amazonaws.com 2>/dev/null; then
            log_success "Enabled StackSets trusted access"
        else
            log_warning "Could not enable StackSets trusted access automatically."
            log_info "You may need to enable it manually in the AWS Organizations console"
            log_info "or run: aws organizations enable-aws-service-access --service-principal stacksets.cloudformation.amazonaws.com"
        fi
    fi
}

create_stacksets_service_role() {
    log_info "Checking StackSets service-linked role..."
    
    if aws iam get-role --role-name AWSServiceRoleForCloudFormationStackSets &> /dev/null; then
        log_success "StackSets service-linked role already exists"
    else
        log_info "Creating StackSets service-linked role..."
        if aws iam create-service-linked-role --aws-service-name stacksets.cloudformation.amazonaws.com 2>/dev/null; then
            log_success "Created StackSets service-linked role"
        else
            log_warning "Could not create StackSets service-linked role automatically"
            log_info "You may need to create it manually or it will be created automatically on first use"
        fi
    fi
}

print_summary() {
    echo
    echo "======================================================================="
    echo -e "${GREEN}                 DEPLOYMENT COMPLETE                           ${NC}"
    echo "======================================================================="
    echo
    echo -e "${BLUE}User Details:${NC}"
    echo "  Name: $USER_NAME"
    echo "  ARN: arn:aws:iam::$ACCOUNT_ID:user/$USER_NAME"
    echo
    echo -e "${BLUE}Attached Policies:${NC}"
    echo "  - IAMFullAccess (managed)"
    echo "  - CloudFormationFullAccess (managed)" 
    echo "  - AWSOrganizationsReadOnlyAccess (managed)"
    echo "  - $POLICY_NAME (custom)"
    echo
    echo -e "${BLUE}Next Steps:${NC}"
    echo "  1. Copy the access credentials to your .env file"
    echo "  2. Run: make dev"
    echo "  3. Open: http://localhost:8080"
    echo "  4. Navigate to StackSet tab to deploy roles"
    echo
    echo -e "${YELLOW}Environment file template: /tmp/iam-manager.env${NC}"
    echo "======================================================================="
}

cleanup() {
    # Clean up temporary files if needed
    if [ -f "/tmp/policy.json" ]; then
        rm -f /tmp/policy.json
    fi
}

# Main execution
main() {
    echo "======================================================================="
    echo -e "${BLUE}          AWS IAM Manager - User Deployment Script             ${NC}"
    echo "======================================================================="
    echo
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Pre-flight checks
    check_aws_cli
    check_aws_credentials
    check_permissions
    
    echo
    log_info "This script will create an IAM user '$USER_NAME' with the following permissions:"
    echo "  - IAMFullAccess (for user management)"
    echo "  - CloudFormationFullAccess (for StackSet operations)"
    echo "  - AWSOrganizationsReadOnlyAccess (for account discovery)"
    echo "  - Custom policy for additional StackSet permissions"
    echo
    
    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Operation cancelled"
        exit 0
    fi
    
    echo
    # Main deployment steps
    create_user
    attach_managed_policies
    create_custom_policy
    attach_custom_policy
    create_access_key
    enable_stacksets_trusted_access
    create_stacksets_service_role
    
    # Summary
    print_summary
}

# Script execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi