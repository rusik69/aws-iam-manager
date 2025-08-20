# AWS IAM Manager

A web application for managing IAM users across multiple AWS accounts in an organization. Features a built-in StackSet deployment system for automated role provisioning. Built with Go backend and Vue.js frontend, served through Go to eliminate Node.js dependency.

## Features

- **Multi-Account Management**: Switch between AWS accounts using cross-account roles
- **User Listing**: View all IAM users across organization accounts
- **Access Key Management**: Create, rotate, and delete access keys
- **Password Status**: Check if users have console passwords set
- **User Details**: View comprehensive user information including ARN, creation date, and access keys
- **🆕 StackSet Deployment**: Deploy IAM roles to all organization accounts with one click from the web interface
- **Permission Validation**: Automatically validate required AWS permissions before deployment
- **Real-time Monitoring**: Track StackSet deployment progress across all accounts

## Architecture

- **Backend**: Go with Gin framework and AWS SDK
- **Frontend**: Vue.js 3 with Vite, compiled and served by Go
- **Deployment**: Docker Compose for easy deployment
- **Authentication**: Uses AWS IAM roles for cross-account access

## Prerequisites

1. AWS Organization with multiple accounts
2. IAM user in master account with permissions to:
   - List organization accounts (`organizations:ListAccounts`)
   - Assume roles in target accounts
   - **CloudFormation StackSets operations** (required for web-based role deployment)
   - Organizations service access
3. `OrganizationAccountAccessRole` (or similar) in each target account **OR** use the built-in StackSet deployment feature
4. Docker and Docker Compose

> **New Feature:** The application now includes a web-based StackSet deployment interface! You can deploy the required IAM roles to all organization accounts directly from the browser instead of manually creating them. See the [StackSet Deployment](#stackset-deployment) section below.

## Quick Start

### Option 1: Automated User Setup (Recommended)

Use the automated script to create an IAM user with all necessary permissions:

```bash
# Clone and navigate to the project
git clone <repository-url>
cd aws-iam-manager

# Deploy IAM user with required permissions
./scripts/deploy-iam-user.sh

# Use the displayed credentials to create your .env file
cp .env.example .env
# Edit .env with the credentials from the script output
```

### Option 2: Manual Setup

1. Clone and navigate to the project:
   ```bash
   git clone <repository-url>
   cd aws-iam-manager
   ```

2. Copy and configure environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your AWS credentials
   ```

3. Build and run with Docker Compose:
   ```bash
   docker-compose up --build
   ```

4. Access the application at `http://localhost:8080`

## Environment Variables

- `AWS_ACCESS_KEY_ID`: Access key for master AWS account
- `AWS_SECRET_ACCESS_KEY`: Secret key for master AWS account  
- `AWS_REGION`: AWS region (default: us-east-1)
- `PORT`: Application port (default: 8080)

## IAM Permissions

### 🚀 Automated Setup (Recommended)

Use the deployment script to automatically create an IAM user with all required permissions:

```bash
./scripts/deploy-iam-user.sh
```

This script will:
- Create an IAM user named `iam-manager`
- Attach all required managed policies (IAMFullAccess, CloudFormationFullAccess, AWSOrganizationsReadOnlyAccess)
- Create a custom policy for additional StackSet permissions
- Generate access keys and display them securely
- Enable StackSets trusted access and service-linked role
- Provide an environment file template

### Manual Setup

If you prefer to set up permissions manually, the IAM user in the master account needs the following permissions. For the new StackSet deployment feature, additional permissions are required:

#### Option 1: Managed Policies (Recommended for StackSet Feature)
Attach these AWS managed policies to the user:
- `IAMFullAccess` - For IAM user management
- `CloudFormationFullAccess` - For StackSet operations
- `AWSOrganizationsReadOnlyAccess` - For organization account discovery

#### Option 2: Custom Policy (Minimum Permissions)
Create a custom policy with these minimum permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "OrganizationsAccess",
            "Effect": "Allow",
            "Action": [
                "organizations:ListAccounts",
                "organizations:DescribeOrganization"
            ],
            "Resource": "*"
        },
        {
            "Sid": "AssumeRolePermissions",
            "Effect": "Allow",
            "Action": [
                "sts:AssumeRole",
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        },
        {
            "Sid": "StackSetPermissions",
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
                "cloudformation:ListStackSetOperations"
            ],
            "Resource": "*"
        }
    ]
}
```

### Target Account Roles
Each target account should have a role (e.g., `OrganizationAccountAccessRole`) that can be assumed by the master account. You can either attach the `IAMFullAccess` managed policy or create a custom policy:

#### Option 1: Managed Policy (Recommended)
Attach the AWS managed policy `IAMFullAccess` to the role.

#### Option 2: Custom Policy
Create a custom policy with these minimum permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:ListUsers",
                "iam:GetUser",
                "iam:ListAccessKeys",
                "iam:CreateAccessKey",
                "iam:DeleteAccessKey",
                "iam:GetLoginProfile"
            ],
            "Resource": "*"
        }
    ]
}
```

Trust policy for the role:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::MASTER-ACCOUNT-ID:root"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
```

## StackSet Deployment

### 🆕 Web-Based Role Deployment

The application now includes a built-in StackSet deployment feature that eliminates the need to manually create IAM roles in each account. You can deploy the required `OrganizationAccountAccessRole` to all organization accounts directly from the web interface.

### How to Use

1. **Start the Application**:
   ```bash
   make dev  # or docker-compose up
   ```

2. **Navigate to StackSet Tab**: 
   - Open `http://localhost:8080`
   - Click on the "StackSet" tab in the navigation

3. **Validate Permissions**:
   - Click "Validate" to check if you have the required permissions
   - The system will automatically detect your AWS account and user

4. **Deploy StackSet**:
   - Click "Deploy StackSet" to create IAM roles in all organization accounts
   - Monitor real-time deployment progress
   - View per-account deployment status

### What Gets Deployed

The StackSet creates the following in each target account:

- **IAM Role**: `OrganizationAccountAccessRole`
- **Permissions**: Full IAM access for user management operations
- **Trust Policy**: Allows your master account user to assume the role
- **External ID**: Account-specific external ID for security (`{AccountId}-iam-manager`)

### Prerequisites for StackSet Deployment

1. **AWS Organizations** with trusted access enabled for CloudFormation StackSets:
   ```bash
   aws organizations enable-aws-service-access --service-principal stacksets.cloudformation.amazonaws.com
   ```

2. **StackSets Service Role** (if using service-managed permissions):
   ```bash
   aws iam create-service-linked-role --aws-service-name stacksets.cloudformation.amazonaws.com
   ```

3. **Master Account Permissions**: See the updated IAM permissions section above

### Benefits

- **One-Click Deployment**: No manual role creation needed
- **Organization-Wide**: Automatically deploys to all active accounts
- **Consistent Configuration**: Ensures identical role setup across accounts
- **Real-Time Monitoring**: Track deployment progress and status
- **Error Handling**: Clear feedback on failed deployments
- **Security**: Built-in external ID protection and least privilege permissions

### Alternative: Manual Setup

If you prefer not to use the StackSet feature, you can still manually create the `OrganizationAccountAccessRole` in each account using the policies shown in the Target Account Roles section above.

## API Endpoints

### IAM Management
- `GET /api/accounts` - List organization accounts
- `GET /api/accounts/:accountId/users` - List users in account
- `GET /api/accounts/:accountId/users/:username` - Get user details
- `POST /api/accounts/:accountId/users/:username/keys` - Create access key
- `DELETE /api/accounts/:accountId/users/:username/keys/:keyId` - Delete access key
- `PUT /api/accounts/:accountId/users/:username/keys/:keyId/rotate` - Rotate access key

### 🆕 StackSet Management
- `GET /api/stackset/status` - Get StackSet deployment status
- `POST /api/stackset/deploy` - Deploy/update StackSet to all accounts
- `GET /api/stackset/deployment/:operationId` - Get deployment operation status
- `DELETE /api/stackset/` - Delete StackSet and all instances
- `GET /api/stackset/validate` - Validate StackSet deployment permissions

## Security Considerations

- Access keys are only displayed once when created/rotated
- All operations require proper IAM permissions
- Uses temporary credentials when assuming roles
- Frontend served by Go eliminates Node.js attack surface

## Development

To run in development mode:

1. Backend:
   ```bash
   cd backend
   go mod tidy
   go run .
   ```

2. Frontend (in separate terminal):
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

The frontend will proxy API calls to the backend running on port 8080.