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
	var accounts []models.Account
	var nextToken *string

	// Paginate through all accounts
	for {
		input := &organizations.ListAccountsInput{}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		result, err := orgClient.ListAccounts(input)
		if err != nil {
			return nil, fmt.Errorf("failed to list organization accounts: %w", err)
		}

		// Process accounts from this page
		for _, account := range result.Accounts {
			// Test if we can access this account
			accessible := s.testAccountAccess(*account.Id)

			accounts = append(accounts, models.Account{
				ID:         *account.Id,
				Name:       *account.Name,
				Accessible: accessible,
			})
		}

		// Check if there are more pages
		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken

		fmt.Printf("[INFO] Retrieved %d accounts so far, continuing pagination...\n", len(accounts))
	}

	fmt.Printf("[INFO] Successfully discovered %d total accounts in organization\n", len(accounts))

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
	var users []models.User
	var marker *string

	// Paginate through all users
	for {
		input := &iam.ListUsersInput{}
		if marker != nil {
			input.Marker = marker
		}

		result, err := iamClient.ListUsers(input)
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
		}

		// Process users from this page
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
		accessKeys, err := s.getAccessKeysForUser(iamClient, user.UserName)
		if err != nil {
			fmt.Printf("[WARNING] Failed to get access keys for user %s: %v\n", *user.UserName, err)
			accessKeys = []models.AccessKey{}
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

		// Check if there are more pages
		if !*result.IsTruncated {
			break
		}
		marker = result.Marker
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
	accessKeys, err := s.getAccessKeysForUser(iamClient, user.UserName)
	if err != nil {
		fmt.Printf("[WARNING] Failed to get access keys for user %s: %v\n", *user.UserName, err)
		accessKeys = []models.AccessKey{}
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

// InvalidateSecurityGroupsCache invalidates the security groups cache
func (s *AWSService) InvalidateSecurityGroupsCache() {
	s.cache.Delete("security-groups")
	// Also invalidate all account-specific security group caches
	s.cache.DeletePattern("security-groups:")
	// Also invalidate all individual security group caches
	s.cache.DeletePattern("security-group:")
}

// InvalidateAccountSecurityGroupsCache invalidates security groups cache for a specific account
func (s *AWSService) InvalidateAccountSecurityGroupsCache(accountID string) {
	s.cache.Delete(fmt.Sprintf("security-groups:%s", accountID))
	// Also invalidate individual security group caches for this account
	s.cache.DeletePattern(fmt.Sprintf("security-group:%s:", accountID))
	// Also invalidate the global cache since it contains this account's data
	s.cache.Delete("security-groups")
}

// InvalidateSecurityGroupCache invalidates cache for a specific security group
func (s *AWSService) InvalidateSecurityGroupCache(accountID, region, groupID string) {
	s.cache.Delete(fmt.Sprintf("security-group:%s:%s:%s", accountID, region, groupID))
	// Also invalidate broader caches that contain this security group
	s.cache.Delete(fmt.Sprintf("security-groups:%s", accountID))
	s.cache.Delete("security-groups")
}

func (s *AWSService) ListSecurityGroups() ([]models.SecurityGroup, error) {
	const cacheKey = "security-groups"

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if sgs, ok := cached.([]models.SecurityGroup); ok {
			return sgs, nil
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
		return []models.SecurityGroup{}, nil
	}

	// Channel to collect results from goroutines
	type accountResult struct {
		sgs       []models.SecurityGroup
		err       error
		accountID string
	}

	resultChan := make(chan accountResult, len(accessibleAccounts))
	var wg sync.WaitGroup

	// Process each account in parallel
	for _, account := range accessibleAccounts {
		wg.Add(1)
		go func(acc models.Account) {
			defer wg.Done()

			sgs, err := s.getSecurityGroupsForAccount(acc)
			resultChan <- accountResult{
				sgs:       sgs,
				err:       err,
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
	var allSGs []models.SecurityGroup
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("[WARNING] Failed to get security groups for account %s: %v\n", result.accountID, result.err)
			continue
		}
		allSGs = append(allSGs, result.sgs...)
	}

	// Cache the result
	s.cache.Set(cacheKey, allSGs, s.cacheTTL)

	return allSGs, nil
}

func (s *AWSService) ListSecurityGroupsByAccount(accountID string) ([]models.SecurityGroup, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("security-groups:%s", accountID)
	if cached, found := s.cache.Get(cacheKey); found {
		if sgs, ok := cached.([]models.SecurityGroup); ok {
			return sgs, nil
		}
	}

	// Get account info
	accounts, err := s.ListAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %v", err)
	}

	var targetAccount models.Account
	found := false
	for _, account := range accounts {
		if account.ID == accountID {
			targetAccount = account
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("account %s not found", accountID)
	}

	if !targetAccount.Accessible {
		return nil, fmt.Errorf("cannot access account %s", accountID)
	}

	// Get security groups for the account
	sgs, err := s.getSecurityGroupsForAccount(targetAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get security groups for account %s: %v", accountID, err)
	}

	// Cache the result
	s.cache.Set(cacheKey, sgs, s.cacheTTL)

	return sgs, nil
}

func (s *AWSService) GetSecurityGroup(accountID, region, groupID string) (*models.SecurityGroup, error) {
	// Check individual cache first
	cacheKey := fmt.Sprintf("security-group:%s:%s:%s", accountID, region, groupID)
	if cached, found := s.cache.Get(cacheKey); found {
		if sg, ok := cached.(models.SecurityGroup); ok {
			return &sg, nil
		}
	}

	// Fetch directly from AWS using specific security group ID
	sg, err := s.fetchSecurityGroupDirect(accountID, region, groupID)
	if err != nil {
		return nil, err
	}

	// Cache the individual security group
	s.cache.Set(cacheKey, *sg, s.cacheTTL)

	return sg, nil
}

func (s *AWSService) fetchSecurityGroupDirect(accountID, region, groupID string) (*models.SecurityGroup, error) {
	// Get session for the account
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	// We'll use the accountID as the account name if we can't get it from cache
	// This avoids the expensive ListAccounts() call
	accountName := accountID

	// Try to get account name from cache first
	if cached, found := s.cache.Get("accounts"); found {
		if accounts, ok := cached.([]models.Account); ok {
			for _, acc := range accounts {
				if acc.ID == accountID {
					accountName = acc.Name
					break
				}
			}
		}
	}

	// Create EC2 client for the specific region
	ec2Client := ec2.New(sess.Copy(&aws.Config{Region: aws.String(region)}))

	// Fetch the specific security group by ID
	result, err := ec2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{aws.String(groupID)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe security group %s: %v", groupID, err)
	}

	if len(result.SecurityGroups) == 0 {
		return nil, fmt.Errorf("security group %s not found in account %s, region %s", groupID, accountID, region)
	}

	sg := result.SecurityGroups[0]

	// Convert ingress rules
	var ingressRules []models.SecurityGroupRule
	for _, rule := range sg.IpPermissions {
		rules := s.convertEC2RuleToModel(rule)
		ingressRules = append(ingressRules, rules...)
	}

	// Convert egress rules
	var egressRules []models.SecurityGroupRule
	for _, rule := range sg.IpPermissionsEgress {
		rules := s.convertEC2RuleToModel(rule)
		egressRules = append(egressRules, rules...)
	}

	// Check for open ports to internet
	hasOpenPorts, openPortsInfo := s.checkForOpenPorts(ingressRules)

	// Check usage of security group
	usageInfo, err := s.checkSecurityGroupUsage(ec2Client, *sg.GroupId)
	if err != nil {
		fmt.Printf("[WARNING] Failed to check usage for security group %s: %v\n", *sg.GroupId, err)
		usageInfo = models.SecurityGroupUsage{}
	}
	isUnused := usageInfo.TotalAttachments == 0

	securityGroup := &models.SecurityGroup{
		GroupID:       *sg.GroupId,
		GroupName:     *sg.GroupName,
		Description:   *sg.Description,
		AccountID:     accountID,
		AccountName:   accountName,
		Region:        region,
		VpcID:         aws.StringValue(sg.VpcId),
		IsDefault:     *sg.GroupName == "default",
		IngressRules:  ingressRules,
		EgressRules:   egressRules,
		HasOpenPorts:  hasOpenPorts,
		OpenPortsInfo: openPortsInfo,
		IsUnused:      isUnused,
		UsageInfo:     usageInfo,
	}

	return securityGroup, nil
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
					ResourceType: "CLB",
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

func (s *AWSService) getSecurityGroupsForAccount(account models.Account) ([]models.SecurityGroup, error) {
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
		return []models.SecurityGroup{}, nil
	}

	// Channel to collect results from region goroutines
	type regionResult struct {
		sgs        []models.SecurityGroup
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
			regionSGs, err := s.getSecurityGroupsForRegion(regionSess, account, regionName)
			if err != nil {
				fmt.Printf("[WARNING] Failed to get security groups in region %s for account %s: %v\n", regionName, account.ID, err)
				regionSGs = []models.SecurityGroup{}
			}

			resultChan <- regionResult{
				sgs:        regionSGs,
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
	var allSGs []models.SecurityGroup
	for result := range resultChan {
		allSGs = append(allSGs, result.sgs...)
	}

	return allSGs, nil
}

func (s *AWSService) getSecurityGroupsForRegion(sess *session.Session, account models.Account, region string) ([]models.SecurityGroup, error) {
	ec2Client := ec2.New(sess)
	var sgs []models.SecurityGroup

	result, err := ec2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, err
	}

	for _, sg := range result.SecurityGroups {
		// Convert ingress rules
		var ingressRules []models.SecurityGroupRule
		for _, rule := range sg.IpPermissions {
			rules := s.convertEC2RuleToModel(rule)
			ingressRules = append(ingressRules, rules...)
		}

		// Convert egress rules
		var egressRules []models.SecurityGroupRule
		for _, rule := range sg.IpPermissionsEgress {
			rules := s.convertEC2RuleToModel(rule)
			egressRules = append(egressRules, rules...)
		}

		// Check for open ports to internet
		hasOpenPorts, openPortsInfo := s.checkForOpenPorts(ingressRules)

		// Check usage of security group
		usageInfo, err := s.checkSecurityGroupUsage(ec2Client, *sg.GroupId)
		if err != nil {
			fmt.Printf("[WARNING] Failed to check usage for security group %s: %v\n", *sg.GroupId, err)
			usageInfo = models.SecurityGroupUsage{}
		}
		isUnused := usageInfo.TotalAttachments == 0

		securityGroup := models.SecurityGroup{
			GroupID:       *sg.GroupId,
			GroupName:     *sg.GroupName,
			Description:   *sg.Description,
			AccountID:     account.ID,
			AccountName:   account.Name,
			Region:        region,
			VpcID:         aws.StringValue(sg.VpcId),
			IsDefault:     *sg.GroupName == "default",
			IngressRules:  ingressRules,
			EgressRules:   egressRules,
			HasOpenPorts:  hasOpenPorts,
			OpenPortsInfo: openPortsInfo,
			IsUnused:      isUnused,
			UsageInfo:     usageInfo,
		}

		sgs = append(sgs, securityGroup)
	}

	return sgs, nil
}

func (s *AWSService) convertEC2RuleToModel(rule *ec2.IpPermission) []models.SecurityGroupRule {
	var rules []models.SecurityGroupRule

	protocol := aws.StringValue(rule.IpProtocol)
	fromPort := aws.Int64Value(rule.FromPort)
	toPort := aws.Int64Value(rule.ToPort)

	// Handle IPv4 CIDR blocks
	for _, ipRange := range rule.IpRanges {
		rules = append(rules, models.SecurityGroupRule{
			IpProtocol: protocol,
			FromPort:   fromPort,
			ToPort:     toPort,
			CidrIPv4:   aws.StringValue(ipRange.CidrIp),
		})
	}

	// Handle IPv6 CIDR blocks
	for _, ipv6Range := range rule.Ipv6Ranges {
		rules = append(rules, models.SecurityGroupRule{
			IpProtocol: protocol,
			FromPort:   fromPort,
			ToPort:     toPort,
			CidrIPv6:   aws.StringValue(ipv6Range.CidrIpv6),
		})
	}

	// Handle security group references
	for _, sgRef := range rule.UserIdGroupPairs {
		rules = append(rules, models.SecurityGroupRule{
			IpProtocol: protocol,
			FromPort:   fromPort,
			ToPort:     toPort,
			GroupID:    aws.StringValue(sgRef.GroupId),
			GroupOwner: aws.StringValue(sgRef.UserId),
		})
	}

	// If no specific ranges are defined, create a rule without CIDR
	if len(rule.IpRanges) == 0 && len(rule.Ipv6Ranges) == 0 && len(rule.UserIdGroupPairs) == 0 {
		rules = append(rules, models.SecurityGroupRule{
			IpProtocol: protocol,
			FromPort:   fromPort,
			ToPort:     toPort,
		})
	}

	return rules
}

func (s *AWSService) checkForOpenPorts(ingressRules []models.SecurityGroupRule) (bool, []models.OpenPortInfo) {
	var openPortsInfo []models.OpenPortInfo
	hasOpenPorts := false

	for _, rule := range ingressRules {
		isOpenToInternet := false
		source := ""

		// Check if rule allows access from anywhere on the internet
		if rule.CidrIPv4 == "0.0.0.0/0" {
			isOpenToInternet = true
			source = "0.0.0.0/0 (IPv4 Internet)"
		} else if rule.CidrIPv6 == "::/0" {
			isOpenToInternet = true
			source = "::/0 (IPv6 Internet)"
		}

		if isOpenToInternet {
			hasOpenPorts = true

			portRange := ""
			if rule.IpProtocol == "-1" {
				portRange = "All ports"
			} else if rule.FromPort == rule.ToPort {
				portRange = fmt.Sprintf("%d", rule.FromPort)
			} else {
				portRange = fmt.Sprintf("%d-%d", rule.FromPort, rule.ToPort)
			}

			description := ""
			switch rule.IpProtocol {
			case "tcp":
				description = "TCP traffic"
			case "udp":
				description = "UDP traffic"
			case "icmp":
				description = "ICMP traffic"
			case "-1":
				description = "All traffic"
			default:
				description = fmt.Sprintf("Protocol %s", rule.IpProtocol)
			}

			openPortsInfo = append(openPortsInfo, models.OpenPortInfo{
				Protocol:    rule.IpProtocol,
				PortRange:   portRange,
				Source:      source,
				Description: description,
			})
		}
	}

	return hasOpenPorts, openPortsInfo
}

// checkSecurityGroupUsage checks if a security group is being used by any resources
func (s *AWSService) checkSecurityGroupUsage(ec2Client *ec2.EC2, groupID string) (models.SecurityGroupUsage, error) {
	usage := models.SecurityGroupUsage{
		AttachedToInstances:       []string{},
		AttachedToNetworkInterfaces: []string{},
		AttachedToLoadBalancers:   []string{},
		ReferencedBySecurityGroups: []string{},
		TotalAttachments:          0,
	}

	// Check EC2 instances
	instancesResult, err := ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance.group-id"),
				Values: []*string{aws.String(groupID)},
			},
		},
	})
	if err != nil {
		return usage, fmt.Errorf("failed to describe instances: %v", err)
	}

	for _, reservation := range instancesResult.Reservations {
		for _, instance := range reservation.Instances {
			if instance.InstanceId != nil {
				usage.AttachedToInstances = append(usage.AttachedToInstances, *instance.InstanceId)
				usage.TotalAttachments++
			}
		}
	}

	// Check network interfaces
	eniResult, err := ec2Client.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-id"),
				Values: []*string{aws.String(groupID)},
			},
		},
	})
	if err != nil {
		return usage, fmt.Errorf("failed to describe network interfaces: %v", err)
	}

	for _, eni := range eniResult.NetworkInterfaces {
		if eni.NetworkInterfaceId != nil {
			usage.AttachedToNetworkInterfaces = append(usage.AttachedToNetworkInterfaces, *eni.NetworkInterfaceId)
			usage.TotalAttachments++
		}
	}

	// Check load balancers (both Classic and Application/Network Load Balancers)
	// Note: For ALB/NLB, we need to use ELBv2 API, but for simplicity, we'll check through ENIs
	// Load balancers are typically attached through ENIs, so they should be caught above

	// Check if this security group is referenced by other security groups
	allSGsResult, err := ec2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return usage, fmt.Errorf("failed to describe all security groups: %v", err)
	}

	for _, sg := range allSGsResult.SecurityGroups {
		if *sg.GroupId == groupID {
			continue // Skip self
		}

		// Check ingress rules
		for _, rule := range sg.IpPermissions {
			for _, userIdGroupPair := range rule.UserIdGroupPairs {
				if userIdGroupPair.GroupId != nil && *userIdGroupPair.GroupId == groupID {
					usage.ReferencedBySecurityGroups = append(usage.ReferencedBySecurityGroups, *sg.GroupId)
					usage.TotalAttachments++
					break
				}
			}
		}

		// Check egress rules
		for _, rule := range sg.IpPermissionsEgress {
			for _, userIdGroupPair := range rule.UserIdGroupPairs {
				if userIdGroupPair.GroupId != nil && *userIdGroupPair.GroupId == groupID {
					// Only add if not already added from ingress rules
					found := false
					for _, existingRef := range usage.ReferencedBySecurityGroups {
						if existingRef == *sg.GroupId {
							found = true
							break
						}
					}
					if !found {
						usage.ReferencedBySecurityGroups = append(usage.ReferencedBySecurityGroups, *sg.GroupId)
						usage.TotalAttachments++
					}
					break
				}
			}
		}
	}

	return usage, nil
}

// DeleteSecurityGroup deletes a security group
func (s *AWSService) DeleteSecurityGroup(accountID, region, groupID string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	// Create session for the specific region
	regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
	ec2Client := ec2.New(regionSess)

	// First, check if the security group exists and get its details
	sgResult, err := ec2Client.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		GroupIds: []*string{aws.String(groupID)},
	})
	if err != nil {
		return fmt.Errorf("failed to describe security group %s: %v", groupID, err)
	}

	if len(sgResult.SecurityGroups) == 0 {
		return fmt.Errorf("security group %s not found", groupID)
	}

	sg := sgResult.SecurityGroups[0]

	// Prevent deletion of default security groups
	if *sg.GroupName == "default" {
		return fmt.Errorf("cannot delete default security group")
	}

	// Check if the security group is in use before attempting deletion
	usage, err := s.checkSecurityGroupUsage(ec2Client, groupID)
	if err != nil {
		// Log warning but proceed with deletion attempt
		fmt.Printf("[WARNING] Failed to check usage for security group %s: %v\n", groupID, err)
	} else if usage.TotalAttachments > 0 {
		return fmt.Errorf("security group %s is still in use (attached to %d resources)", groupID, usage.TotalAttachments)
	}

	// Attempt to delete the security group
	_, err = ec2Client.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(groupID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete security group %s: %v", groupID, err)
	}

	// Invalidate the specific security group cache and related caches
	s.InvalidateSecurityGroupCache(accountID, region, groupID)

	return nil
}

// getAccessKeysForUser retrieves access keys for a user with last used information
func (s *AWSService) getAccessKeysForUser(iamClient *iam.IAM, userName *string) ([]models.AccessKey, error) {
	keysResult, err := iamClient.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: userName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list access keys: %v", err)
	}

	var accessKeys []models.AccessKey
	for _, key := range keysResult.AccessKeyMetadata {
		accessKey := models.AccessKey{
			AccessKeyID: *key.AccessKeyId,
			Status:      *key.Status,
			CreateDate:  *key.CreateDate,
		}

		// Get last used information
		lastUsedResult, err := iamClient.GetAccessKeyLastUsed(&iam.GetAccessKeyLastUsedInput{
			AccessKeyId: key.AccessKeyId,
		})
		if err != nil {
			fmt.Printf("[WARNING] Failed to get last used info for key %s: %v\n", *key.AccessKeyId, err)
		} else if lastUsedResult.AccessKeyLastUsed != nil && lastUsedResult.AccessKeyLastUsed.LastUsedDate != nil {
			accessKey.LastUsedDate = lastUsedResult.AccessKeyLastUsed.LastUsedDate
			if lastUsedResult.AccessKeyLastUsed.ServiceName != nil {
				accessKey.LastUsedService = *lastUsedResult.AccessKeyLastUsed.ServiceName
			}
			if lastUsedResult.AccessKeyLastUsed.Region != nil {
				accessKey.LastUsedRegion = *lastUsedResult.AccessKeyLastUsed.Region
			}
		}

		accessKeys = append(accessKeys, accessKey)
	}

	return accessKeys, nil
}

