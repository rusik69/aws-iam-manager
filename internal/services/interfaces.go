package services

import "github.com/rusik69/aws-iam-manager/internal/models"

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
	// Cache management methods
	ClearCache()
	InvalidateAccountCache(accountID string)
	InvalidateUserCache(accountID, username string)
	InvalidatePublicIPsCache()
}
