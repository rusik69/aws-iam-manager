// Package services provides business logic and external service integrations
package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"os"
	"sync"
	"time"

	"github.com/rusik69/aws-iam-manager/internal/config"
	"github.com/rusik69/aws-iam-manager/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/sts"
)

// CacheEntry represents a cached value with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// Cache represents the caching layer
type Cache struct {
	mu    sync.RWMutex
	items map[string]CacheEntry
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]CacheEntry),
	}
}

// Get retrieves a value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}
	
	if time.Now().After(entry.ExpiresAt) {
		// Item has expired, clean it up
		delete(c.items, key)
		return nil, false
	}
	
	return entry.Data, true
}

// Set stores a value in cache with TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items[key] = CacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a value from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.items, key)
}

// Clear removes all values from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]CacheEntry)
}

// DeletePattern removes all keys matching a pattern (simple prefix matching)
func (c *Cache) DeletePattern(pattern string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	for key := range c.items {
		if len(key) >= len(pattern) && key[:len(pattern)] == pattern {
			delete(c.items, key)
		}
	}
}

type AWSService struct {
	masterSession *session.Session
	config        config.Config
	cache         *Cache
	cacheTTL      time.Duration
}

func NewAWSService(cfg config.Config) *AWSService {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	if region == "" {
		region = "us-east-1"
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create AWS session: %v", err))
	}

	return &AWSService{
		masterSession: sess,
		config:        cfg,
		cache:         NewCache(),
		cacheTTL:      5 * time.Minute, // Default 5 minute TTL
	}
}

func (s *AWSService) getSessionForAccount(accountID string) (*session.Session, error) {
	if accountID == "" {
		return s.masterSession, nil
	}

	stsClient := sts.New(s.masterSession)
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, s.config.RoleName)

	// Generate ExternalId as per CloudFormation template: {accountId}-iam-manager
	externalId := fmt.Sprintf("%s-iam-manager", accountID)
	
	result, err := stsClient.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("IAMManager"),
		DurationSeconds: aws.Int64(3600),
		ExternalId:      aws.String(externalId),
	})
	if err != nil {
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
		return nil, fmt.Errorf("failed to create session with assumed role: %v", err)
	}

	return sess, nil
}

func (s *AWSService) ListAccounts() ([]models.Account, error) {
	const cacheKey = "accounts"
	
	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if accounts, ok := cached.([]models.Account); ok {
			return accounts, nil
		}
	}

	orgClient := organizations.New(s.masterSession)
	result, err := orgClient.ListAccounts(&organizations.ListAccountsInput{})
	if err != nil {
		return nil, err
	}

	var accounts []models.Account
	for _, account := range result.Accounts {
		// Test if we can access this account
		accessible := s.testAccountAccess(*account.Id)
		
		accounts = append(accounts, models.Account{
			ID:         *account.Id,
			Name:       *account.Name,
			Accessible: accessible,
		})
	}

	// Cache the result
	s.cache.Set(cacheKey, accounts, s.cacheTTL)
	
	return accounts, nil
}

// testAccountAccess checks if we can assume role in the given account
func (s *AWSService) testAccountAccess(accountID string) bool {
	_, err := s.getSessionForAccount(accountID)
	if err != nil {
		fmt.Printf("[WARNING] Cannot access account %s: %v\n", accountID, err)
		return false
	}
	return true
}

func (s *AWSService) ListUsers(accountID string) ([]models.User, error) {
	cacheKey := fmt.Sprintf("users:%s", accountID)
	
	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if users, ok := cached.([]models.User); ok {
			return users, nil
		}
	}

	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		fmt.Printf("[WARNING] Cannot access account %s, skipping user listing: %v\n", accountID, err)
		// Return empty list instead of error for inaccessible accounts
		return []models.User{}, nil
	}

	iamClient := iam.New(sess)
	result, err := iamClient.ListUsers(&iam.ListUsersInput{})
	if err != nil {
		return nil, err
	}

	var users []models.User
	for _, user := range result.Users {
		// Check if user has password
		passwordSet := false
		_, err := iamClient.GetLoginProfile(&iam.GetLoginProfileInput{
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

		users = append(users, models.User{
			Username:    *user.UserName,
			UserID:      *user.UserId,
			Arn:         *user.Arn,
			CreateDate:  *user.CreateDate,
			PasswordSet: passwordSet,
			AccessKeys:  accessKeys,
		})
	}

	// Cache the result
	s.cache.Set(cacheKey, users, s.cacheTTL)
	
	return users, nil
}

func (s *AWSService) GetUser(accountID, username string) (*models.User, error) {
	cacheKey := fmt.Sprintf("user:%s:%s", accountID, username)
	
	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if user, ok := cached.(*models.User); ok {
			return user, nil
		}
	}

	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot access account %s: %w", accountID, err)
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

	// Cache the result
	s.cache.Set(cacheKey, userResponse, s.cacheTTL)

	return userResponse, nil
}

func (s *AWSService) CreateAccessKey(accountID, username string) (map[string]any, error) {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	iamClient := iam.New(sess)
	result, err := iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return nil, err
	}

	// Invalidate related cache entries
	s.cache.Delete(fmt.Sprintf("user:%s:%s", accountID, username))
	s.cache.Delete(fmt.Sprintf("users:%s", accountID))
	// Also invalidate accounts cache in case it includes user/key counts
	s.cache.Delete("accounts")

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
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	iamClient := iam.New(sess)
	_, err = iamClient.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		UserName:    aws.String(username),
		AccessKeyId: aws.String(keyID),
	})
	
	if err == nil {
		// Invalidate related cache entries
		s.cache.Delete(fmt.Sprintf("user:%s:%s", accountID, username))
		s.cache.Delete(fmt.Sprintf("users:%s", accountID))
		// Also invalidate accounts cache in case it includes user/key counts
		s.cache.Delete("accounts")
	}
	
	return err
}

func (s *AWSService) RotateAccessKey(accountID, username, keyID string) (map[string]any, error) {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot access account %s: %w", accountID, err)
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

	// Invalidate related cache entries
	s.cache.Delete(fmt.Sprintf("user:%s:%s", accountID, username))
	s.cache.Delete(fmt.Sprintf("users:%s", accountID))
	// Also invalidate accounts cache in case it includes user/key counts
	s.cache.Delete("accounts")

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

func (s *AWSService) DeleteUser(accountID, username string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	iamClient := iam.New(sess)

	// First, delete all access keys for the user
	keysResult, err := iamClient.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return fmt.Errorf("failed to list access keys: %v", err)
	}

	for _, key := range keysResult.AccessKeyMetadata {
		_, err = iamClient.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			UserName:    aws.String(username),
			AccessKeyId: key.AccessKeyId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete access key %s: %v", *key.AccessKeyId, err)
		}
	}

	// Remove user from all groups
	groupsResult, err := iamClient.ListGroupsForUser(&iam.ListGroupsForUserInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return fmt.Errorf("failed to list groups for user: %v", err)
	}

	for _, group := range groupsResult.Groups {
		_, err = iamClient.RemoveUserFromGroup(&iam.RemoveUserFromGroupInput{
			UserName:  aws.String(username),
			GroupName: group.GroupName,
		})
		if err != nil {
			return fmt.Errorf("failed to remove user from group %s: %v", *group.GroupName, err)
		}
	}

	// Delete the login profile if it exists
	_, err = iamClient.DeleteLoginProfile(&iam.DeleteLoginProfileInput{
		UserName: aws.String(username),
	})
	// Ignore error if login profile doesn't exist
	
	// Finally, delete the user
	_, err = iamClient.DeleteUser(&iam.DeleteUserInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	// Invalidate related cache entries
	s.cache.Delete(fmt.Sprintf("user:%s:%s", accountID, username))
	s.cache.Delete(fmt.Sprintf("users:%s", accountID))
	// Also invalidate accounts cache in case it includes user/key counts
	s.cache.Delete("accounts")

	return nil
}

func (s *AWSService) ListPublicIPs() ([]models.PublicIP, error) {
	const cacheKey = "public-ips"
	
	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if ips, ok := cached.([]models.PublicIP); ok {
			return ips, nil
		}
	}

	accounts, err := s.ListAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %v", err)
	}

	// Filter accessible accounts
	var accessibleAccounts []models.Account
	for _, account := range accounts {
		if !account.Accessible {
			fmt.Printf("[WARNING] Skipping inaccessible account %s\n", account.ID)
			continue
		}
		accessibleAccounts = append(accessibleAccounts, account)
	}

	if len(accessibleAccounts) == 0 {
		return []models.PublicIP{}, nil
	}

	// Channel to collect results from goroutines
	type accountResult struct {
		ips []models.PublicIP
		err error
		accountID string
	}

	resultChan := make(chan accountResult, len(accessibleAccounts))
	var wg sync.WaitGroup

	// Process each account in parallel
	for _, account := range accessibleAccounts {
		wg.Add(1)
		go func(acc models.Account) {
			defer wg.Done()
			
			ips, err := s.getPublicIPsForAccount(acc)
			resultChan <- accountResult{
				ips: ips,
				err: err,
				accountID: acc.ID,
			}
		}(account)
	}

	// Wait for all goroutines to complete and close channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var allIPs []models.PublicIP
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("[WARNING] Failed to get IPs for account %s: %v\n", result.accountID, result.err)
			continue
		}
		allIPs = append(allIPs, result.ips...)
	}

	// Cache the result
	s.cache.Set(cacheKey, allIPs, s.cacheTTL)

	return allIPs, nil
}

// ClearCache clears all cached data
func (s *AWSService) ClearCache() {
	s.cache.Clear()
}

// InvalidateAccountCache invalidates cache entries for a specific account
func (s *AWSService) InvalidateAccountCache(accountID string) {
	s.cache.Delete(fmt.Sprintf("users:%s", accountID))
	s.cache.DeletePattern(fmt.Sprintf("user:%s:", accountID))
	// Also invalidate accounts cache
	s.cache.Delete("accounts")
}

// InvalidateUserCache invalidates cache entries for a specific user
func (s *AWSService) InvalidateUserCache(accountID, username string) {
	s.cache.Delete(fmt.Sprintf("user:%s:%s", accountID, username))
	s.cache.Delete(fmt.Sprintf("users:%s", accountID))
	// Also invalidate accounts cache in case it includes user/key counts
	s.cache.Delete("accounts")
}

// InvalidatePublicIPsCache invalidates the public IPs cache
func (s *AWSService) InvalidatePublicIPsCache() {
	s.cache.Delete("public-ips")
}

func (s *AWSService) DeleteUserPassword(accountID, username string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	iamClient := iam.New(sess)

	// Delete the login profile (console password)
	_, err = iamClient.DeleteLoginProfile(&iam.DeleteLoginProfileInput{
		UserName: aws.String(username),
	})
	if err != nil {
		// Check if it's because the login profile doesn't exist
		if awsErr, ok := err.(interface{ Code() string }); ok && awsErr.Code() == "NoSuchEntity" {
			return fmt.Errorf("user %s does not have a console password", username)
		}
		return fmt.Errorf("failed to delete user password: %v", err)
	}

	// Invalidate related cache entries
	s.cache.Delete(fmt.Sprintf("user:%s:%s", accountID, username))
	s.cache.Delete(fmt.Sprintf("users:%s", accountID))
	// Also invalidate accounts cache in case it includes user/password counts
	s.cache.Delete("accounts")

	return nil
}

func (s *AWSService) RotateUserPassword(accountID, username string) (map[string]any, error) {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	iamClient := iam.New(sess)

	// Check if user already has a login profile
	hasLoginProfile := true
	_, err = iamClient.GetLoginProfile(&iam.GetLoginProfileInput{
		UserName: aws.String(username),
	})
	if err != nil {
		if awsErr, ok := err.(interface{ Code() string }); ok && awsErr.Code() == "NoSuchEntity" {
			hasLoginProfile = false
		} else {
			return nil, fmt.Errorf("failed to check login profile: %v", err)
		}
	}

	// Generate a new password
	newPassword := s.generatePassword()

	if hasLoginProfile {
		// Update existing login profile
		_, err = iamClient.UpdateLoginProfile(&iam.UpdateLoginProfileInput{
			UserName: aws.String(username),
			Password: aws.String(newPassword),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update user password: %v", err)
		}
	} else {
		// Create new login profile
		_, err = iamClient.CreateLoginProfile(&iam.CreateLoginProfileInput{
			UserName: aws.String(username),
			Password: aws.String(newPassword),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create user password: %v", err)
		}
	}

	// Invalidate related cache entries
	s.cache.Delete(fmt.Sprintf("user:%s:%s", accountID, username))
	s.cache.Delete(fmt.Sprintf("users:%s", accountID))
	// Also invalidate accounts cache in case it includes user/password counts
	s.cache.Delete("accounts")

	response := map[string]any{
		"message":      "User password rotated successfully",
		"new_password": newPassword,
		"username":     username,
	}

	return response, nil
}

func (s *AWSService) getPublicIPsForAccount(account models.Account) ([]models.PublicIP, error) {
	sess, err := s.getSessionForAccount(account.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot access account %s: %w", account.ID, err)
	}

	// Get all regions
	ec2Client := ec2.New(sess)
	regionsResult, err := ec2Client.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %v", err)
	}

	if len(regionsResult.Regions) == 0 {
		return []models.PublicIP{}, nil
	}

	// Channel to collect results from region goroutines
	type regionResult struct {
		ips []models.PublicIP
		regionName string
	}

	resultChan := make(chan regionResult, len(regionsResult.Regions))
	var wg sync.WaitGroup

	// Process each region in parallel
	for _, region := range regionsResult.Regions {
		wg.Add(1)
		go func(regionName string) {
			defer wg.Done()
			
			// Create session for this region
			regionSess := sess.Copy(&aws.Config{Region: aws.String(regionName)})
			var regionIPs []models.PublicIP
			
			// Get EC2 instances
			ec2IPs, err := s.getEC2PublicIPs(regionSess, account, regionName)
			if err != nil {
				fmt.Printf("[WARNING] Failed to get EC2 IPs in region %s for account %s: %v\n", regionName, account.ID, err)
			} else {
				regionIPs = append(regionIPs, ec2IPs...)
			}

			// Get Load Balancers
			elbIPs, err := s.getELBPublicIPs(regionSess, account, regionName)
			if err != nil {
				fmt.Printf("[WARNING] Failed to get ELB IPs in region %s for account %s: %v\n", regionName, account.ID, err)
			} else {
				regionIPs = append(regionIPs, elbIPs...)
			}

			// Get NAT Gateways
			natIPs, err := s.getNATPublicIPs(regionSess, account, regionName)
			if err != nil {
				fmt.Printf("[WARNING] Failed to get NAT IPs in region %s for account %s: %v\n", regionName, account.ID, err)
			} else {
				regionIPs = append(regionIPs, natIPs...)
			}
			
			resultChan <- regionResult{
				ips: regionIPs,
				regionName: regionName,
			}
		}(*region.RegionName)
	}

	// Wait for all goroutines to complete and close channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var allIPs []models.PublicIP
	for result := range resultChan {
		allIPs = append(allIPs, result.ips...)
	}

	return allIPs, nil
}

func (s *AWSService) getEC2PublicIPs(sess *session.Session, account models.Account, region string) ([]models.PublicIP, error) {
	ec2Client := ec2.New(sess)
	var ips []models.PublicIP

	result, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			if instance.PublicIpAddress != nil && *instance.PublicIpAddress != "" {
				instanceName := ""
				for _, tag := range instance.Tags {
					if *tag.Key == "Name" {
						instanceName = *tag.Value
						break
					}
				}

				ips = append(ips, models.PublicIP{
					IPAddress:    *instance.PublicIpAddress,
					AccountID:    account.ID,
					AccountName:  account.Name,
					Region:       region,
					ResourceType: "EC2",
					ResourceID:   *instance.InstanceId,
					ResourceName: instanceName,
					State:        *instance.State.Name,
				})
			}
		}
	}

	return ips, nil
}

func (s *AWSService) getELBPublicIPs(sess *session.Session, account models.Account, region string) ([]models.PublicIP, error) {
	var ips []models.PublicIP

	// Get ALB/NLB (ELBv2)
	elbv2IPs, err := s.getELBv2PublicIPs(sess, account, region)
	if err != nil {
		fmt.Printf("[WARNING] Failed to get ELBv2 IPs in region %s for account %s: %v\n", region, account.ID, err)
	} else {
		ips = append(ips, elbv2IPs...)
	}

	// Get Classic Load Balancers (ELBv1)
	elbv1IPs, err := s.getClassicELBPublicIPs(sess, account, region)
	if err != nil {
		fmt.Printf("[WARNING] Failed to get Classic ELB IPs in region %s for account %s: %v\n", region, account.ID, err)
	} else {
		ips = append(ips, elbv1IPs...)
	}

	return ips, nil
}

func (s *AWSService) getELBv2PublicIPs(sess *session.Session, account models.Account, region string) ([]models.PublicIP, error) {
	elbClient := elbv2.New(sess)
	var ips []models.PublicIP

	result, err := elbClient.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}

	for _, lb := range result.LoadBalancers {
		if lb.Scheme != nil && *lb.Scheme == "internet-facing" && lb.DNSName != nil {
			lbType := "ALB"
			if lb.Type != nil {
				lbType = string(*lb.Type)
			}

			// For NLB with static IPs, try to get them directly
			if lb.Type != nil && *lb.Type == "network" {
				for _, az := range lb.AvailabilityZones {
					if az.LoadBalancerAddresses != nil {
						for _, addr := range az.LoadBalancerAddresses {
							if addr.IpAddress != nil && *addr.IpAddress != "" {
								ips = append(ips, models.PublicIP{
									IPAddress:    *addr.IpAddress,
									AccountID:    account.ID,
									AccountName:  account.Name,
									Region:       region,
									ResourceType: "NLB",
									ResourceID:   *lb.LoadBalancerArn,
									ResourceName: *lb.LoadBalancerName,
									State:        string(*lb.State.Code),
								})
							}
						}
					}
				}
			}

			// For ALB and NLB without static IPs, resolve DNS name
			if len(ips) == 0 || (lb.Type != nil && *lb.Type != "network") {
				resolvedIPs, err := s.resolveDNSToIPs(*lb.DNSName)
				if err != nil {
					fmt.Printf("[WARNING] Failed to resolve DNS %s: %v\n", *lb.DNSName, err)
					continue
				}

				for _, ip := range resolvedIPs {
					ips = append(ips, models.PublicIP{
						IPAddress:    ip,
						AccountID:    account.ID,
						AccountName:  account.Name,
						Region:       region,
						ResourceType: lbType,
						ResourceID:   *lb.LoadBalancerArn,
						ResourceName: *lb.LoadBalancerName,
						State:        string(*lb.State.Code),
					})
				}
			}
		}
	}

	return ips, nil
}

func (s *AWSService) getClassicELBPublicIPs(sess *session.Session, account models.Account, region string) ([]models.PublicIP, error) {
	// Import classic ELB package
	elbClient := elb.New(sess)
	var ips []models.PublicIP

	result, err := elbClient.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}

	for _, lb := range result.LoadBalancerDescriptions {
		if lb.Scheme != nil && *lb.Scheme == "internet-facing" && lb.DNSName != nil {
			// Resolve DNS name to get IP addresses
			resolvedIPs, err := s.resolveDNSToIPs(*lb.DNSName)
			if err != nil {
				fmt.Printf("[WARNING] Failed to resolve DNS %s: %v\n", *lb.DNSName, err)
				continue
			}

			for _, ip := range resolvedIPs {
				ips = append(ips, models.PublicIP{
					IPAddress:    ip,
					AccountID:    account.ID,
					AccountName:  account.Name,
					Region:       region,
					ResourceType: "ELB",
					ResourceID:   *lb.LoadBalancerName, // Classic ELB doesn't have ARN
					ResourceName: *lb.LoadBalancerName,
					State:        "active", // Classic ELB doesn't have detailed state
				})
			}
		}
	}

	return ips, nil
}

func (s *AWSService) resolveDNSToIPs(dnsName string) ([]string, error) {
	ips, err := net.LookupIP(dnsName)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			result = append(result, ipv4.String())
		}
	}

	return result, nil
}

func (s *AWSService) getNATPublicIPs(sess *session.Session, account models.Account, region string) ([]models.PublicIP, error) {
	ec2Client := ec2.New(sess)
	var ips []models.PublicIP

	result, err := ec2Client.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{})
	if err != nil {
		return nil, err
	}

	for _, nat := range result.NatGateways {
		for _, addr := range nat.NatGatewayAddresses {
			if addr.PublicIp != nil && *addr.PublicIp != "" {
				natName := ""
				for _, tag := range nat.Tags {
					if *tag.Key == "Name" {
						natName = *tag.Value
						break
					}
				}

				ips = append(ips, models.PublicIP{
					IPAddress:    *addr.PublicIp,
					AccountID:    account.ID,
					AccountName:  account.Name,
					Region:       region,
					ResourceType: "NAT",
					ResourceID:   *nat.NatGatewayId,
					ResourceName: natName,
					State:        string(*nat.State),
				})
			}
		}
	}

	return ips, nil
}

// generatePassword generates a secure random password
func (s *AWSService) generatePassword() string {
	// AWS password requirements: 8-128 characters, at least 3 of 4 character types
	// (uppercase, lowercase, numbers, symbols)
	const (
		length    = 16
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		numbers   = "0123456789"
		symbols   = "!@#$%^&*"
	)

	allChars := uppercase + lowercase + numbers + symbols
	password := make([]byte, length)

	// Ensure at least one character from each type
	charSets := []string{uppercase, lowercase, numbers, symbols}
	for i, charset := range charSets {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[n.Int64()]
	}

	// Fill the rest with random characters
	for i := len(charSets); i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		password[i] = allChars[n.Int64()]
	}

	// Shuffle the password to avoid predictable patterns
	for i := length - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password)
}
