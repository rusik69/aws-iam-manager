# AWS IAM Manager Scripts

This directory previously contained bash scripts for AWS IAM Manager operations.

## 🔄 Migration Notice

**The bash scripts have been replaced with a modern Go CLI tool.**

### Use the Go CLI instead:

```bash
# Build the CLI (requires Go 1.21+)
./build-cli.sh

# Deploy IAM user and resources
./bin/iam-manager deploy

# Deploy StackSet to all organization accounts
./bin/iam-manager stackset-deploy

# Check deployment status
./bin/iam-manager status
```

### Benefits of the Go CLI:

✅ **Better Performance** - Native AWS SDK, no subprocess overhead  
✅ **Type Safety** - Compile-time error checking  
✅ **Cross-Platform** - Works on Linux, macOS, Windows  
✅ **Better Error Handling** - Structured error messages  
✅ **Maintainable** - Clean, organized codebase  
✅ **Modern Tooling** - Built with industry-standard Go tools

## 📋 Available Commands (Go CLI)

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `deploy` | Deploy IAM user and resources | `./bin/iam-manager deploy` |
| `remove` | Remove IAM user and resources | `./bin/iam-manager remove` |
| `status` | Show current deployment status | `./bin/iam-manager status` |

### Cross-Account Access

| Command | Description | Example |
|---------|-------------|---------|
| `create-role` | Create IAM role for cross-account access | `./bin/iam-manager create-role` |
| `remove-role` | Remove IAM role and resources | `./bin/iam-manager remove-role` |

### StackSet Management

| Command | Description | Example |
|---------|-------------|---------|
| `stackset-deploy` | Deploy StackSet to all organization accounts | `./bin/iam-manager stackset-deploy` |
| `stackset-status` | Check StackSet deployment status | `./bin/iam-manager stackset-status` |
| `stackset-delete` | Remove StackSet from all accounts | `./bin/iam-manager stackset-delete` |

## 🔧 Prerequisites

Before running any commands, ensure you have:

1. **AWS CLI** installed and configured
   ```bash
   aws configure
   ```

2. **jq** installed for JSON processing
   ```bash
   # Ubuntu/Debian
   sudo apt-get install jq
   
   # CentOS/RHEL
   sudo yum install jq
   
   # macOS
   brew install jq
   ```

3. **Appropriate IAM permissions** for the operations you want to perform

## 📖 Detailed Usage

### 1. Initial Setup

Deploy the IAM Manager user and resources:

```bash
./scripts/iam-manager.sh deploy
```

This will:
- Create the `iam-manager` IAM user
- Attach necessary managed policies
- Create and attach custom policies
- Generate access keys
- Set up StackSets integration
- Create environment file template

### 2. Fix Cross-Account Access

If you get role assumption errors like:
```
AccessDenied: User: arn:aws:iam::257801542589:user/iam-manager is not authorized to perform: sts:AssumeRole
```

Deploy StackSet to all organization accounts for automatic setup:

```bash
# Deploy to all accounts automatically
./scripts/iam-manager.sh stackset-deploy
```

### 3. Check Status

View the current deployment status:

```bash
./scripts/iam-manager.sh status
```

This shows:
- IAM user status and access keys
- IAM role status (if created)
- Custom policies status
- Organizations integration status

### 4. Deploy to Organization (Alternative)

For organization-wide deployment using StackSets:

```bash
# Deploy roles to all organization accounts automatically
./scripts/iam-manager.sh stackset-deploy

# Check deployment status
./scripts/iam-manager.sh stackset-status
```

### 5. Clean Up

Remove all IAM Manager resources:

```bash
# Remove IAM user and policies
./scripts/iam-manager.sh remove

# Remove StackSet deployment (if used)
./scripts/iam-manager.sh stackset-delete
```

## 🔒 Security Best Practices

1. **Credential Management**
   - Store AWS credentials securely
   - Use IAM roles when possible instead of long-term access keys
   - Regularly rotate access keys

2. **Least Privilege**
   - The script creates users/roles with minimal necessary permissions
   - Review and customize policies based on your needs

3. **Cross-Account Access**
   - Use the `fix-trust` command to enable access to specific accounts
   - Consider using the `create-role` approach for better security

## 🎯 Common Use Cases

### Use Case 1: New Organization Setup

```bash
# 1. Deploy in management account
./scripts/iam-manager.sh deploy

# 2. Deploy to all organization accounts
./scripts/iam-manager.sh stackset-deploy

# 3. Start using IAM Manager
make dev
```


### Use Case 2: Role-Based Access (More Secure)

```bash
# 1. Deploy user first
./scripts/iam-manager.sh deploy

# 2. Create role for cross-account access
./scripts/iam-manager.sh create-role

# 3. Update .env file with role ARN
echo "IAM_ROLE_ARN=arn:aws:iam::257801542589:role/IAMManagerRole" >> .env
```

### Use Case 3: Troubleshooting

```bash
# Check what's deployed
./scripts/iam-manager.sh status

# Remove everything and start fresh
./scripts/iam-manager.sh remove
./scripts/iam-manager.sh deploy
```

## 🐛 Troubleshooting

### Common Issues

1. **AWS CLI not configured**
   ```bash
   aws configure
   ```

2. **jq not installed**
   ```bash
   sudo apt-get install jq  # Ubuntu/Debian
   ```

3. **Permission denied errors**
   - Ensure you have IAM admin permissions
   - Check if you're in the correct AWS account

4. **Role assumption failures**
   - Use `fix-trust` command to fix trust policies
   - Verify account IDs are correct

### Getting Help

```bash
# Show detailed help
./scripts/iam-manager.sh --help

# Check current status
./scripts/iam-manager.sh status
```

## 📁 Script Files

| File | Purpose | Status |
|------|---------|--------|
| `iam-manager.sh` | **Unified management script with all functionality** | ✅ Active |
| `README.md` | **Comprehensive documentation** | ✅ Active |

> **Note**: All individual scripts have been consolidated into the single `iam-manager.sh` unified script.

## 🔄 Migration from Individual Scripts

All individual IAM management scripts have been consolidated into the unified script:

```bash
# All functionality now in one script
./scripts/iam-manager.sh deploy                    # Was: ./scripts/deploy-iam-user.sh
./scripts/iam-manager.sh remove                    # Was: ./scripts/deploy-iam-user.sh remove
./scripts/iam-manager.sh create-role               # Was: ./scripts/create-iam-role.sh
./scripts/iam-manager.sh remove-role               # New functionality!
./scripts/iam-manager.sh stackset-deploy           # Was: ./scripts/deploy-stackset.sh
./scripts/iam-manager.sh stackset-status           # Was: ./scripts/deploy-stackset.sh status
./scripts/iam-manager.sh stackset-delete           # Was: ./scripts/deploy-stackset.sh delete
./scripts/iam-manager.sh status                    # New functionality!
```

## 📊 Environment Variables

The script respects these environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `AWS_ACCESS_KEY_ID` | AWS access key | Required |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | Required |
| `AWS_REGION` | AWS region | `us-east-1` |
| `IAM_ROLE_NAME` | Role name to assume | `OrganizationAccountAccessRole` |

## 🎉 Success!

Once everything is set up:

1. Copy credentials from `/tmp/iam-manager.env` to your `.env` file
2. Run `make dev` to start the application
3. Open http://localhost:8080
4. Click "Manage Users" to access cross-account user management

Happy IAM managing! 🚀