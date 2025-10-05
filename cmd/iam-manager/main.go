package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/spf13/cobra"
)

const (
	version = "2.0.0"

	// Default configuration
	defaultUserName     = "iam-manager"
	defaultPolicyName   = "IAMManagerCustomPolicy"
	defaultRoleName     = "IAMManagerRole"
	defaultStackSetName = "IAMManagerRoleStackSet"
	defaultRegions      = "us-east-1"
	defaultOrgRoleName  = "IAMManagerCrossAccountRole"

	// Colors for output
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// IAMManager handles AWS IAM operations
type IAMManager struct {
	iamClient    *iam.Client
	cfnClient    *cloudformation.Client
	orgsClient   *organizations.Client
	stsClient    *sts.Client
	accountID    string
	userName     string
	policyName   string
	roleName     string
	stackSetName string
	regions      string
	orgRoleName  string
}

func main() {
	var rootCmd = &cobra.Command{
		Use:     "iam-manager",
		Short:   "AWS IAM Manager - Unified management tool for IAM operations",
		Version: version,
		Long: `AWS IAM Manager provides unified management functionality for AWS IAM operations.
This tool handles user creation, policy management, StackSet deployment, and cross-account access setup.`,
	}

	// Add AWS commands
	rootCmd.AddCommand(deployCmd())
	rootCmd.AddCommand(removeCmd())
	rootCmd.AddCommand(createRoleCmd())
	rootCmd.AddCommand(removeRoleCmd())
	rootCmd.AddCommand(stacksetDeployCmd())
	rootCmd.AddCommand(stacksetStatusCmd())
	rootCmd.AddCommand(stacksetDeleteCmd())
	rootCmd.AddCommand(statusCmd())

	// Add Azure commands
	azureCmd := &cobra.Command{
		Use:   "azure",
		Short: "Azure AD operations",
		Long:  `Manage Azure AD enterprise applications and service principals.`,
	}
	azureCmd.AddCommand(azureListAppsCmd())
	azureCmd.AddCommand(azureDeleteAppCmd())
	rootCmd.AddCommand(azureCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newIAMManager(ctx context.Context) (*IAMManager, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)

	// Get caller identity to determine account ID
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}

	return &IAMManager{
		iamClient:    iam.NewFromConfig(cfg),
		cfnClient:    cloudformation.NewFromConfig(cfg),
		orgsClient:   organizations.NewFromConfig(cfg),
		stsClient:    stsClient,
		accountID:    *identity.Account,
		userName:     getEnvOrDefault("IAM_USER_NAME", defaultUserName),
		policyName:   getEnvOrDefault("IAM_POLICY_NAME", defaultPolicyName),
		roleName:     getEnvOrDefault("IAM_ROLE_NAME", defaultRoleName),
		stackSetName: getEnvOrDefault("STACK_SET_NAME", defaultStackSetName),
		regions:      getEnvOrDefault("REGIONS", defaultRegions),
		orgRoleName:  getEnvOrDefault("IAM_ORG_ROLE_NAME", defaultOrgRoleName),
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Logging functions
func logInfo(message string) {
	fmt.Printf("%s[INFO]%s %s\n", colorBlue, colorReset, message)
}

func logSuccess(message string) {
	fmt.Printf("%s[SUCCESS]%s %s\n", colorGreen, colorReset, message)
}

func logWarning(message string) {
	fmt.Printf("%s[WARNING]%s %s\n", colorYellow, colorReset, message)
}

func logError(message string) {
	fmt.Printf("%s[ERROR]%s %s\n", colorRed, colorReset, message)
}

func printHeader(title string) {
	fmt.Println("=======================================================================")
	fmt.Printf("%s           %s              %s\n", colorBlue, title, colorReset)
	fmt.Println("=======================================================================")
	fmt.Println()
}

// Command implementations
func deployCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Deploy IAM user and resources",
		Long:  `Deploy the IAM Manager user with all required permissions and policies.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			manager, err := newIAMManager(ctx)
			if err != nil {
				logError(fmt.Sprintf("Failed to initialize: %v", err))
				os.Exit(1)
			}

			if err := manager.deployUser(ctx); err != nil {
				logError(fmt.Sprintf("Deploy failed: %v", err))
				os.Exit(1)
			}
		},
	}
}

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove IAM user and resources",
		Long:  `Remove the IAM Manager user and all associated resources.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			manager, err := newIAMManager(ctx)
			if err != nil {
				logError(fmt.Sprintf("Failed to initialize: %v", err))
				os.Exit(1)
			}

			if err := manager.removeUser(ctx); err != nil {
				logError(fmt.Sprintf("Remove failed: %v", err))
				os.Exit(1)
			}
		},
	}
}

func createRoleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-role",
		Short: "Create IAM role for cross-account access",
		Long:  `Create an IAM role that can be used for cross-account access instead of users.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			manager, err := newIAMManager(ctx)
			if err != nil {
				logError(fmt.Sprintf("Failed to initialize: %v", err))
				os.Exit(1)
			}

			if err := manager.createRole(ctx); err != nil {
				logError(fmt.Sprintf("Create role failed: %v", err))
				os.Exit(1)
			}
		},
	}
}

func removeRoleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-role",
		Short: "Remove IAM role and resources",
		Long:  `Remove the IAM role and all associated resources.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			manager, err := newIAMManager(ctx)
			if err != nil {
				logError(fmt.Sprintf("Failed to initialize: %v", err))
				os.Exit(1)
			}

			if err := manager.removeRole(ctx); err != nil {
				logError(fmt.Sprintf("Remove role failed: %v", err))
				os.Exit(1)
			}
		},
	}
}

func stacksetDeployCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stackset-deploy",
		Short: "Deploy StackSet for organization setup",
		Long:  `Deploy StackSet to all organization accounts for cross-account access.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			manager, err := newIAMManager(ctx)
			if err != nil {
				logError(fmt.Sprintf("Failed to initialize: %v", err))
				os.Exit(1)
			}

			if err := manager.deployStackSet(ctx); err != nil {
				logError(fmt.Sprintf("StackSet deploy failed: %v", err))
				os.Exit(1)
			}
		},
	}
}

func stacksetStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stackset-status",
		Short: "Show StackSet deployment status",
		Long:  `Check the current status of StackSet deployment across organization accounts.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			manager, err := newIAMManager(ctx)
			if err != nil {
				logError(fmt.Sprintf("Failed to initialize: %v", err))
				os.Exit(1)
			}

			if err := manager.stackSetStatus(ctx); err != nil {
				logError(fmt.Sprintf("StackSet status failed: %v", err))
				os.Exit(1)
			}
		},
	}
}

func stacksetDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stackset-delete",
		Short: "Delete StackSet and all instances",
		Long:  `Delete the StackSet and all instances from organization accounts.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			manager, err := newIAMManager(ctx)
			if err != nil {
				logError(fmt.Sprintf("Failed to initialize: %v", err))
				os.Exit(1)
			}

			if err := manager.deleteStackSet(ctx); err != nil {
				logError(fmt.Sprintf("StackSet delete failed: %v", err))
				os.Exit(1)
			}
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current deployment status",
		Long:  `Display the current status of all IAM Manager resources and configurations.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			manager, err := newIAMManager(ctx)
			if err != nil {
				logError(fmt.Sprintf("Failed to initialize: %v", err))
				os.Exit(1)
			}

			if err := manager.showStatus(ctx); err != nil {
				logError(fmt.Sprintf("Status check failed: %v", err))
				os.Exit(1)
			}
		},
	}
}