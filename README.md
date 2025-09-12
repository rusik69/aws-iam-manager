# AWS IAM Manager

A comprehensive web application for managing IAM users across multiple AWS accounts in an organization. Features automated StackSet deployment, cross-account role management, and a modern web interface. Built with Go backend and Vue.js frontend with Docker support.

## ✨ Features

### Core Functionality
- **🏢 Multi-Account Management**: Switch between AWS accounts using cross-account roles
- **👥 User Management**: View, create, and manage IAM users across organization accounts
- **🔑 Access Key Management**: Create, rotate, and delete access keys securely
- **🔒 Password Management**: Check and manage console password status
- **📊 User Details**: Comprehensive user information including ARN, creation date, and permissions

### Advanced Features
- **📦 StackSet Deployment**: Deploy IAM roles to all organization accounts with one click
- **✅ Permission Validation**: Automatically validate required AWS permissions before deployment
- **📈 Real-time Monitoring**: Track StackSet deployment progress across all accounts
- **🌙 Dark/Light Theme**: Modern responsive UI with theme switching
- **🛡️ Security**: External ID protection and least privilege permissions

## 🏗️ Architecture

- **Backend**: Go 1.21+ with Gin framework and AWS SDK v2
- **Frontend**: Vue.js 3 with Vite, compiled and served by Go (no Node.js runtime dependency)
- **Deployment**: Docker Compose for easy deployment
- **Authentication**: Uses AWS IAM roles for cross-account access
- **Security**: External ID validation and proper credential management

## 📋 Prerequisites

1. **AWS Organization** with multiple accounts
2. **IAM user** in master account with permissions to:
   - List organization accounts (`organizations:ListAccounts`)
   - Assume roles in target accounts (`sts:AssumeRole`)
   - CloudFormation StackSets operations (for automated role deployment)
   - Organizations service access
3. **Target account roles** (`IAMManagerCrossAccountRole`) **OR** use our automated StackSet deployment
4. **Docker and Docker Compose** (for containerized deployment)
5. **Go 1.21+** (for CLI and local development)

## 🚀 Quick Start

### Option 1: Automated Setup (Recommended)

Use our modern Go CLI with Makefile targets for the best experience:

```bash
# 1. Clone the repository
git clone https://github.com/rusik69/aws-iam-manager.git
cd aws-iam-manager

# 2. Check AWS configuration
make check-aws-config

# 3. Configure AWS credentials (if needed)
aws configure
# OR set environment variables:
# export AWS_ACCESS_KEY_ID=your_key
# export AWS_SECRET_ACCESS_KEY=your_secret
# export AWS_REGION=us-east-1

# 4. Deploy IAM user with required permissions
make deploy-user

# 5. Deploy roles to all organization accounts
make deploy-stackset

# 6. Check deployment status
make status-stackset

# 7. Start the web application
make dev
```

Access the application at `http://localhost:8080`

### Option 2: Manual Docker Setup

```bash
# 1. Clone and configure
git clone https://github.com/rusik69/aws-iam-manager.git
cd aws-iam-manager

# 2. Configure environment
cp .env.example .env
# Edit .env with your AWS credentials

# 3. Build and run
docker-compose up --build
```

## ⚙️ Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Required AWS credentials
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here
AWS_REGION=us-east-1

# Optional IAM configuration
IAM_ORG_ROLE_NAME=IAMManagerCrossAccountRole
IAM_USER_NAME=iam-manager
IAM_POLICY_NAME=IAMManagerCustomPolicy
STACK_SET_NAME=IAMManagerRoleStackSet

# Application settings
PORT=8080
```

## 🛠️ Makefile Targets

### Build & Development
- `make help` - Show all available targets
- `make build-frontend` - Build Vue.js frontend
- `make build-backend` - Build Go backend server
- `make build-cli` - Build Go CLI application
- `make dev` - Run Docker development environment

### AWS IAM Management
- `make deploy-user` - Deploy IAM user and resources
- `make remove-user` - Remove IAM user and resources
- `make create-role` - Create IAM role for cross-account access
- `make remove-role` - Remove IAM role and resources
- `make deploy-stackset` - Deploy StackSet for organization setup
- `make status-stackset` - Show StackSet deployment status
- `make delete-stackset` - Delete StackSet and all instances
- `make cli-status` - Show current deployment status

### Setup & Configuration
- `make check-aws-config` - Verify AWS credentials and configuration

### Testing & Quality
- `make test` - Run all tests
- `make lint` - Lint all code
- `make check` - Run all checks (fmt + lint + test)

## 🔐 IAM Permissions

### Master Account User

The CLI deployment script creates an IAM user with these managed policies:
- `IAMFullAccess` - For IAM user management
- `CloudFormationFullAccess` - For StackSet operations  
- `AWSOrganizationsReadOnlyAccess` - For organization account discovery

### Custom Policy (Minimum Permissions)

If you prefer minimal permissions, create a custom policy:

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

## 📦 StackSet Deployment

### Web Interface (Recommended)

1. **Start the application**: `make dev`
2. **Open browser**: `http://localhost:8080`
3. **Navigate to StackSet tab**
4. **Validate permissions** and **deploy with one click**

### What Gets Deployed

The StackSet creates in each target account:

- **IAM Role**: `IAMManagerCrossAccountRole`
- **Permissions**: Full IAM access for user management operations
- **Trust Policy**: Allows your master account user to assume the role
- **External ID**: Account-specific external ID for security (`{AccountId}-iam-manager`)
- **Security**: Built-in protection against confused deputy attacks

### CLI Commands

```bash
# Deploy StackSet to all organization accounts
make deploy-stackset

# Check deployment status
make status-stackset

# Get detailed status information
make cli-status

# Delete StackSet and all instances
make delete-stackset
```

## 🌐 API Endpoints

### IAM Management
- `GET /api/accounts` - List organization accounts
- `GET /api/accounts/:accountId/users` - List users in account
- `GET /api/accounts/:accountId/users/:username` - Get user details
- `POST /api/accounts/:accountId/users/:username/keys` - Create access key
- `DELETE /api/accounts/:accountId/users/:username/keys/:keyId` - Delete access key
- `PUT /api/accounts/:accountId/users/:username/keys/:keyId/rotate` - Rotate access key

### StackSet Management
- `GET /api/stackset/status` - Get StackSet deployment status
- `POST /api/stackset/deploy` - Deploy/update StackSet to all accounts
- `GET /api/stackset/deployment/:operationId` - Get deployment operation status
- `DELETE /api/stackset/` - Delete StackSet and all instances
- `GET /api/stackset/validate` - Validate StackSet deployment permissions

## 🏗️ Development

### Prerequisites
- Go 1.21+
- Node.js 18+ (for frontend development)
- AWS CLI configured
- Docker & Docker Compose

### Backend Development
```bash
make build-backend
cd backend && go run ./cmd/server
```

### Frontend Development
```bash
make dev-frontend
# Runs on http://localhost:5173 with API proxy to :8080
```

### Full Development Environment
```bash
make dev
# Builds everything and runs with Docker Compose
```

### Testing
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Lint code
make lint

# Run all checks
make check
```

## 🔒 Security Considerations

- **Access Keys**: Only displayed once when created/rotated
- **IAM Permissions**: All operations require proper AWS permissions
- **Temporary Credentials**: Uses temporary credentials when assuming roles
- **External ID**: Prevents confused deputy attacks
- **Audit Trail**: All operations logged in CloudTrail
- **Frontend Security**: Served by Go eliminates Node.js attack surface
- **HTTPS Ready**: Built-in support for TLS termination

## 🛠️ Troubleshooting

### Common Issues

**AWS Credentials Not Configured**
```bash
make check-aws-config
```

**Permission Denied**
```bash
# Check if your user has required permissions
aws sts get-caller-identity
aws organizations list-accounts
```

**StackSet Deployment Fails**
```bash
# Check StackSet status
make status-stackset

# View detailed logs
make logs
```

**Container Won't Start**
```bash
# Check container logs
docker-compose logs aws-iam-manager

# Rebuild containers
make clean
make rebuild
```

## 📚 Additional Documentation

- [CloudFormation Template](cloudformation/iam-manager-role.yaml) - StackSet deployment template
- [Contributing Guidelines](CONTRIBUTING.md) - Development and contribution guide
- [Environment Variables](.env.example) - Configuration options

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Run tests: `make check`
4. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/rusik69/aws-iam-manager/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rusik69/aws-iam-manager/discussions)
- **Documentation**: This README and inline code documentation

---

**Made with ❤️ for AWS Organizations** - Simplifying cross-account IAM management