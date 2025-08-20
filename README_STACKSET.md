# AWS IAM Manager - StackSet Deployment

## 🚀 Quick Start

### Step 1: Create IAM User (Automated)
Create an IAM user with all required permissions:
```bash
./scripts/deploy-iam-user.sh
```

### Step 2: Deploy Roles 
## 🆕 Web Interface Deployment (Recommended)

The easiest way to deploy IAM roles is through the built-in web interface:

1. **Start the application**: `make dev`
2. **Open browser**: `http://localhost:8080`
3. **Click "StackSet" tab** in navigation
4. **Validate permissions** and **deploy with one click**

## Command Line Deployment

Deploy IAM roles across all AWS Organization accounts to enable cross-account IAM user management.

### Prerequisites

1. **AWS Organizations** enabled with all target accounts
2. **IAM user** with StackSet deployment permissions:
   - **Automated**: Run `./scripts/deploy-iam-user.sh` to create user with all required permissions
   - **Manual**: See [REQUIRED_PERMISSIONS.md](docs/REQUIRED_PERMISSIONS.md) for detailed setup
3. **Trusted access** enabled for CloudFormation StackSets (automatically handled by the script)

### 1-Command Deployment

```bash
# Navigate to scripts directory
cd scripts

# Run automated deployment
./deploy-stackset.sh
```

### What This Creates

- **IAM Role**: `OrganizationAccountAccessRole` in each account
- **Permissions**: Full IAM access for user management
- **Security**: External ID protection and least privilege access
- **Trust**: Master account `iam-manager` user can assume roles

### Configuration

Update these variables in `scripts/deploy-stackset.sh`:

```bash
MASTER_ACCOUNT_ID="257801542589"    # Your master account ID
MASTER_USER_NAME="iam-manager"      # Your IAM user name
ROLE_NAME="OrganizationAccountAccessRole"  # Role name in target accounts
```

### Verification

1. **Check deployment status**:
   ```bash
   ./deploy-stackset.sh status
   ```

2. **Test application access**:
   - Start the application: `make dev`
   - Open: http://localhost:8080
   - Navigate through accounts and users

### Files Created

```
aws-iam-manager/
├── cloudformation/
│   └── iam-manager-role.yaml        # CloudFormation template
├── scripts/
│   └── deploy-stackset.sh           # Automated deployment script
├── docs/
│   ├── STACKSET_DEPLOYMENT.md      # Detailed deployment guide
│   └── REQUIRED_PERMISSIONS.md     # Permission requirements
└── README_STACKSET.md              # This file
```

### Troubleshooting

- **Permission denied**: Check [REQUIRED_PERMISSIONS.md](docs/REQUIRED_PERMISSIONS.md)
- **No accounts found**: Verify AWS Organizations access
- **Role assumption fails**: Check external ID format and trust policy

### Support

- **Detailed Guide**: [STACKSET_DEPLOYMENT.md](docs/STACKSET_DEPLOYMENT.md)
- **Permissions**: [REQUIRED_PERMISSIONS.md](docs/REQUIRED_PERMISSIONS.md)
- **Application Logs**: `make logs` to see detailed error messages

### Cleanup

```bash
./deploy-stackset.sh delete
```

## Security Notes

- Roles use external IDs to prevent confused deputy attacks
- Each role trusts only the specific master account user
- All operations are logged in CloudTrail for auditing
- Permissions follow least privilege principle for IAM user management