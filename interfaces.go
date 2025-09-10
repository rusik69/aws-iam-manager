package main


type AWSServiceInterface interface {
	ListAccounts() ([]Account, error)
	ListUsers(accountID string) ([]User, error)
	GetUser(accountID, username string) (*User, error)
	CreateAccessKey(accountID, username string) (map[string]any, error)
	DeleteAccessKey(accountID, username, keyID string) error
	RotateAccessKey(accountID, username, keyID string) (map[string]any, error)
	DeleteUser(accountID, username string) error
}
