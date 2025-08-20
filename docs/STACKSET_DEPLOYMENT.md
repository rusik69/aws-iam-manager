# AWS IAM Manager StackSet Deployment Guide

This guide explains how to deploy the AWS IAM Manager role across all accounts in your AWS Organization using CloudFormation StackSets.

> **⚡ Quick Start**: The AWS IAM Manager application now includes a **web-based StackSet deployment interface**! You can deploy roles to all accounts with one click from the application's StackSet tab. This guide covers both the web interface and traditional command-line methods.

## Overview

The StackSet deployment creates an IAM role in each account that allows the master account's IAM Manager application to perform user management operations across all organization accounts.

## Prerequisites

### 1. AWS Organizations Setup
- AWS Organizations must be enabled in your master account
- All target accounts must be part of the organization
- Trusted access for CloudFormation StackSets must be enabled

### 2. Required Permissions
The IAM user deploying the StackSet needs these permissions in the master account:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudformation:*",
                "organizations:ListAccounts",
                "organizations:DescribeOrganization",
                "iam:CreateRole",
                "iam:DeleteRole",
                "iam:AttachRolePolicy",
                "iam:DetachRolePolicy",
                "iam:PutRolePolicy",
                "iam:DeleteRolePolicy",
                "iam:PassRole"
            ],
            "Resource": "*"
        }
    ]
}
```

### 3. StackSets Service-Linked Role
Enable trusted access for CloudFormation StackSets:

```bash
aws organizations enable-aws-service-access --service-principal stacksets.cloudformation.amazonaws.com
```

Create the service-linked role:

```bash
aws iam create-service-linked-role --aws-service-name stacksets.cloudformation.amazonaws.com
```

## Deployment Options

### Option 1: Web Interface (Recommended) 🆕

The easiest way to deploy the StackSet is through the application's web interface:

1. **Start the application**:
   ```bash
   make dev
   ```

2. **Navigate to StackSet tab**:
   - Open `http://localhost:8080`
   - Click on "StackSet" in the navigation

3. **Validate permissions**:
   - Click "Validate" to check your AWS permissions
   - Ensure you have StackSet and Organizations access

4. **Deploy with one click**:
   - Click "Deploy StackSet" button
   - Monitor real-time deployment progress
   - View per-account deployment status

### Option 2: Automated Script

The automated script handles the entire deployment process:

```bash
cd /home/rusik/aws-iam-manager/scripts
./deploy-stackset.sh
```

#### Script Commands:
- `./deploy-stackset.sh` - Deploy or update the StackSet
- `./deploy-stackset.sh status` - Check current deployment status
- `./deploy-stackset.sh delete` - Delete the StackSet and all instances
- `./deploy-stackset.sh help` - Show help information

### Option 3: Manual AWS CLI Commands

#### Step 1: Create the StackSet

```bash
aws cloudformation create-stack-set \
    --stack-set-name "IAMManagerRoleStackSet" \
    --template-body file://cloudformation/iam-manager-role.yaml \
    --parameters \
        ParameterKey=MasterAccountId,ParameterValue=257801542589 \
        ParameterKey=RoleName,ParameterValue=OrganizationAccountAccessRole \
        ParameterKey=MasterUserName,ParameterValue=iam-manager \
    --capabilities CAPABILITY_NAMED_IAM \
    --description "IAM Manager Role for cross-account access" \
    --permission-model SELF_MANAGED
```

#### Step 2: Deploy to All Accounts

First, get all organization account IDs:

```bash
ACCOUNTS=$(aws organizations list-accounts --query 'Accounts[?Status==`ACTIVE`].Id' --output text | tr '\t' ',')
```

Then deploy to all accounts:

```bash
aws cloudformation create-stack-instances \
    --stack-set-name "IAMManagerRoleStackSet" \
    --accounts $ACCOUNTS \
    --regions us-east-1
```

#### Step 3: Monitor Deployment

```bash
# Get operation ID from the create-stack-instances output
OPERATION_ID="your-operation-id"

# Monitor progress
aws cloudformation describe-stack-set-operation \
    --stack-set-name "IAMManagerRoleStackSet" \
    --operation-id $OPERATION_ID
```

### Option 4: AWS Console

1. Open the CloudFormation console in your master account
2. Navigate to StackSets
3. Click "Create StackSet"
4. Upload the template file: `cloudformation/iam-manager-role.yaml`
5. Set parameters:
   - MasterAccountId: `257801542589`
   - RoleName: `OrganizationAccountAccessRole`
   - MasterUserName: `iam-manager`
6. Deploy to all organization accounts

## Configuration Parameters

| Parameter | Description | Default Value |
|-----------|-------------|---------------|
| MasterAccountId | AWS Account ID of the master account | 257801542589 |
| RoleName | Name of the IAM role to create | OrganizationAccountAccessRole |
| MasterUserName | IAM user name in master account | iam-manager |

## What Gets Deployed

The StackSet creates the following resources in each target account:

### 1. IAM Role: `OrganizationAccountAccessRole`
- **Purpose**: Allows cross-account access from the master account
- **Trust Policy**: Trusts the master account and specific IAM user
- **Permissions**: Full IAM access via managed policy and custom policy

### 2. IAM Policy: `IAMManagerCustomPermissions`
- **Attached to**: The IAM role
- **Permissions**: Detailed IAM user, group, and policy management permissions

## Security Features

### 1. External ID
- Each role uses the target account ID as an external ID
- Format: `{AccountId}-iam-manager`
- Prevents confused deputy attacks

### 2. Least Privilege Principle
- Role can only be assumed by specific IAM user in master account
- Permissions focused on IAM user management operations

### 3. Auditing
- All role assumptions are logged in CloudTrail
- Role usage can be monitored via AWS Config

## Verification

### 1. Check StackSet Status

```bash
aws cloudformation list-stack-instances \
    --stack-set-name "IAMManagerRoleStackSet" \
    --query 'Summaries[*].[Account,Region,Status]' \
    --output table
```

### 2. Test Role Assumption

```bash
# Test assuming role in a target account
aws sts assume-role \
    --role-arn "arn:aws:iam::TARGET-ACCOUNT-ID:role/OrganizationAccountAccessRole" \
    --role-session-name "IAMManagerTest" \
    --external-id "TARGET-ACCOUNT-ID-iam-manager"
```

### 3. Verify Application Access

1. Update the AWS IAM Manager application with valid credentials
2. Navigate to the application at `http://localhost:8080`
3. Click on any account to view users
4. Verify that you can see and manage IAM users

## Troubleshooting

### Common Issues

#### 1. StackSet Creation Fails
- **Cause**: Insufficient permissions
- **Solution**: Ensure the deploying user has required permissions
- **Check**: Verify trusted access is enabled for StackSets

#### 2. Stack Instance Deployment Fails
- **Cause**: Account is not part of organization or suspended
- **Solution**: Check account status in AWS Organizations
- **Check**: Verify account has sufficient quotas for IAM resources

#### 3. Role Assumption Fails
- **Cause**: External ID mismatch or trust policy issues
- **Solution**: Verify external ID format and trust policy configuration
- **Check**: Ensure master account ID and user name are correct

#### 4. Permission Denied Errors
- **Cause**: Role doesn't have sufficient permissions
- **Solution**: Review and update the custom policy if needed
- **Check**: Verify IAMFullAccess policy is attached

### Debug Commands

```bash
# Check StackSet details
aws cloudformation describe-stack-set --stack-set-name "IAMManagerRoleStackSet"

# Check specific stack instance
aws cloudformation describe-stack-instance \
    --stack-set-name "IAMManagerRoleStackSet" \
    --stack-instance-account TARGET-ACCOUNT-ID \
    --stack-instance-region us-east-1

# List all organization accounts
aws organizations list-accounts --query 'Accounts[*].[Id,Name,Status]' --output table

# Check if trusted access is enabled
aws organizations list-aws-service-access-for-organization
```

## Cleanup

To remove the StackSet and all deployed roles:

```bash
# Using the script
./deploy-stackset.sh delete

# Or manually
aws cloudformation delete-stack-instances \
    --stack-set-name "IAMManagerRoleStackSet" \
    --retain-stacks false \
    --regions us-east-1 \
    --deployment-targets OrganizationalUnitIds=r-*

# Wait for instances to be deleted, then delete the StackSet
aws cloudformation delete-stack-set --stack-set-name "IAMManagerRoleStackSet"
```

## Best Practices

1. **Test First**: Deploy to a few test accounts before deploying organization-wide
2. **Monitor**: Set up CloudWatch alarms for StackSet operation failures
3. **Review**: Regularly review role usage and permissions
4. **Update**: Keep the role permissions current with application requirements
5. **Backup**: Document your StackSet configuration for disaster recovery

## Support

For issues with the deployment:
1. Check the CloudFormation console for detailed error messages
2. Review CloudTrail logs for API call failures
3. Verify AWS Organizations configuration
4. Test with a single account first before deploying organization-wide