package services

import (
	"context"

	"github.com/rusik69/aws-iam-manager/internal/models"
)

type AWSServiceInterface interface {
	ListAccounts() ([]models.Account, error)
	ListUsers(accountID string) ([]models.User, error)
	GetUser(accountID, username string) (*models.User, error)
	CreateAccessKey(accountID, username string) (map[string]any, error)
	DeleteAccessKey(accountID, username, keyID string) error
	RotateAccessKey(accountID, username, keyID string) (map[string]any, error)
	DeleteUser(accountID, username string) error
	DeleteUserPassword(accountID, username string) error
	RotateUserPassword(accountID, username string) (map[string]any, error)
	ListPublicIPs() ([]models.PublicIP, error)
	ListSecurityGroups() ([]models.SecurityGroup, error)
	ListSecurityGroupsByAccount(accountID string) ([]models.SecurityGroup, error)
	GetSecurityGroup(accountID, region, groupID string) (*models.SecurityGroup, error)
	DeleteSecurityGroup(accountID, region, groupID string) error
	// Cache management methods
	ClearCache()
	InvalidateAccountCache(accountID string)
	InvalidateUserCache(accountID, username string)
	InvalidatePublicIPsCache()
	InvalidateSecurityGroupsCache()
	InvalidateAccountSecurityGroupsCache(accountID string)
}

type AzureServiceInterface interface {
	ListEnterpriseApplications(ctx context.Context) ([]models.AzureEnterpriseApplication, error)
	GetEnterpriseApplication(ctx context.Context, appID string) (*models.AzureEnterpriseApplication, error)
	DeleteEnterpriseApplication(ctx context.Context, appID string) error
	// Cache management methods
	ClearCache()
	InvalidateEnterpriseApplicationsCache()
	InvalidateEnterpriseApplicationCache(appID string)
}
