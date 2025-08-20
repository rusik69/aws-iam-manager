# Required IAM Permissions for AWS IAM Manager StackSet Deployment

This document outlines the exact IAM permissions required for deploying and managing the AWS IAM Manager StackSet.

## Master Account Permissions

### 1. StackSet Deployment User Permissions

The IAM user deploying the StackSet requires these permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "CloudFormationStackSetManagement",
            "Effect": "Allow",
            "Action": [
                "cloudformation:CreateStackSet",
                "cloudformation:UpdateStackSet",
                "cloudformation:DeleteStackSet",
                "cloudformation:DescribeStackSet",
                "cloudformation:ListStackSets",
                "cloudformation:CreateStackInstances",
                "cloudformation:DeleteStackInstances",
                "cloudformation:DescribeStackInstance",
                "cloudformation:ListStackInstances",
                "cloudformation:DescribeStackSetOperation",
                "cloudformation:ListStackSetOperations",
                "cloudformation:StopStackSetOperation"
            ],
            "Resource": "*"
        },
        {
            "Sid": "OrganizationsReadAccess",
            "Effect": "Allow",
            "Action": [
                "organizations:ListAccounts",
                "organizations:DescribeOrganization",
                "organizations:DescribeAccount",
                "organizations:ListAWSServiceAccessForOrganization",
                "organizations:ListRoots",
                "organizations:ListOrganizationalUnitsForParent",
                "organizations:ListAccountsForParent"
            ],
            "Resource": "*"
        },
        {
            "Sid": "IAMRoleManagement",
            "Effect": "Allow",
            "Action": [
                "iam:CreateRole",
                "iam:DeleteRole",
                "iam:GetRole",
                "iam:UpdateRole",
                "iam:AttachRolePolicy",
                "iam:DetachRolePolicy",
                "iam:PutRolePolicy",
                "iam:DeleteRolePolicy",
                "iam:GetRolePolicy",
                "iam:ListRolePolicies",
                "iam:ListAttachedRolePolicies",
                "iam:PassRole",
                "iam:TagRole",
                "iam:UntagRole",
                "iam:ListRoleTags"
            ],
            "Resource": [
                "arn:aws:iam::*:role/OrganizationAccountAccessRole",
                "arn:aws:iam::*:role/AWSCloudFormationStackSetExecutionRole"
            ]
        },
        {
            "Sid": "CloudFormationTemplateAccess",
            "Effect": "Allow",
            "Action": [
                "cloudformation:ValidateTemplate",
                "cloudformation:GetTemplate"
            ],
            "Resource": "*"
        },
        {
            "Sid": "STSAssumeRole",
            "Effect": "Allow",
            "Action": [
                "sts:AssumeRole"
            ],
            "Resource": [
                "arn:aws:iam::*:role/AWSCloudFormationStackSetExecutionRole"
            ]
        }
    ]
}
```

### 2. Create Managed Policy

Save the above JSON as `stackset-deployment-policy.json` and create the policy:

```bash
aws iam create-policy \
    --policy-name "StackSetDeploymentPolicy" \
    --policy-document file://stackset-deployment-policy.json \
    --description "Permissions for deploying AWS IAM Manager StackSet"
```

### 3. Attach Policy to User

```bash
aws iam attach-user-policy \
    --user-name "iam-manager" \
    --policy-arn "arn:aws:iam::257801542589:policy/StackSetDeploymentPolicy"
```

## Organization-Level Setup

### 1. Enable Trusted Access for StackSets

```bash
aws organizations enable-aws-service-access \
    --service-principal stacksets.cloudformation.amazonaws.com
```

### 2. Create StackSets Service-Linked Role

```bash
aws iam create-service-linked-role \
    --aws-service-name stacksets.cloudformation.amazonaws.com
```

### 3. Verify Trusted Access

```bash
aws organizations list-aws-service-access-for-organization \
    --query 'EnabledServicePrincipals[?ServicePrincipal==`stacksets.cloudformation.amazonaws.com`]'
```

## Target Account Permissions

### Self-Managed StackSets (Current Approach)

For self-managed StackSets, each target account needs the `AWSCloudFormationStackSetExecutionRole`:

#### 1. Execution Role Policy

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:CreateRole",
                "iam:DeleteRole",
                "iam:GetRole",
                "iam:UpdateRole",
                "iam:AttachRolePolicy",
                "iam:DetachRolePolicy",
                "iam:PutRolePolicy",
                "iam:DeleteRolePolicy",
                "iam:GetRolePolicy",
                "iam:ListRolePolicies",
                "iam:ListAttachedRolePolicies",
                "iam:TagRole",
                "iam:UntagRole",
                "iam:CreatePolicy",
                "iam:DeletePolicy",
                "iam:GetPolicy",
                "iam:GetPolicyVersion",
                "iam:ListPolicyVersions"
            ],
            "Resource": "*"
        }
    ]
}
```

#### 2. Trust Policy for Execution Role

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::257801542589:root"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
```

#### 3. Create Execution Role in Each Account

This can be done via StackSet as well. Create `execution-role.yaml`:

```yaml
AWSTemplateFormatVersion: '2010-09-09'
Description: 'StackSet Execution Role for CloudFormation'

Parameters:
  MasterAccountId:
    Type: String
    Default: '257801542589'

Resources:
  ExecutionRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: AWSCloudFormationStackSetExecutionRole
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              AWS: !Sub 'arn:aws:iam::${MasterAccountId}:root'
            Action: 'sts:AssumeRole'
      ManagedPolicyArns:
        - 'arn:aws:iam::aws:policy/PowerUserAccess'
      Policies:
        - PolicyName: 'IAMManagement'
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action:
                  - 'iam:*'
                Resource: '*'
```

## Runtime Permissions (Application Usage)

### 1. IAM Manager User Permissions

The `iam-manager` user needs these permissions for daily operations:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "OrganizationsAccess",
            "Effect": "Allow",
            "Action": [
                "organizations:ListAccounts",
                "organizations:DescribeAccount"
            ],
            "Resource": "*"
        },
        {
            "Sid": "AssumeRolePermissions",
            "Effect": "Allow",
            "Action": [
                "sts:AssumeRole"
            ],
            "Resource": "arn:aws:iam::*:role/OrganizationAccountAccessRole",
            "Condition": {
                "StringEquals": {
                    "sts:ExternalId": "*-iam-manager"
                }
            }
        },
        {
            "Sid": "GetCallerIdentity",
            "Effect": "Allow",
            "Action": [
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        }
    ]
}
```

### 2. Created Role Permissions in Target Accounts

The `OrganizationAccountAccessRole` created by the StackSet has:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:ListUsers",
                "iam:GetUser",
                "iam:CreateUser",
                "iam:DeleteUser",
                "iam:UpdateUser",
                "iam:TagUser",
                "iam:UntagUser",
                "iam:ListUserTags",
                "iam:GetLoginProfile",
                "iam:CreateLoginProfile",
                "iam:UpdateLoginProfile",
                "iam:DeleteLoginProfile",
                "iam:ListAccessKeys",
                "iam:CreateAccessKey",
                "iam:DeleteAccessKey",
                "iam:UpdateAccessKey",
                "iam:GetAccessKeyLastUsed",
                "iam:ListUserPolicies",
                "iam:ListAttachedUserPolicies",
                "iam:AttachUserPolicy",
                "iam:DetachUserPolicy",
                "iam:PutUserPolicy",
                "iam:DeleteUserPolicy",
                "iam:ListGroupsForUser",
                "iam:AddUserToGroup",
                "iam:RemoveUserFromGroup",
                "iam:ListGroups",
                "iam:GetGroup",
                "iam:CreateGroup",
                "iam:DeleteGroup",
                "iam:UpdateGroup",
                "iam:ListPolicies",
                "iam:GetPolicy",
                "iam:GetPolicyVersion",
                "iam:ListPolicyVersions",
                "iam:GetAccountSummary",
                "iam:GetAccountPasswordPolicy"
            ],
            "Resource": "*"
        }
    ]
}
```

## Verification Commands

### 1. Check User Permissions

```bash
# Check current user's permissions
aws sts get-caller-identity

# Test StackSet permissions
aws cloudformation list-stack-sets

# Test Organizations permissions
aws organizations list-accounts
```

### 2. Verify StackSet Deployment Permissions

```bash
# Test StackSet creation (dry-run)
aws cloudformation validate-template \
    --template-body file://cloudformation/iam-manager-role.yaml

# Check if execution roles exist in target accounts
aws sts assume-role \
    --role-arn "arn:aws:iam::TARGET-ACCOUNT:role/AWSCloudFormationStackSetExecutionRole" \
    --role-session-name "StackSetTest"
```

### 3. Test Runtime Permissions

```bash
# Test assuming the created role
aws sts assume-role \
    --role-arn "arn:aws:iam::TARGET-ACCOUNT:role/OrganizationAccountAccessRole" \
    --role-session-name "IAMManagerTest" \
    --external-id "TARGET-ACCOUNT-iam-manager"
```

## Minimal Permission Sets

### For Testing (Reduced Scope)

If you want to test with a subset of accounts first:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudformation:CreateStackSet",
                "cloudformation:CreateStackInstances",
                "cloudformation:DescribeStackSet",
                "cloudformation:DescribeStackInstance",
                "cloudformation:DescribeStackSetOperation"
            ],
            "Resource": "*",
            "Condition": {
                "StringEquals": {
                    "cloudformation:StackSetName": "IAMManagerRoleStackSet"
                }
            }
        }
    ]
}
```

### For Production (Enhanced Security)

Add additional conditions and restrictions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudformation:*"
            ],
            "Resource": "*",
            "Condition": {
                "StringLike": {
                    "cloudformation:StackSetName": "IAMManager*"
                }
            }
        },
        {
            "Effect": "Allow",
            "Action": [
                "organizations:ListAccounts"
            ],
            "Resource": "*",
            "Condition": {
                "IpAddress": {
                    "aws:SourceIp": ["YOUR-IP-RANGE"]
                }
            }
        }
    ]
}
```

## Troubleshooting Permission Issues

### Common Permission Errors

1. **AccessDenied on CreateStackSet**
   - Missing `cloudformation:CreateStackSet`
   - Check if user has the StackSet management permissions

2. **AccessDenied on ListAccounts**
   - Missing `organizations:ListAccounts`
   - Verify Organizations service is accessible

3. **AccessDenied on CreateStackInstances**
   - Missing execution role in target accounts
   - Check if trusted access is enabled

4. **AccessDenied during role assumption**
   - Check external ID format
   - Verify trust policy in target role

### Debug Steps

1. **Check Current Permissions**:
   ```bash
   aws iam simulate-principal-policy \
       --policy-source-arn "arn:aws:iam::257801542589:user/iam-manager" \
       --action-names "cloudformation:CreateStackSet" \
       --resource-arns "*"
   ```

2. **Test Individual API Calls**:
   ```bash
   aws cloudformation list-stack-sets
   aws organizations list-accounts
   aws sts get-caller-identity
   ```

3. **Check CloudTrail for Detailed Errors**:
   ```bash
   aws logs filter-log-events \
       --log-group-name "CloudTrail/CloudFormationEvents" \
       --start-time 1640995200000 \
       --filter-pattern "ERROR"
   ```