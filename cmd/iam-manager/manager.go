package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	orgtypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Deploy user functionality
func (m *IAMManager) deployUser(ctx context.Context) error {
	printHeader("AWS IAM Manager - User Deployment")

	logInfo("This will create an IAM user with the following permissions:")
	fmt.Println("  - IAMFullAccess (managed)")
	fmt.Println("  - AWSCloudFormationFullAccess (managed)")
	fmt.Println("  - AWSOrganizationsReadOnlyAccess (managed)")
	fmt.Println("  - Custom policy for additional StackSet permissions")
	fmt.Println()

	if !m.confirmAction("Do you want to continue?") {
		logInfo("Operation cancelled")
		return nil
	}

	if err := m.createUser(ctx); err != nil {
		return err
	}

	if err := m.attachManagedPolicies(ctx); err != nil {
		return err
	}

	if err := m.createCustomPolicy(ctx); err != nil {
		return err
	}

	if err := m.attachCustomPolicy(ctx); err != nil {
		return err
	}

	if err := m.createAccessKey(ctx); err != nil {
		return err
	}

	if err := m.enableStackSetsAccess(ctx); err != nil {
		return err
	}

	if err := m.createStackSetRoles(ctx); err != nil {
		return err
	}

	m.printDeploymentSummary()
	return nil
}

func (m *IAMManager) createUser(ctx context.Context) error {
	logInfo(fmt.Sprintf("Creating IAM user: %s", m.userName))

	// Check if user exists
	_, err := m.iamClient.GetUser(ctx, &iam.GetUserInput{
		UserName: aws.String(m.userName),
	})

	if err == nil {
		logWarning(fmt.Sprintf("User %s already exists", m.userName))
		if !m.confirmAction("Do you want to continue and update the user's policies?") {
			return fmt.Errorf("operation cancelled")
		}
		return nil
	}

	// Create user
	_, err = m.iamClient.CreateUser(ctx, &iam.CreateUserInput{
		UserName: aws.String(m.userName),
		Path:     aws.String("/"),
		Tags: []iamtypes.Tag{
			{Key: aws.String("Purpose"), Value: aws.String("IAMManager")},
			{Key: aws.String("CreatedBy"), Value: aws.String("iam-manager-go")},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	logSuccess(fmt.Sprintf("Created user: %s", m.userName))
	return nil
}

func (m *IAMManager) attachManagedPolicies(ctx context.Context) error {
	logInfo("Attaching managed policies...")

	managedPolicies := []string{
		"arn:aws:iam::aws:policy/IAMFullAccess",
		"arn:aws:iam::aws:policy/AWSCloudFormationFullAccess",
		"arn:aws:iam::aws:policy/AWSOrganizationsReadOnlyAccess",
	}

	for _, policyArn := range managedPolicies {
		_, err := m.iamClient.AttachUserPolicy(ctx, &iam.AttachUserPolicyInput{
			UserName:  aws.String(m.userName),
			PolicyArn: aws.String(policyArn),
		})

		if err != nil {
			logWarning(fmt.Sprintf("Failed to attach policy %s: %v", policyArn, err))
		} else {
			logSuccess(fmt.Sprintf("Attached policy: %s", policyArn))
		}
	}

	return nil
}

func (m *IAMManager) createCustomPolicy(ctx context.Context) error {
	logInfo("Creating custom policy...")

	policyDocument := `{
		"Version": "2012-10-17",
		"Statement": [
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
					"cloudformation:ListStackSetOperations",
					"cloudformation:ListStackSetOperationResults"
				],
				"Resource": "*"
			},
			{
				"Sid": "OrganizationsServiceAccess",
				"Effect": "Allow",
				"Action": [
					"organizations:EnableAWSServiceAccess",
					"organizations:ListAWSServiceAccessForOrganization"
				],
				"Resource": "*"
			}
		]
	}`

	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", m.accountID, m.policyName)

	// Check if policy exists
	_, err := m.iamClient.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: aws.String(policyArn),
	})

	if err == nil {
		logWarning(fmt.Sprintf("Custom policy %s already exists", m.policyName))
		return nil
	}

	// Create policy
	_, err = m.iamClient.CreatePolicy(ctx, &iam.CreatePolicyInput{
		PolicyName:     aws.String(m.policyName),
		PolicyDocument: aws.String(policyDocument),
		Description:    aws.String("Custom policy for IAM Manager StackSet operations"),
		Tags: []iamtypes.Tag{
			{Key: aws.String("Purpose"), Value: aws.String("IAMManager")},
			{Key: aws.String("CreatedBy"), Value: aws.String("iam-manager-go")},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create custom policy: %w", err)
	}

	logSuccess(fmt.Sprintf("Created custom policy: %s", m.policyName))
	return nil
}

func (m *IAMManager) attachCustomPolicy(ctx context.Context) error {
	logInfo("Attaching custom policy...")

	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", m.accountID, m.policyName)

	_, err := m.iamClient.AttachUserPolicy(ctx, &iam.AttachUserPolicyInput{
		UserName:  aws.String(m.userName),
		PolicyArn: aws.String(policyArn),
	})

	if err != nil {
		return fmt.Errorf("failed to attach custom policy: %w", err)
	}

	logSuccess("Custom policy attached successfully")
	return nil
}

func (m *IAMManager) createAccessKey(ctx context.Context) error {
	logInfo("Creating access key...")

	result, err := m.iamClient.CreateAccessKey(ctx, &iam.CreateAccessKeyInput{
		UserName: aws.String(m.userName),
	})

	if err != nil {
		return fmt.Errorf("failed to create access key: %w", err)
	}

	accessKey := result.AccessKey
	logSuccess("Access key created successfully")

	// Write to temporary file
	envContent := fmt.Sprintf(`# AWS IAM Manager Environment Configuration
# Generated on %s

AWS_ACCESS_KEY_ID=%s
AWS_SECRET_ACCESS_KEY=%s
AWS_REGION=us-east-1
PORT=8080

# Optional: Set this if you want to use a specific cross-account role name
# IAM_ORG_ROLE_NAME=IAMManagerCrossAccountRole
`, time.Now().Format("2006-01-02 15:04:05"), *accessKey.AccessKeyId, *accessKey.SecretAccessKey)

	envFile := "/tmp/iam-manager.env"
	if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
		logWarning(fmt.Sprintf("Failed to write environment file: %v", err))
	} else {
		logSuccess(fmt.Sprintf("Environment file created: %s", envFile))
	}

	fmt.Println()
	fmt.Printf("%sAccess Key Details:%s\n", colorCyan, colorReset)
	fmt.Printf("Access Key ID: %s\n", *accessKey.AccessKeyId)
	fmt.Printf("Secret Access Key: %s\n", *accessKey.SecretAccessKey)
	fmt.Println()
	fmt.Printf("%sIMPORTANT: Copy these credentials to your .env file before they disappear!%s\n", colorYellow, colorReset)

	return nil
}

func (m *IAMManager) enableStackSetsAccess(ctx context.Context) error {
	logInfo("Enabling StackSets trusted access...")

	_, err := m.orgsClient.EnableAWSServiceAccess(ctx, &organizations.EnableAWSServiceAccessInput{
		ServicePrincipal: aws.String("stacksets.cloudformation.amazonaws.com"),
	})

	if err != nil {
		// Check if it's already enabled
		services, listErr := m.orgsClient.ListAWSServiceAccessForOrganization(ctx, &organizations.ListAWSServiceAccessForOrganizationInput{})
		if listErr == nil {
			for _, service := range services.EnabledServicePrincipals {
				if *service.ServicePrincipal == "stacksets.cloudformation.amazonaws.com" {
					logSuccess("StackSets trusted access already enabled")
					return nil
				}
			}
		}

		logWarning(fmt.Sprintf("Failed to enable StackSets trusted access: %v", err))
		return nil
	}

	logSuccess("StackSets trusted access enabled")
	return nil
}

func (m *IAMManager) createStackSetRoles(ctx context.Context) error {
	if err := m.createServiceLinkedRole(ctx); err != nil {
		return err
	}

	return m.createAdministrationRole(ctx)
}

func (m *IAMManager) createServiceLinkedRole(ctx context.Context) error {
	logInfo("Checking StackSets service-linked role...")

	// Check if role exists
	_, err := m.iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String("AWSServiceRoleForCloudFormationStackSets"),
	})

	if err == nil {
		logSuccess("StackSets service-linked role already exists")
		return nil
	}

	// Create service-linked role
	_, err = m.iamClient.CreateServiceLinkedRole(ctx, &iam.CreateServiceLinkedRoleInput{
		AWSServiceName: aws.String("stacksets.cloudformation.amazonaws.com"),
	})

	if err != nil {
		logWarning("Could not create StackSets service-linked role automatically")
		logInfo("It will be created automatically on first use")
		return nil
	}

	logSuccess("Created StackSets service-linked role")
	return nil
}

func (m *IAMManager) createAdministrationRole(ctx context.Context) error {
	logInfo("Checking StackSet administration role...")

	roleName := "AWSCloudFormationStackSetAdministrationRole"

	// Check if role exists
	_, err := m.iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})

	if err == nil {
		logSuccess("StackSet administration role already exists")
		return nil
	}

	logInfo(fmt.Sprintf("Creating StackSet administration role: %s", roleName))

	trustPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": "cloudformation.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`

	// Create role
	_, err = m.iamClient.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
		Description:              aws.String("Allows CloudFormation to perform StackSet operations"),
		Tags: []iamtypes.Tag{
			{Key: aws.String("Purpose"), Value: aws.String("StackSetAdministration")},
			{Key: aws.String("CreatedBy"), Value: aws.String("iam-manager-go")},
		},
	})

	if err != nil {
		logWarning(fmt.Sprintf("Failed to create StackSet administration role: %v", err))
		return nil
	}

	logSuccess("Created StackSet administration role")

	// Attach policy
	logInfo("Attaching StackSet administration policy...")
	_, err = m.iamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSCloudFormationStackSetAdministrationRole"),
	})

	if err != nil {
		logWarning("Failed to attach StackSet administration policy")
		return nil
	}

	logSuccess("Attached StackSet administration policy")
	return nil
}

func (m *IAMManager) printDeploymentSummary() {
	fmt.Println()
	fmt.Println("=======================================================================")
	fmt.Printf("%s                 DEPLOYMENT COMPLETE                           %s\n", colorGreen, colorReset)
	fmt.Println("=======================================================================")
	fmt.Println()
	fmt.Printf("%sUser Details:%s\n", colorBlue, colorReset)
	fmt.Printf("  Name: %s\n", m.userName)
	fmt.Printf("  ARN: arn:aws:iam::%s:user/%s\n", m.accountID, m.userName)
	fmt.Println()
	fmt.Printf("%sAttached Policies:%s\n", colorBlue, colorReset)
	fmt.Println("  - IAMFullAccess (managed)")
	fmt.Println("  - AWSCloudFormationFullAccess (managed)")
	fmt.Println("  - AWSOrganizationsReadOnlyAccess (managed)")
	fmt.Printf("  - %s (custom)\n", m.policyName)
	fmt.Println()
	fmt.Printf("%sStackSet Configuration:%s\n", colorBlue, colorReset)
	fmt.Println("  - Organizations trusted access: Enabled")
	fmt.Println("  - Service-linked role: AWSServiceRoleForCloudFormationStackSets")
	fmt.Println("  - Administration role: AWSCloudFormationStackSetAdministrationRole")
	fmt.Println()
	fmt.Printf("%sNext Steps:%s\n", colorBlue, colorReset)
	fmt.Println("  1. Copy the access credentials to your .env file")
	fmt.Println("  2. Run: make dev")
	fmt.Println("  3. Open: http://localhost:8080")
	fmt.Println("  4. Navigate to StackSet tab to deploy roles")
	fmt.Println()
	fmt.Printf("%sEnvironment file template: /tmp/iam-manager.env%s\n", colorYellow, colorReset)
	fmt.Println("=======================================================================")
}

// Helper function to confirm user actions
func (m *IAMManager) confirmAction(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// Remove user functionality
func (m *IAMManager) removeUser(ctx context.Context) error {
	printHeader("AWS IAM Manager - User Removal")

	// Check if user exists
	_, err := m.iamClient.GetUser(ctx, &iam.GetUserInput{
		UserName: aws.String(m.userName),
	})

	if err != nil {
		logWarning(fmt.Sprintf("User %s does not exist", m.userName))
		return nil
	}

	logInfo(fmt.Sprintf("Found user: %s", m.userName))
	fmt.Println()

	logWarning(fmt.Sprintf("This will permanently delete the IAM user '%s' and all associated:", m.userName))
	fmt.Println("  - Access keys")
	fmt.Println("  - Attached policies")
	fmt.Printf("  - Custom policy: %s\n", m.policyName)
	fmt.Println()
	fmt.Printf("%sNote: StackSet administration role (AWSCloudFormationStackSetAdministrationRole)%s\n", colorYellow, colorReset)
	fmt.Printf("%swill be preserved as it may be used by other services.%s\n", colorYellow, colorReset)
	fmt.Println()

	if !m.confirmAction("Are you sure you want to continue?") {
		logInfo("Operation cancelled")
		return nil
	}

	// Remove access keys
	if err := m.removeAccessKeys(ctx); err != nil {
		return err
	}

	// Detach managed policies
	if err := m.detachManagedPolicies(ctx); err != nil {
		return err
	}

	// Remove custom policy
	if err := m.removeCustomPolicy(ctx); err != nil {
		return err
	}

	// Delete user
	if err := m.deleteUser(ctx); err != nil {
		return err
	}

	m.printRemovalSummary()
	return nil
}

func (m *IAMManager) removeAccessKeys(ctx context.Context) error {
	logInfo("Removing access keys...")

	// List access keys
	result, err := m.iamClient.ListAccessKeys(ctx, &iam.ListAccessKeysInput{
		UserName: aws.String(m.userName),
	})

	if err != nil {
		return fmt.Errorf("failed to list access keys: %w", err)
	}

	if len(result.AccessKeyMetadata) == 0 {
		logInfo("No access keys found")
		return nil
	}

	// Delete each access key
	for _, keyMetadata := range result.AccessKeyMetadata {
		logInfo(fmt.Sprintf("Deleting access key: %s", *keyMetadata.AccessKeyId))

		_, err := m.iamClient.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
			UserName:    aws.String(m.userName),
			AccessKeyId: keyMetadata.AccessKeyId,
		})

		if err != nil {
			logWarning(fmt.Sprintf("Failed to delete access key %s: %v", *keyMetadata.AccessKeyId, err))
		} else {
			logSuccess(fmt.Sprintf("Deleted access key: %s", *keyMetadata.AccessKeyId))
		}
	}

	return nil
}

func (m *IAMManager) detachManagedPolicies(ctx context.Context) error {
	logInfo("Detaching managed policies...")

	managedPolicies := []string{
		"arn:aws:iam::aws:policy/IAMFullAccess",
		"arn:aws:iam::aws:policy/AWSCloudFormationFullAccess",
		"arn:aws:iam::aws:policy/AWSOrganizationsReadOnlyAccess",
	}

	for _, policyArn := range managedPolicies {
		_, err := m.iamClient.DetachUserPolicy(ctx, &iam.DetachUserPolicyInput{
			UserName:  aws.String(m.userName),
			PolicyArn: aws.String(policyArn),
		})

		if err != nil {
			logWarning(fmt.Sprintf("Failed to detach policy %s: %v", policyArn, err))
		} else {
			logSuccess(fmt.Sprintf("Detached policy: %s", policyArn))
		}
	}

	return nil
}

func (m *IAMManager) removeCustomPolicy(ctx context.Context) error {
	logInfo("Removing custom policy...")

	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", m.accountID, m.policyName)

	// Detach policy from user first
	_, err := m.iamClient.DetachUserPolicy(ctx, &iam.DetachUserPolicyInput{
		UserName:  aws.String(m.userName),
		PolicyArn: aws.String(policyArn),
	})

	if err != nil {
		logWarning(fmt.Sprintf("Failed to detach custom policy: %v", err))
	} else {
		logSuccess("Detached custom policy from user")
	}

	// Delete the policy
	_, err = m.iamClient.DeletePolicy(ctx, &iam.DeletePolicyInput{
		PolicyArn: aws.String(policyArn),
	})

	if err != nil {
		logWarning(fmt.Sprintf("Failed to delete custom policy: %v", err))
	} else {
		logSuccess("Deleted custom policy")
	}

	return nil
}

func (m *IAMManager) deleteUser(ctx context.Context) error {
	logInfo(fmt.Sprintf("Deleting user: %s", m.userName))

	_, err := m.iamClient.DeleteUser(ctx, &iam.DeleteUserInput{
		UserName: aws.String(m.userName),
	})

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logSuccess(fmt.Sprintf("Deleted user: %s", m.userName))
	return nil
}

func (m *IAMManager) printRemovalSummary() {
	fmt.Println()
	fmt.Println("=======================================================================")
	fmt.Printf("%s                 USER REMOVAL COMPLETE                        %s\n", colorGreen, colorReset)
	fmt.Println("=======================================================================")
	fmt.Println()
	fmt.Printf("%sRemoved Resources:%s\n", colorBlue, colorReset)
	fmt.Printf("  • IAM User: %s\n", m.userName)
	fmt.Println("  • All associated access keys")
	fmt.Println("  • All attached managed policies")
	fmt.Printf("  • Custom policy: %s\n", m.policyName)
	fmt.Println()
	fmt.Printf("%sNote: StackSet administration roles were preserved%s\n", colorYellow, colorReset)
	fmt.Println("=======================================================================")
}

func (m *IAMManager) createRole(ctx context.Context) error {
	printHeader("AWS IAM Manager - Create Role")

	logInfo("This will create an IAM Manager role for cross-account access:")
	fmt.Printf("  - Role Name: %s\n", m.roleName)
	fmt.Printf("  - Account: %s\n", m.accountID)
	fmt.Println("  - Full IAM permissions for user management")
	fmt.Println("  - Trust relationship allowing master account access")
	fmt.Println()

	if !m.confirmAction("Do you want to continue?") {
		logInfo("Operation cancelled")
		return nil
	}

	// Check if role already exists
	_, err := m.iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(m.roleName),
	})

	if err == nil {
		logWarning(fmt.Sprintf("Role %s already exists", m.roleName))
		if !m.confirmAction("Do you want to continue and update the role's policies?") {
			return fmt.Errorf("operation cancelled")
		}
	} else {
		// Create the role
		if err := m.createIAMRole(ctx); err != nil {
			return err
		}
	}

	// Attach policies to role
	if err := m.attachPoliciesToRole(ctx); err != nil {
		return err
	}

	m.printRoleCreationSummary()
	return nil
}

func (m *IAMManager) createIAMRole(ctx context.Context) error {
	logInfo(fmt.Sprintf("Creating IAM role: %s", m.roleName))

	trustPolicy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": "arn:aws:iam::%s:root"
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`, m.accountID)

	_, err := m.iamClient.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(m.roleName),
		AssumeRolePolicyDocument: aws.String(trustPolicy),
		Description:              aws.String("IAM Manager role for cross-account access"),
		Path:                     aws.String("/"),
		Tags: []iamtypes.Tag{
			{Key: aws.String("Purpose"), Value: aws.String("IAMManager")},
			{Key: aws.String("CreatedBy"), Value: aws.String("iam-manager-go")},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	logSuccess(fmt.Sprintf("Created role: %s", m.roleName))
	return nil
}

func (m *IAMManager) attachPoliciesToRole(ctx context.Context) error {
	logInfo("Attaching policies to role...")

	managedPolicies := []string{
		"arn:aws:iam::aws:policy/IAMFullAccess",
		"arn:aws:iam::aws:policy/AWSCloudFormationFullAccess",
		"arn:aws:iam::aws:policy/AWSOrganizationsReadOnlyAccess",
	}

	for _, policyArn := range managedPolicies {
		_, err := m.iamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			RoleName:  aws.String(m.roleName),
			PolicyArn: aws.String(policyArn),
		})

		if err != nil {
			logWarning(fmt.Sprintf("Failed to attach policy %s: %v", policyArn, err))
		} else {
			logSuccess(fmt.Sprintf("Attached policy: %s", policyArn))
		}
	}

	// Also attach the custom policy if it exists
	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", m.accountID, m.policyName)
	_, err := m.iamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(m.roleName),
		PolicyArn: aws.String(policyArn),
	})

	if err != nil {
		logWarning(fmt.Sprintf("Failed to attach custom policy: %v", err))
		logInfo("Make sure to run 'deploy' command first to create the custom policy")
	} else {
		logSuccess("Attached custom policy")
	}

	return nil
}

func (m *IAMManager) printRoleCreationSummary() {
	fmt.Println()
	fmt.Println("=======================================================================")
	fmt.Printf("%s                 ROLE CREATION COMPLETE                       %s\n", colorGreen, colorReset)
	fmt.Println("=======================================================================")
	fmt.Println()
	fmt.Printf("%sRole Details:%s\n", colorBlue, colorReset)
	fmt.Printf("  Name: %s\n", m.roleName)
	fmt.Printf("  ARN: arn:aws:iam::%s:role/%s\n", m.accountID, m.roleName)
	fmt.Println()
	fmt.Printf("%sAttached Policies:%s\n", colorBlue, colorReset)
	fmt.Println("  - IAMFullAccess (managed)")
	fmt.Println("  - AWSCloudFormationFullAccess (managed)")
	fmt.Println("  - AWSOrganizationsReadOnlyAccess (managed)")
	fmt.Printf("  - %s (custom)\n", m.policyName)
	fmt.Println()
	fmt.Printf("%sUsage:%s\n", colorBlue, colorReset)
	fmt.Println("  Set IAM_ROLE_ARN in your .env file:")
	fmt.Printf("  IAM_ROLE_ARN=arn:aws:iam::%s:role/%s\n", m.accountID, m.roleName)
	fmt.Println("=======================================================================")
}

func (m *IAMManager) removeRole(ctx context.Context) error {
	printHeader("AWS IAM Manager - Remove Role")

	// Check if role exists
	_, err := m.iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(m.roleName),
	})

	if err != nil {
		logWarning(fmt.Sprintf("Role %s does not exist", m.roleName))
		return nil
	}

	logInfo(fmt.Sprintf("Found role: %s", m.roleName))
	fmt.Println()

	logWarning(fmt.Sprintf("This will permanently delete the IAM role '%s' and:", m.roleName))
	fmt.Println("  - Detach all attached policies")
	fmt.Println("  - Remove the role completely")
	fmt.Println()
	fmt.Printf("%sNote: This will NOT affect the IAM user or custom policies%s\n", colorYellow, colorReset)
	fmt.Println()

	if !m.confirmAction("Are you sure you want to continue?") {
		logInfo("Operation cancelled")
		return nil
	}

	// Detach policies from role
	if err := m.detachPoliciesFromRole(ctx); err != nil {
		return err
	}

	// Delete the role
	if err := m.deleteRole(ctx); err != nil {
		return err
	}

	m.printRoleRemovalSummary()
	return nil
}

func (m *IAMManager) detachPoliciesFromRole(ctx context.Context) error {
	logInfo("Detaching policies from role...")

	// List attached policies
	result, err := m.iamClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(m.roleName),
	})

	if err != nil {
		return fmt.Errorf("failed to list attached policies: %w", err)
	}

	// Detach each policy
	for _, policy := range result.AttachedPolicies {
		logInfo(fmt.Sprintf("Detaching policy: %s", *policy.PolicyArn))

		_, err := m.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
			RoleName:  aws.String(m.roleName),
			PolicyArn: policy.PolicyArn,
		})

		if err != nil {
			logWarning(fmt.Sprintf("Failed to detach policy %s: %v", *policy.PolicyArn, err))
		} else {
			logSuccess(fmt.Sprintf("Detached policy: %s", *policy.PolicyArn))
		}
	}

	return nil
}

func (m *IAMManager) deleteRole(ctx context.Context) error {
	logInfo(fmt.Sprintf("Deleting role: %s", m.roleName))

	_, err := m.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(m.roleName),
	})

	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	logSuccess(fmt.Sprintf("Deleted role: %s", m.roleName))
	return nil
}

func (m *IAMManager) printRoleRemovalSummary() {
	fmt.Println()
	fmt.Println("=======================================================================")
	fmt.Printf("%s                 ROLE REMOVAL COMPLETE                        %s\n", colorGreen, colorReset)
	fmt.Println("=======================================================================")
	fmt.Println()
	fmt.Printf("%sRemoved Resources:%s\n", colorBlue, colorReset)
	fmt.Printf("  • IAM Role: %s\n", m.roleName)
	fmt.Println("  • All attached policies (detached)")
	fmt.Println()
	fmt.Printf("%sNote: IAM user and custom policies were preserved%s\n", colorYellow, colorReset)
	fmt.Println("=======================================================================")
}

func (m *IAMManager) deployStackSet(ctx context.Context) error {
	printHeader("Deploy StackSet for Organization Setup")

	logInfo("StackSet Configuration:")
	fmt.Printf("  StackSet Name: %s\n", m.stackSetName)
	fmt.Printf("  Template: cloudformation/iam-manager-role.yaml\n")
	fmt.Printf("  Master Account: %s\n", m.accountID)
	fmt.Printf("  Master User: %s\n", m.userName)
	fmt.Printf("  Cross-Account Role Name: %s\n", m.orgRoleName)
	fmt.Printf("  Regions: %s\n", m.regions)
	fmt.Println()

	// Check if template exists
	if _, err := os.Stat("cloudformation/iam-manager-role.yaml"); os.IsNotExist(err) {
		logError("CloudFormation template not found: cloudformation/iam-manager-role.yaml")
		logInfo("Please ensure the CloudFormation template is available.")
		return fmt.Errorf("template file not found")
	}

	// Verify organization access
	if err := m.verifyOrganizationAccess(ctx); err != nil {
		return err
	}

	// Get root OU
	rootOU, err := m.getRootOU(ctx)
	if err != nil {
		return err
	}

	// Deploy StackSet
	return m.deployStackSetToOrganization(ctx, rootOU)
}

func (m *IAMManager) verifyOrganizationAccess(ctx context.Context) error {
	logInfo("Verifying organization access...")

	org, err := m.orgsClient.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		return fmt.Errorf("unable to access organization: %w", err)
	}

	logSuccess(fmt.Sprintf("Organization ID: %s", *org.Organization.Id))
	return nil
}

func (m *IAMManager) getRootOU(ctx context.Context) (string, error) {
	roots, err := m.orgsClient.ListRoots(ctx, &organizations.ListRootsInput{})
	if err != nil {
		return "", fmt.Errorf("failed to list roots: %w", err)
	}

	if len(roots.Roots) == 0 {
		return "", fmt.Errorf("no root organizational unit found")
	}

	rootOU := *roots.Roots[0].Id
	logSuccess(fmt.Sprintf("Root OU ID: %s", rootOU))
	return rootOU, nil
}

func (m *IAMManager) deployStackSetToOrganization(ctx context.Context, rootOU string) error {
	logInfo("Deploying StackSet to organization...")

	// Read the CloudFormation template
	templateBody, err := os.ReadFile("cloudformation/iam-manager-role.yaml")
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Check if StackSet already exists
	_, err = m.cfnClient.DescribeStackSet(ctx, &cloudformation.DescribeStackSetInput{
		StackSetName: aws.String(m.stackSetName),
	})

	if err != nil {
		// StackSet doesn't exist, create it
		logInfo("Creating new StackSet...")

		_, err := m.cfnClient.CreateStackSet(ctx, &cloudformation.CreateStackSetInput{
			StackSetName: aws.String(m.stackSetName),
			TemplateBody: aws.String(string(templateBody)),
			Description:  aws.String("IAM Manager role deployment for organization accounts"),
			Parameters: []cfntypes.Parameter{
				{
					ParameterKey:   aws.String("MasterAccountId"),
					ParameterValue: aws.String(m.accountID),
				},
				{
					ParameterKey:   aws.String("RoleName"),
					ParameterValue: aws.String(m.orgRoleName),
				},
				{
					ParameterKey:   aws.String("MasterUserName"),
					ParameterValue: aws.String(m.userName),
				},
			},
			Capabilities: []cfntypes.Capability{
				cfntypes.CapabilityCapabilityIam,
				cfntypes.CapabilityCapabilityNamedIam,
			},
			PermissionModel: cfntypes.PermissionModelsServiceManaged,
			AutoDeployment: &cfntypes.AutoDeployment{
				Enabled:                      aws.Bool(true),
				RetainStacksOnAccountRemoval: aws.Bool(false),
			},
			Tags: []cfntypes.Tag{
				{Key: aws.String("Purpose"), Value: aws.String("IAMManager")},
				{Key: aws.String("CreatedBy"), Value: aws.String("iam-manager-go")},
			},
		})

		if err != nil {
			return fmt.Errorf("failed to create StackSet: %w", err)
		}

		logSuccess("StackSet created successfully")

		// Wait a moment for StackSet creation to complete
		time.Sleep(2 * time.Second)
	} else {
		// StackSet exists, update it
		logInfo("Updating existing StackSet...")

		_, err := m.cfnClient.UpdateStackSet(ctx, &cloudformation.UpdateStackSetInput{
			StackSetName: aws.String(m.stackSetName),
			TemplateBody: aws.String(string(templateBody)),
			Description:  aws.String("IAM Manager role deployment for organization accounts"),
			Parameters: []cfntypes.Parameter{
				{
					ParameterKey:   aws.String("MasterAccountId"),
					ParameterValue: aws.String(m.accountID),
				},
				{
					ParameterKey:   aws.String("RoleName"),
					ParameterValue: aws.String(m.orgRoleName),
				},
				{
					ParameterKey:   aws.String("MasterUserName"),
					ParameterValue: aws.String(m.userName),
				},
			},
			Capabilities: []cfntypes.Capability{
				cfntypes.CapabilityCapabilityIam,
				cfntypes.CapabilityCapabilityNamedIam,
			},
			PermissionModel: cfntypes.PermissionModelsServiceManaged,
			AutoDeployment: &cfntypes.AutoDeployment{
				Enabled:                      aws.Bool(true),
				RetainStacksOnAccountRemoval: aws.Bool(false),
			},
		})

		if err != nil {
			return fmt.Errorf("failed to update StackSet: %w", err)
		}

		logSuccess("StackSet updated successfully")
	}

	// Deploy to organizational unit
	logInfo(fmt.Sprintf("Deploying StackSet instances to organizational unit: %s", rootOU))
	logInfo("Deployment Configuration:")
	logInfo("  • Parallel deployment to ALL accounts simultaneously (100% concurrency)")
	logInfo("  • High failure tolerance: up to 50% of accounts can fail")
	logInfo("  • Strict failure tolerance mode for better resilience")

	deployResult, err := m.cfnClient.CreateStackInstances(ctx, &cloudformation.CreateStackInstancesInput{
		StackSetName: aws.String(m.stackSetName),
		DeploymentTargets: &cfntypes.DeploymentTargets{
			OrganizationalUnitIds: []string{rootOU},
		},
		Regions: []string{m.regions},
		OperationPreferences: &cfntypes.StackSetOperationPreferences{
			MaxConcurrentPercentage:    aws.Int32(100), // Deploy to all accounts simultaneously
			FailureTolerancePercentage: aws.Int32(50),  // Allow 50% of accounts to fail
			ConcurrencyMode:            cfntypes.ConcurrencyModeStrictFailureTolerance,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create stack instances: %w", err)
	}

	deployOperationId := *deployResult.OperationId
	logSuccess(fmt.Sprintf("Stack instances deployment started (Operation ID: %s)", deployOperationId))

	// Monitor deployment progress
	logInfo("Monitoring deployment progress...")
	for {
		status, err := m.cfnClient.DescribeStackSetOperation(ctx, &cloudformation.DescribeStackSetOperationInput{
			StackSetName: aws.String(m.stackSetName),
			OperationId:  aws.String(deployOperationId),
		})

		if err != nil {
			logWarning(fmt.Sprintf("Failed to check deployment status: %v", err))
			break
		}

		switch status.StackSetOperation.Status {
		case cfntypes.StackSetOperationStatusSucceeded:
			logSuccess("Deployment completed successfully!")

			// Show deployment summary
			if err := m.showDeploymentSummary(ctx, deployOperationId); err != nil {
				logWarning(fmt.Sprintf("Failed to show deployment summary: %v", err))
			}
			return nil

		case cfntypes.StackSetOperationStatusFailed, cfntypes.StackSetOperationStatusStopped:
			logError("Deployment failed!")
			if err := m.showDeploymentSummary(ctx, deployOperationId); err != nil {
				logWarning(fmt.Sprintf("Failed to show deployment summary: %v", err))
			}
			return fmt.Errorf("stackset deployment failed with status: %s", status.StackSetOperation.Status)

		case cfntypes.StackSetOperationStatusRunning:
			logInfo("Deployment in progress...")
			time.Sleep(10 * time.Second)
			continue

		default:
			logInfo(fmt.Sprintf("Deployment status: %s", status.StackSetOperation.Status))
			time.Sleep(5 * time.Second)
		}
	}

	return nil
}

func (m *IAMManager) showDeploymentSummary(ctx context.Context, operationId string) error {
	// Get operation results
	result, err := m.cfnClient.ListStackSetOperationResults(ctx, &cloudformation.ListStackSetOperationResultsInput{
		StackSetName: aws.String(m.stackSetName),
		OperationId:  aws.String(operationId),
	})

	if err != nil {
		return fmt.Errorf("failed to get operation results: %w", err)
	}

	fmt.Println()
	fmt.Printf("%sDeployment Summary:%s\n", colorCyan, colorReset)
	fmt.Printf("Operation ID: %s\n", operationId)
	fmt.Printf("Total Accounts: %d\n", len(result.Summaries))
	fmt.Printf("Concurrency: 100%% (parallel deployment to all accounts)\n")
	fmt.Printf("Failure Tolerance: 50%% (up to %d accounts can fail)\n", len(result.Summaries)/2)
	fmt.Println()

	successCount := 0
	failureCount := 0

	for _, summary := range result.Summaries {
		status := summary.Status
		account := *summary.Account
		region := *summary.Region

		switch status {
		case cfntypes.StackSetOperationResultStatusSucceeded:
			logSuccess(fmt.Sprintf("✓ Account %s (%s): %s", account, region, status))
			successCount++
		case cfntypes.StackSetOperationResultStatusFailed:
			reason := "Unknown error"
			if summary.StatusReason != nil {
				reason = *summary.StatusReason
			}
			logError(fmt.Sprintf("✗ Account %s (%s): %s - %s", account, region, status, reason))
			failureCount++
		default:
			logInfo(fmt.Sprintf("• Account %s (%s): %s", account, region, status))
		}
	}

	fmt.Println()
	fmt.Printf("%sResults: %d successful, %d failed%s\n", colorCyan, successCount, failureCount, colorReset)
	return nil
}

func (m *IAMManager) stackSetStatus(ctx context.Context) error {
	printHeader("StackSet Status Check")

	// Check if StackSet exists
	stackSetResult, err := m.cfnClient.DescribeStackSet(ctx, &cloudformation.DescribeStackSetInput{
		StackSetName: aws.String(m.stackSetName),
	})

	if err != nil {
		logWarning(fmt.Sprintf("StackSet '%s' not found", m.stackSetName))
		logInfo("Run 'stackset-deploy' to create and deploy the StackSet")
		return nil
	}

	stackSet := stackSetResult.StackSet
	fmt.Printf("%sStackSet Information:%s\n", colorCyan, colorReset)
	fmt.Printf("Name: %s\n", *stackSet.StackSetName)
	fmt.Printf("Status: %s\n", stackSet.Status)
	fmt.Printf("Description: %s\n", *stackSet.Description)
	fmt.Printf("Permission Model: %s\n", stackSet.PermissionModel)
	fmt.Println()

	// List stack instances with pagination
	var allInstances []cfntypes.StackInstanceSummary
	var nextToken *string

	for {
		input := &cloudformation.ListStackInstancesInput{
			StackSetName: aws.String(m.stackSetName),
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		instancesResult, err := m.cfnClient.ListStackInstances(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to list stack instances: %w", err)
		}

		allInstances = append(allInstances, instancesResult.Summaries...)

		if instancesResult.NextToken == nil {
			break
		}
		nextToken = instancesResult.NextToken
	}

	if len(allInstances) == 0 {
		logWarning("No stack instances found")
		logInfo("The StackSet exists but has not been deployed to any accounts")
		return nil
	}

	fmt.Printf("%sStack Instances (%d total):%s\n", colorCyan, len(allInstances), colorReset)
	fmt.Println()

	successCount := 0
	failureCount := 0
	otherCount := 0

	for _, instance := range allInstances {
		account := *instance.Account
		region := *instance.Region
		status := instance.Status

		switch status {
		case cfntypes.StackInstanceStatusCurrent:
			logSuccess(fmt.Sprintf("✓ Account %s (%s): %s", account, region, status))
			successCount++
		case cfntypes.StackInstanceStatusOutdated:
			reason := "Unknown error"
			if instance.StatusReason != nil {
				reason = *instance.StatusReason
			}
			logError(fmt.Sprintf("✗ Account %s (%s): %s - %s", account, region, status, reason))
			failureCount++
		case cfntypes.StackInstanceStatusInoperable:
			reason := "Unknown error"
			if instance.StatusReason != nil {
				reason = *instance.StatusReason
			}
			logError(fmt.Sprintf("✗ Account %s (%s): %s - %s", account, region, status, reason))
			failureCount++
		default:
			logInfo(fmt.Sprintf("• Account %s (%s): %s", account, region, status))
			otherCount++
		}
	}

	fmt.Println()
	fmt.Printf("%sStatus Summary: %d current, %d failed, %d other%s\n", colorCyan, successCount, failureCount, otherCount, colorReset)

	// Show recent operations
	operationsResult, err := m.cfnClient.ListStackSetOperations(ctx, &cloudformation.ListStackSetOperationsInput{
		StackSetName: aws.String(m.stackSetName),
		MaxResults:   aws.Int32(5),
	})

	if err == nil && len(operationsResult.Summaries) > 0 {
		fmt.Println()
		fmt.Printf("%sRecent Operations:%s\n", colorCyan, colorReset)

		for _, operation := range operationsResult.Summaries {
			operationId := *operation.OperationId
			status := operation.Status
			action := operation.Action

			timestamp := ""
			if operation.CreationTimestamp != nil {
				timestamp = operation.CreationTimestamp.Format("2006-01-02 15:04:05")
			}

			switch status {
			case cfntypes.StackSetOperationStatusSucceeded:
				logSuccess(fmt.Sprintf("✓ %s: %s (%s) - %s", operationId[:8], action, status, timestamp))
			case cfntypes.StackSetOperationStatusFailed:
				logError(fmt.Sprintf("✗ %s: %s (%s) - %s", operationId[:8], action, status, timestamp))
			default:
				logInfo(fmt.Sprintf("• %s: %s (%s) - %s", operationId[:8], action, status, timestamp))
			}
		}
	}

	return nil
}

func (m *IAMManager) deleteStackSet(ctx context.Context) error {
	printHeader("Delete StackSet")

	// Check if StackSet exists
	stackSetResult, err := m.cfnClient.DescribeStackSet(ctx, &cloudformation.DescribeStackSetInput{
		StackSetName: aws.String(m.stackSetName),
	})

	if err != nil {
		logWarning(fmt.Sprintf("StackSet '%s' not found", m.stackSetName))
		return nil
	}

	stackSet := stackSetResult.StackSet
	logInfo(fmt.Sprintf("Found StackSet: %s", *stackSet.StackSetName))
	fmt.Println()

	// List stack instances with pagination
	allInstances, err := m.listAllStackInstances(ctx)
	if err != nil {
		return fmt.Errorf("failed to list stack instances: %w", err)
	}

	logWarning("This will permanently delete:")
	fmt.Printf("  - StackSet: %s\n", m.stackSetName)
	if len(allInstances) > 0 {
		fmt.Printf("  - All %d stack instances across organization accounts\n", len(allInstances))
		fmt.Printf("  - IAM Manager role '%s' from ALL organization accounts\n", m.orgRoleName)
		fmt.Println("  - All managed and inline policies attached to these roles")
		fmt.Println("  - All CloudFormation stacks created by the StackSet")
		fmt.Println()
		fmt.Printf("%sIMPORTANT: This will remove cross-account access roles from ALL accounts%s\n", colorYellow, colorReset)
		fmt.Printf("%sYou will lose the ability to manage IAM in those accounts until redeployed%s\n", colorYellow, colorReset)
	}
	fmt.Println()

	if !m.confirmAction("Are you sure you want to continue?") {
		logInfo("Operation cancelled")
		return nil
	}

	// First, manually remove IAM roles from all accounts before deleting stack instances
	if len(allInstances) > 0 {
		logInfo("Removing IAM roles from all accounts...")
		if err := m.removeIAMRolesFromAccounts(ctx, allInstances); err != nil {
			logWarning(fmt.Sprintf("Failed to remove IAM roles from some accounts: %v", err))
			logInfo("Continuing with StackSet deletion...")
		}
	}

	// Delete all stack instances
	if len(allInstances) > 0 {
		logInfo("Deleting all stack instances...")
		logInfo("Deletion Configuration:")
		logInfo("  • Parallel deletion from ALL accounts simultaneously (100% concurrency)")
		logInfo("  • High failure tolerance: up to 50% of accounts can fail")
		logInfo("  • Strict failure tolerance mode for better resilience")

		// Get the root OU for deletion
		rootOU, err := m.getRootOU(ctx)
		if err != nil {
			return err
		}

		deleteResult, err := m.cfnClient.DeleteStackInstances(ctx, &cloudformation.DeleteStackInstancesInput{
			StackSetName: aws.String(m.stackSetName),
			DeploymentTargets: &cfntypes.DeploymentTargets{
				OrganizationalUnitIds: []string{rootOU},
			},
			Regions:      []string{m.regions},
			RetainStacks: aws.Bool(false),
			OperationPreferences: &cfntypes.StackSetOperationPreferences{
				MaxConcurrentPercentage:    aws.Int32(100), // Delete from all accounts simultaneously
				FailureTolerancePercentage: aws.Int32(50),  // Allow 50% of accounts to fail
				ConcurrencyMode:            cfntypes.ConcurrencyModeStrictFailureTolerance,
			},
		})

		if err != nil {
			return fmt.Errorf("failed to delete stack instances: %w", err)
		}

		deleteOperationId := *deleteResult.OperationId
		logSuccess(fmt.Sprintf("Stack instances deletion started (Operation ID: %s)", deleteOperationId))

		// Monitor deletion progress
		logInfo("Monitoring deletion progress...")
		for {
			status, err := m.cfnClient.DescribeStackSetOperation(ctx, &cloudformation.DescribeStackSetOperationInput{
				StackSetName: aws.String(m.stackSetName),
				OperationId:  aws.String(deleteOperationId),
			})

			if err != nil {
				logWarning(fmt.Sprintf("Failed to check deletion status: %v", err))
				break
			}

			switch status.StackSetOperation.Status {
			case cfntypes.StackSetOperationStatusSucceeded:
				logSuccess("All stack instances deleted successfully!")
				break

			case cfntypes.StackSetOperationStatusFailed, cfntypes.StackSetOperationStatusStopped:
				logError("Stack instances deletion failed!")
				if err := m.showDeploymentSummary(ctx, deleteOperationId); err != nil {
					logWarning(fmt.Sprintf("Failed to show deletion summary: %v", err))
				}
				return fmt.Errorf("stack instances deletion failed with status: %s", status.StackSetOperation.Status)

			case cfntypes.StackSetOperationStatusRunning:
				logInfo("Deletion in progress...")
				time.Sleep(10 * time.Second)
				continue

			default:
				logInfo(fmt.Sprintf("Deletion status: %s", status.StackSetOperation.Status))
				time.Sleep(5 * time.Second)
				continue
			}

			break
		}

		// Verify all instances are deleted
		logInfo("Verifying all instances are deleted...")
		for retries := 0; retries < 10; retries++ {
			instancesResult, err := m.cfnClient.ListStackInstances(ctx, &cloudformation.ListStackInstancesInput{
				StackSetName: aws.String(m.stackSetName),
			})

			if err != nil {
				logWarning(fmt.Sprintf("Failed to verify instance deletion: %v", err))
				break
			}

			if len(instancesResult.Summaries) == 0 {
				logSuccess("All instances successfully deleted")
				break
			}

			logInfo(fmt.Sprintf("Still %d instances remaining, waiting...", len(instancesResult.Summaries)))
			time.Sleep(5 * time.Second)
		}
	}

	// Delete the StackSet itself
	logInfo("Deleting StackSet...")
	_, err = m.cfnClient.DeleteStackSet(ctx, &cloudformation.DeleteStackSetInput{
		StackSetName: aws.String(m.stackSetName),
	})

	if err != nil {
		return fmt.Errorf("failed to delete StackSet: %w", err)
	}

	logSuccess("StackSet deleted successfully!")

	m.printDeletionSummary()
	return nil
}

func (m *IAMManager) removeIAMRolesFromAccounts(ctx context.Context, instances []cfntypes.StackInstanceSummary) error {
	logInfo("Parallel IAM role removal across all accounts...")
	
	// Extract unique account IDs
	accountSet := make(map[string]bool)
	for _, instance := range instances {
		accountSet[*instance.Account] = true
	}
	
	// Convert to slice for parallel processing
	var accountIDs []string
	for accountID := range accountSet {
		accountIDs = append(accountIDs, accountID)
	}
	
	logInfo(fmt.Sprintf("Removing role '%s' from %d accounts", m.orgRoleName, len(accountIDs)))
	
	// Use channels to collect results from parallel operations
	type accountResult struct {
		accountID string
		success   bool
		error     error
	}
	
	resultChan := make(chan accountResult, len(accountIDs))
	
	// Process each account in parallel
	for _, accountID := range accountIDs {
		go func(accID string) {
			success := m.removeRoleFromAccount(ctx, accID)
			resultChan <- accountResult{
				accountID: accID,
				success:   success,
				error:     nil, // We're handling errors internally in removeRoleFromAccount
			}
		}(accountID)
	}
	
	// Collect results
	successCount := 0
	failureCount := 0
	
	for i := 0; i < len(accountIDs); i++ {
		result := <-resultChan
		if result.success {
			successCount++
		} else {
			failureCount++
		}
	}
	
	logInfo(fmt.Sprintf("Role removal completed: %d successful, %d failed", successCount, failureCount))
	
	if failureCount > 0 {
		logWarning(fmt.Sprintf("%d accounts had failures during role removal", failureCount))
		logInfo("This is expected for accounts where the role doesn't exist or access is denied")
	}
	
	return nil
}

func (m *IAMManager) removeRoleFromAccount(ctx context.Context, accountID string) bool {
	// Skip the master account (current account)
	if accountID == m.accountID {
		logInfo(fmt.Sprintf("Skipping master account %s", accountID))
		return true
	}
	
	// Create credentials for cross-account access
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, m.orgRoleName)
	externalID := fmt.Sprintf("%s-iam-manager", accountID)
	
	logInfo(fmt.Sprintf("Attempting to remove role from account %s...", accountID))
	
	// Assume role in the target account
	result, err := m.stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("iam-manager-cleanup"),
		ExternalId:      aws.String(externalID),
		DurationSeconds: aws.Int32(900), // 15 minutes
	})
	
	if err != nil {
		logWarning(fmt.Sprintf("Cannot assume role in account %s: %v", accountID, err))
		logInfo(fmt.Sprintf("This is expected if the role doesn't exist in account %s", accountID))
		return false
	}
	
	// Create IAM client with assumed role credentials
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logWarning(fmt.Sprintf("Failed to load config for account %s: %v", accountID, err))
		return false
	}
	
	// Create temporary credentials
	creds := result.Credentials
	cfg.Credentials = aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     *creds.AccessKeyId,
			SecretAccessKey: *creds.SecretAccessKey,
			SessionToken:    *creds.SessionToken,
		}, nil
	}))
	
	targetIAMClient := iam.NewFromConfig(cfg)
	
	// List and detach managed policies from the role
	listPoliciesResult, err := targetIAMClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(m.orgRoleName),
	})
	
	if err != nil {
		logWarning(fmt.Sprintf("Failed to list policies for role in account %s: %v", accountID, err))
	} else {
		// Detach all managed policies
		for _, policy := range listPoliciesResult.AttachedPolicies {
			_, err := targetIAMClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
				RoleName:  aws.String(m.orgRoleName),
				PolicyArn: policy.PolicyArn,
			})
			if err != nil {
				logWarning(fmt.Sprintf("Failed to detach policy %s from role in account %s: %v", *policy.PolicyArn, accountID, err))
			}
		}
	}
	
	// List and delete inline policies
	listInlinePoliciesResult, err := targetIAMClient.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: aws.String(m.orgRoleName),
	})
	
	if err != nil {
		logWarning(fmt.Sprintf("Failed to list inline policies for role in account %s: %v", accountID, err))
	} else {
		// Delete all inline policies
		for _, policyName := range listInlinePoliciesResult.PolicyNames {
			_, err := targetIAMClient.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
				RoleName:   aws.String(m.orgRoleName),
				PolicyName: aws.String(policyName),
			})
			if err != nil {
				logWarning(fmt.Sprintf("Failed to delete inline policy %s from role in account %s: %v", policyName, accountID, err))
			}
		}
	}
	
	// Finally, delete the role itself
	_, err = targetIAMClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(m.orgRoleName),
	})
	
	if err != nil {
		logWarning(fmt.Sprintf("Failed to delete role in account %s: %v", accountID, err))
		return false
	}
	
	logSuccess(fmt.Sprintf("Successfully removed role from account %s", accountID))
	return true
}

func (m *IAMManager) printDeletionSummary() {
	fmt.Println()
	fmt.Println("=======================================================================")
	fmt.Printf("%s              STACKSET DELETION COMPLETE                   %s\n", colorGreen, colorReset)
	fmt.Println("=======================================================================")
	fmt.Println()
	fmt.Printf("%sDeleted Resources:%s\n", colorBlue, colorReset)
	fmt.Printf("  • StackSet: %s\n", m.stackSetName)
	fmt.Println("  • All stack instances across organization accounts")
	fmt.Printf("  • IAM Manager roles: %s (manually removed from each account)\n", m.orgRoleName)
	fmt.Println("  • All associated IAM policies (managed and inline)")
	fmt.Println("  • CloudFormation stacks created by the StackSet")
	fmt.Println()
	fmt.Printf("%sNote: Master account IAM user and policies were preserved%s\n", colorYellow, colorReset)
	fmt.Println("=======================================================================")
}

func (m *IAMManager) showStatus(ctx context.Context) error {
	printHeader("AWS IAM Manager - Status Check")

	fmt.Printf("%sEnvironment Information:%s\n", colorCyan, colorReset)
	fmt.Printf("Account ID: %s\n", m.accountID)
	fmt.Printf("Region: %s\n", m.regions)
	fmt.Printf("User Name: %s\n", m.userName)
	fmt.Printf("Role Name: %s\n", m.roleName)
	fmt.Printf("StackSet Name: %s\n", m.stackSetName)
	fmt.Println()

	// Check user status
	fmt.Printf("%sIAM User Status:%s\n", colorCyan, colorReset)
	userResult, err := m.iamClient.GetUser(ctx, &iam.GetUserInput{
		UserName: aws.String(m.userName),
	})

	if err != nil {
		logWarning(fmt.Sprintf("User '%s' does not exist", m.userName))
		logInfo("Run 'deploy' command to create the user")
	} else {
		user := userResult.User
		logSuccess(fmt.Sprintf("User '%s' exists", m.userName))
		fmt.Printf("  ARN: %s\n", *user.Arn)
		fmt.Printf("  Created: %s\n", user.CreateDate.Format("2006-01-02 15:04:05"))

		// Check access keys
		accessKeysResult, err := m.iamClient.ListAccessKeys(ctx, &iam.ListAccessKeysInput{
			UserName: aws.String(m.userName),
		})

		if err == nil {
			if len(accessKeysResult.AccessKeyMetadata) > 0 {
				logSuccess(fmt.Sprintf("Access keys: %d configured", len(accessKeysResult.AccessKeyMetadata)))
				for _, key := range accessKeysResult.AccessKeyMetadata {
					fmt.Printf("    %s (%s) - Created: %s\n", *key.AccessKeyId, key.Status, key.CreateDate.Format("2006-01-02 15:04:05"))
				}
			} else {
				logWarning("No access keys found")
			}
		}

		// Check attached policies
		attachedPoliciesResult, err := m.iamClient.ListAttachedUserPolicies(ctx, &iam.ListAttachedUserPoliciesInput{
			UserName: aws.String(m.userName),
		})

		if err == nil {
			if len(attachedPoliciesResult.AttachedPolicies) > 0 {
				logSuccess(fmt.Sprintf("Attached policies: %d", len(attachedPoliciesResult.AttachedPolicies)))
				for _, policy := range attachedPoliciesResult.AttachedPolicies {
					fmt.Printf("    %s\n", *policy.PolicyName)
				}
			} else {
				logWarning("No policies attached")
			}
		}
	}
	fmt.Println()

	// Check custom policy status
	fmt.Printf("%sCustom Policy Status:%s\n", colorCyan, colorReset)
	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/%s", m.accountID, m.policyName)
	_, err = m.iamClient.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: aws.String(policyArn),
	})

	if err != nil {
		logWarning(fmt.Sprintf("Custom policy '%s' does not exist", m.policyName))
	} else {
		logSuccess(fmt.Sprintf("Custom policy '%s' exists", m.policyName))
		fmt.Printf("  ARN: %s\n", policyArn)
	}
	fmt.Println()

	// Check IAM role status (alternative to user)
	fmt.Printf("%sIAM Role Status (Alternative Access Method):%s\n", colorCyan, colorReset)
	_, err = m.iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String(m.roleName),
	})

	if err != nil {
		logInfo(fmt.Sprintf("Role '%s' does not exist", m.roleName))
		logInfo("Run 'create-role' command to create it as an alternative to user-based access")
	} else {
		logSuccess(fmt.Sprintf("Role '%s' exists", m.roleName))
		fmt.Printf("  ARN: arn:aws:iam::%s:role/%s\n", m.accountID, m.roleName)
	}
	fmt.Println()

	// Check StackSet status
	fmt.Printf("%sStackSet Status:%s\n", colorCyan, colorReset)
	stackSetResult, err := m.cfnClient.DescribeStackSet(ctx, &cloudformation.DescribeStackSetInput{
		StackSetName: aws.String(m.stackSetName),
	})

	if err != nil {
		logWarning(fmt.Sprintf("StackSet '%s' does not exist", m.stackSetName))
		logInfo("Run 'stackset-deploy' command to create and deploy StackSet")
	} else {
		stackSet := stackSetResult.StackSet
		logSuccess(fmt.Sprintf("StackSet '%s' exists", m.stackSetName))
		fmt.Printf("  Status: %s\n", stackSet.Status)
		fmt.Printf("  Permission Model: %s\n", stackSet.PermissionModel)

		// Check stack instances
		allStackInstances, err := m.listAllStackInstances(ctx)

		if err == nil {
			totalInstances := len(allStackInstances)
			if totalInstances > 0 {
				successCount := 0
				failureCount := 0
				otherCount := 0

				for _, instance := range allStackInstances {
					switch instance.Status {
					case cfntypes.StackInstanceStatusCurrent:
						successCount++
					case cfntypes.StackInstanceStatusOutdated, cfntypes.StackInstanceStatusInoperable:
						failureCount++
					default:
						otherCount++
					}
				}

				logSuccess(fmt.Sprintf("Stack instances: %d total", totalInstances))
				if successCount > 0 {
					fmt.Printf("    ✓ %d current\n", successCount)
				}
				if failureCount > 0 {
					fmt.Printf("    ✗ %d failed\n", failureCount)
				}
				if otherCount > 0 {
					fmt.Printf("    • %d other states\n", otherCount)
				}
			} else {
				logWarning("No stack instances found")
				logInfo("StackSet exists but has not been deployed to any accounts")
			}
		}
	}
	fmt.Println()

	// Check organization access
	fmt.Printf("%sOrganizations Integration:%s\n", colorCyan, colorReset)
	org, err := m.orgsClient.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		logWarning("Unable to access AWS Organizations")
		logInfo("Make sure you have appropriate permissions and are in the master account")
	} else {
		logSuccess("AWS Organizations access confirmed")
		fmt.Printf("  Organization ID: %s\n", *org.Organization.Id)
		fmt.Printf("  Master Account: %s\n", *org.Organization.MasterAccountId)

		// List organization accounts with pagination
		activeAccounts := 0
		totalAccounts := 0
		var nextToken *string

		for {
			input := &organizations.ListAccountsInput{}
			if nextToken != nil {
				input.NextToken = nextToken
			}

			accountsResult, err := m.orgsClient.ListAccounts(ctx, input)
			if err != nil {
				logWarning(fmt.Sprintf("Unable to list organization accounts: %v", err))
				break
			}

			for _, account := range accountsResult.Accounts {
				totalAccounts++
				if account.Status == orgtypes.AccountStatusActive {
					activeAccounts++
				}
			}

			if accountsResult.NextToken == nil {
				break
			}
			nextToken = accountsResult.NextToken
		}

		if totalAccounts > 0 {
			logSuccess(fmt.Sprintf("Organization has %d total accounts (%d active)", totalAccounts, activeAccounts))
		}
	}
	fmt.Println()

	// Check StackSets trusted service
	fmt.Printf("%sStackSets Service Integration:%s\n", colorCyan, colorReset)
	services, err := m.orgsClient.ListAWSServiceAccessForOrganization(ctx, &organizations.ListAWSServiceAccessForOrganizationInput{})
	if err == nil {
		stackSetsEnabled := false
		for _, service := range services.EnabledServicePrincipals {
			if *service.ServicePrincipal == "stacksets.cloudformation.amazonaws.com" {
				stackSetsEnabled = true
				break
			}
		}

		if stackSetsEnabled {
			logSuccess("StackSets trusted service access is enabled")
		} else {
			logWarning("StackSets trusted service access is NOT enabled")
			logInfo("This is required for SERVICE_MANAGED StackSet operations")
		}
	} else {
		logWarning("Unable to check StackSets service integration")
	}

	// Check StackSet administration role
	_, err = m.iamClient.GetRole(ctx, &iam.GetRoleInput{
		RoleName: aws.String("AWSCloudFormationStackSetAdministrationRole"),
	})

	if err != nil {
		logWarning("StackSet administration role does not exist")
		logInfo("This role is required for SELF_MANAGED StackSet operations")
	} else {
		logSuccess("StackSet administration role exists")
	}

	fmt.Println()
	m.printStatusSummary()
	return nil
}

func (m *IAMManager) printStatusSummary() {
	fmt.Println("=======================================================================")
	fmt.Printf("%s                    STATUS SUMMARY                           %s\n", colorBlue, colorReset)
	fmt.Println("=======================================================================")
	fmt.Println()
	fmt.Printf("%sReady to Use:%s\n", colorGreen, colorReset)
	fmt.Println("  ✓ Check that IAM user/role exists and has access keys")
	fmt.Println("  ✓ Ensure StackSet is deployed to organization accounts")
	fmt.Println("  ✓ Verify AWS Organizations integration is working")
	fmt.Println()
	fmt.Printf("%sNext Steps (if needed):%s\n", colorYellow, colorReset)
	fmt.Println("  1. Run 'deploy' to create IAM user with required permissions")
	fmt.Println("  2. Run 'stackset-deploy' to deploy roles to all organization accounts")
	fmt.Println("  3. Copy credentials to .env file and start the web application")
	fmt.Println("  4. Use 'status' command regularly to monitor deployment health")
	fmt.Println("=======================================================================")
}

// listAllStackInstances returns all stack instances with pagination
func (m *IAMManager) listAllStackInstances(ctx context.Context) ([]cfntypes.StackInstanceSummary, error) {
	var allInstances []cfntypes.StackInstanceSummary
	var nextToken *string

	for {
		input := &cloudformation.ListStackInstancesInput{
			StackSetName: aws.String(m.stackSetName),
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		result, err := m.cfnClient.ListStackInstances(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list stack instances: %w", err)
		}

		allInstances = append(allInstances, result.Summaries...)

		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	return allInstances, nil
}
