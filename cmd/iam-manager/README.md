# IAM Manager CLI

A Go-based command-line tool that replaces the bash script for managing AWS IAM resources and StackSets.

## Features

- **User Management**: Create and manage IAM users with proper permissions
- **Policy Management**: Attach managed and custom policies automatically  
- **StackSet Operations**: Deploy IAM roles across organization accounts
- **Cross-Account Access**: Set up roles for multi-account management
- **Status Checking**: View current deployment status and configurations

## Installation

### Prerequisites

- Go 1.21+ installed
- AWS CLI configured with appropriate credentials
- Access to AWS Organizations (for StackSet operations)

### Build from source

```bash
# Clone the repository
git clone https://github.com/rusik69/aws-iam-manager
cd aws-iam-manager

# Build the CLI tool
make build-cli

# Or run directly
go run ./cmd/iam-manager
```

### Install globally

```bash
# Install to $GOPATH/bin
make install

# Or directly with go install
go install ./cmd/iam-manager
```

## Usage

### Available Commands

```bash
# Show help
iam-manager --help

# Deploy IAM user and resources
iam-manager deploy

# Remove IAM user and resources  
iam-manager remove

# Create IAM role for cross-account access
iam-manager create-role

# Remove IAM role and resources
iam-manager remove-role

# Deploy StackSet to organization accounts
iam-manager stackset-deploy

# Check StackSet deployment status
iam-manager stackset-status

# Delete StackSet and all instances
iam-manager stackset-delete

# Show current deployment status
iam-manager status
```

### Development

```bash
# Run directly from source
make dev-cli

# Run specific commands
make cli-help
make cli-deploy
make cli-status
make cli-stackset-deploy

# Build and test
make build-cli
make test
make fmt
make lint
```

## Configuration

The CLI tool uses the same environment variables as the web application:

```bash
# AWS Configuration
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"

# Optional customization
export IAM_USER_NAME="iam-manager"
export IAM_POLICY_NAME="IAMManagerCustomPolicy"  
export IAM_ROLE_NAME="IAMManagerRole"
export STACK_SET_NAME="IAMManagerRoleStackSet"
export REGIONS="us-east-1"
```

## Migration from Bash Script

The Go CLI provides the same functionality as the bash script with these benefits:

### Advantages over Bash Script

✅ **Better Error Handling** - Structured error messages and recovery  
✅ **Type Safety** - Compile-time checking prevents runtime errors  
✅ **Native AWS SDK** - Direct API calls instead of AWS CLI subprocess calls  
✅ **Faster Execution** - No shell process overhead  
✅ **Cross-Platform** - Runs on Windows, Linux, macOS  
✅ **Better Logging** - Structured, colored output  
✅ **Maintainable Code** - Clean separation of concerns  

### Command Mapping

| Bash Script | Go CLI |
|-------------|---------|
| `./scripts/iam-manager.sh deploy` | `iam-manager deploy` |
| `./scripts/iam-manager.sh remove` | `iam-manager remove` |
| `./scripts/iam-manager.sh create-role` | `iam-manager create-role` |
| `./scripts/iam-manager.sh remove-role` | `iam-manager remove-role` |
| `./scripts/iam-manager.sh stackset-deploy` | `iam-manager stackset-deploy` |
| `./scripts/iam-manager.sh stackset-status` | `iam-manager stackset-status` |
| `./scripts/iam-manager.sh stackset-delete` | `iam-manager stackset-delete` |
| `./scripts/iam-manager.sh status` | `iam-manager status` |

## Architecture

The CLI is built with:

- **Cobra** - Command-line interface framework
- **AWS SDK for Go v2** - Native AWS API integration
- **Structured logging** - Clear, colorized output
- **Context handling** - Proper cancellation and timeouts
- **Error wrapping** - Detailed error context

## Implementation Status

- ✅ **User Management** - Fully implemented with policy attachment
- ✅ **StackSet Administration Roles** - Service-linked and administration roles
- ✅ **Organization Integration** - Trusted access and OU discovery
- ✅ **Status Checking** - Basic resource verification
- 🚧 **StackSet Deployment** - Structure ready, needs CloudFormation integration
- 🚧 **User Removal** - Placeholder implementation
- 🚧 **Role Management** - Placeholder implementation
- 🚧 **StackSet Operations** - Placeholder implementation

## Troubleshooting

### Build Issues

**Missing go.sum entry for module:**
```bash
# Solution: Run dependency download first
go mod tidy
go mod download
```

**Go not found:**
```bash
# Install Go 1.21+ from https://golang.org/doc/install
# Or use package manager:
# Ubuntu/Debian: sudo apt install golang-go
# CentOS/RHEL: sudo yum install golang
# macOS: brew install go
```

**Permission denied:**
```bash
# Make build script executable
chmod +x build-cli.sh
```

### Runtime Issues

**AWS credentials not found:**
```bash
# Set environment variables
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key" 
export AWS_REGION="us-east-1"

# Or configure AWS CLI
aws configure
```

**Organization access denied:**
```bash
# Ensure you're in the management account
# Verify Organizations permissions
aws organizations describe-organization
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.