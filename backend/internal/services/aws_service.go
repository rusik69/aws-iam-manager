// Package services provides business logic and external service integrations
package services

import (
	"fmt"
	"log"
	"os"
	"time"

	"aws-iam-manager/internal/models"
	"aws-iam-manager/pkg/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/sts"
)

type AWSService struct {
	masterSession *session.Session
	config        *config.Config
}

func NewAWSService(cfg *config.Config) *AWSService {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	if region == "" {
		region = "us-east-1"
	}

	log.Printf("[INFO] Initializing AWS service with region: %s", region)
	
	// Log whether credentials are set (but not the actual values)
	if accessKey == "" {
		log.Printf("[WARN] AWS_ACCESS_KEY_ID is not set")
	} else {
		log.Printf("[INFO] AWS_ACCESS_KEY_ID is set (length: %d)", len(accessKey))
	}
	
	if secretKey == "" {
		log.Printf("[WARN] AWS_SECRET_ACCESS_KEY is not set")
	} else {
		log.Printf("[INFO] AWS_SECRET_ACCESS_KEY is set (length: %d)", len(secretKey))
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create AWS session: %v", err)
		panic(fmt.Sprintf("Failed to create AWS session: %v", err))
	}

	log.Printf("[INFO] AWS service initialized successfully with role name: %s", cfg.RoleName)
	return &AWSService{
		masterSession: sess,
		config:        cfg,
	}
}

func (s *AWSService) getSessionForAccount(accountID string) (*session.Session, error) {
	start := time.Now()
	if accountID == "" {
		log.Printf("[DEBUG] Using master session (no account ID provided)")
		return s.masterSession, nil
	}

	log.Printf("[DEBUG] Assuming role for account: %s", accountID)
	stsClient := sts.New(s.masterSession)
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, s.config.RoleName)
	log.Printf("[DEBUG] Role ARN: %s", roleArn)

	result, err := stsClient.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("IAMManager"),
		DurationSeconds: aws.Int64(3600),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to assume role %s: %v (took %v)", roleArn, err, time.Since(start))
		return nil, fmt.Errorf("failed to assume role: %v", err)
	}

	creds := result.Credentials
	sess, err := session.NewSession(&aws.Config{
		Region: s.masterSession.Config.Region,
		Credentials: credentials.NewStaticCredentials(
			*creds.AccessKeyId,
			*creds.SecretAccessKey,
			*creds.SessionToken,
		),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create session with assumed role: %v (took %v)", err, time.Since(start))
		return nil, fmt.Errorf("failed to create session with assumed role: %v", err)
	}

	log.Printf("[DEBUG] Successfully assumed role for account %s (took %v)", accountID, time.Since(start))
	return sess, nil
}

func (s *AWSService) ListAccounts() ([]models.Account, error) {
	start := time.Now()
	log.Printf("[DEBUG] Starting ListAccounts operation")
	
	orgClient := organizations.New(s.masterSession)

	result, err := orgClient.ListAccounts(&organizations.ListAccountsInput{})
	if err != nil {
		log.Printf("[ERROR] AWS Organizations ListAccounts failed: %v (took %v)", err, time.Since(start))
		return nil, err
	}

	var accounts []models.Account
	for _, account := range result.Accounts {
		accounts = append(accounts, models.Account{
			ID:   *account.Id,
			Name: *account.Name,
		})
	}

	log.Printf("[DEBUG] ListAccounts completed: found %d accounts (took %v)", len(accounts), time.Since(start))
	return accounts, nil
}

func (s *AWSService) ListUsers(accountID string) ([]models.User, error) {
	start := time.Now()
	log.Printf("[DEBUG] Starting ListUsers for account: %s", accountID)
	
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		log.Printf("[ERROR] Failed to get session for account %s: %v (took %v)", accountID, err, time.Since(start))
		return nil, err
	}

	iamClient := iam.New(sess)
	result, err := iamClient.ListUsers(&iam.ListUsersInput{})
	if err != nil {
		log.Printf("[ERROR] AWS IAM ListUsers failed for account %s: %v (took %v)", accountID, err, time.Since(start))
		return nil, err
	}

	log.Printf("[DEBUG] Found %d users in account %s, processing details...", len(result.Users), accountID)
	var users []models.User
	for _, user := range result.Users {
		// Check if user has password
		passwordSet := false
		_, err := iamClient.GetLoginProfile(&iam.GetLoginProfileInput{
			UserName: user.UserName,
		})
		if err == nil {
			passwordSet = true
		} else {
			log.Printf("[DEBUG] User %s in account %s has no login profile", *user.UserName, accountID)
		}

		// Get access keys
		keysResult, err := iamClient.ListAccessKeys(&iam.ListAccessKeysInput{
			UserName: user.UserName,
		})
		var accessKeys []models.AccessKey
		if err == nil {
			for _, key := range keysResult.AccessKeyMetadata {
				accessKeys = append(accessKeys, models.AccessKey{
					AccessKeyID: *key.AccessKeyId,
					Status:      *key.Status,
					CreateDate:  *key.CreateDate,
				})
			}
			log.Printf("[DEBUG] User %s in account %s has %d access keys", *user.UserName, accountID, len(accessKeys))
		} else {
			log.Printf("[WARN] Failed to get access keys for user %s in account %s: %v", *user.UserName, accountID, err)
		}

		users = append(users, models.User{
			Username:    *user.UserName,
			UserID:      *user.UserId,
			Arn:         *user.Arn,
			CreateDate:  *user.CreateDate,
			PasswordSet: passwordSet,
			AccessKeys:  accessKeys,
		})
	}

	log.Printf("[DEBUG] ListUsers completed for account %s: processed %d users (took %v)", accountID, len(users), time.Since(start))
	return users, nil
}

func (s *AWSService) GetUser(accountID, username string) (*models.User, error) {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return nil, err
	}

	iamClient := iam.New(sess)
	result, err := iamClient.GetUser(&iam.GetUserInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return nil, err
	}

	user := result.User

	// Check if user has password
	passwordSet := false
	_, err = iamClient.GetLoginProfile(&iam.GetLoginProfileInput{
		UserName: user.UserName,
	})
	if err == nil {
		passwordSet = true
	}

	// Get access keys
	keysResult, err := iamClient.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: user.UserName,
	})
	var accessKeys []models.AccessKey
	if err == nil {
		for _, key := range keysResult.AccessKeyMetadata {
			accessKeys = append(accessKeys, models.AccessKey{
				AccessKeyID: *key.AccessKeyId,
				Status:      *key.Status,
				CreateDate:  *key.CreateDate,
			})
		}
	}

	userResponse := &models.User{
		Username:    *user.UserName,
		UserID:      *user.UserId,
		Arn:         *user.Arn,
		CreateDate:  *user.CreateDate,
		PasswordSet: passwordSet,
		AccessKeys:  accessKeys,
	}

	return userResponse, nil
}

func (s *AWSService) CreateAccessKey(accountID, username string) (map[string]any, error) {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return nil, err
	}

	iamClient := iam.New(sess)
	result, err := iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return nil, err
	}

	key := result.AccessKey
	response := map[string]any{
		"access_key_id":     *key.AccessKeyId,
		"secret_access_key": *key.SecretAccessKey,
		"status":            *key.Status,
		"create_date":       *key.CreateDate,
	}

	return response, nil
}

func (s *AWSService) DeleteAccessKey(accountID, username, keyID string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return err
	}

	iamClient := iam.New(sess)
	_, err = iamClient.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		UserName:    aws.String(username),
		AccessKeyId: aws.String(keyID),
	})
	return err
}

func (s *AWSService) RotateAccessKey(accountID, username, keyID string) (map[string]any, error) {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return nil, err
	}

	iamClient := iam.New(sess)

	// Create new key
	createResult, err := iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return nil, err
	}

	// Delete old key
	_, err = iamClient.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		UserName:    aws.String(username),
		AccessKeyId: aws.String(keyID),
	})
	if err != nil {
		// If deletion fails, try to delete the new key we just created
		_, _ = iamClient.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			UserName:    aws.String(username),
			AccessKeyId: createResult.AccessKey.AccessKeyId,
		})
		return nil, fmt.Errorf("failed to delete old key: %v", err)
	}

	key := createResult.AccessKey
	response := map[string]any{
		"access_key_id":     *key.AccessKeyId,
		"secret_access_key": *key.SecretAccessKey,
		"status":            *key.Status,
		"create_date":       *key.CreateDate,
		"message":           "Access key rotated successfully",
	}

	return response, nil
}
