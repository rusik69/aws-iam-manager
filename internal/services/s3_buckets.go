package services

import (
	"fmt"
	"sync"

	"github.com/rusik69/aws-iam-manager/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// ============================================================================
// S3 BUCKET MANAGEMENT
// ============================================================================

func (s *AWSService) ListS3Buckets() ([]models.S3Bucket, error) {
	const cacheKey = "s3-buckets"

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if buckets, ok := cached.([]models.S3Bucket); ok {
			return buckets, nil
		}
	}

	accounts, err := s.ListAccounts()
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %v", err)
	}

	// Filter accessible accounts
	var accessibleAccounts []models.Account
	for _, account := range accounts {
		if account.Accessible {
			accessibleAccounts = append(accessibleAccounts, account)
		}
	}

	if len(accessibleAccounts) == 0 {
		return []models.S3Bucket{}, nil
	}

	// Channel to collect results from goroutines
	type accountResult struct {
		buckets   []models.S3Bucket
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

			buckets, err := s.getS3BucketsForAccount(acc)
			resultChan <- accountResult{
				buckets:   buckets,
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
	var allBuckets []models.S3Bucket
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("[WARNING] Failed to get S3 buckets for account %s: %v\n", result.accountID, result.err)
			continue
		}
		allBuckets = append(allBuckets, result.buckets...)
	}

	// Cache the result
	s.cache.Set(cacheKey, allBuckets, s.cacheTTL)

	return allBuckets, nil
}

// getS3BucketsForAccount returns all S3 buckets for a specific account
func (s *AWSService) getS3BucketsForAccount(account models.Account) ([]models.S3Bucket, error) {
	sess, err := s.getSessionForAccount(account.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot access account %s: %w", account.ID, err)
	}

	// S3 ListBuckets is global, but we need to use a specific region
	// Use us-east-1 as the default region for listing buckets
	s3Client := s3.New(sess.Copy(&aws.Config{Region: aws.String("us-east-1")}))

	// List all buckets
	bucketsResult, err := s3Client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %v", err)
	}

	if len(bucketsResult.Buckets) == 0 {
		return []models.S3Bucket{}, nil
	}

	// Channel to collect bucket details from goroutines
	type bucketResult struct {
		bucket models.S3Bucket
		err    error
	}

	resultChan := make(chan bucketResult, len(bucketsResult.Buckets))
	var wg sync.WaitGroup

	// Get detailed information for each bucket in parallel
	for _, bucket := range bucketsResult.Buckets {
		wg.Add(1)
		go func(bkt *s3.Bucket) {
			defer wg.Done()

			bucketDetail, err := s.getS3BucketDetails(sess, account, bkt)
			if err != nil {
				fmt.Printf("[WARNING] Failed to get details for bucket %s: %v\n", *bkt.Name, err)
				resultChan <- bucketResult{err: err}
				return
			}

			resultChan <- bucketResult{bucket: bucketDetail}
		}(bucket)
	}

	// Wait for all goroutines to complete and close channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var buckets []models.S3Bucket
	for result := range resultChan {
		if result.err == nil {
			buckets = append(buckets, result.bucket)
		}
	}

	return buckets, nil
}

// getS3BucketDetails gets detailed information about a specific S3 bucket
func (s *AWSService) getS3BucketDetails(sess *session.Session, account models.Account, bucket *s3.Bucket) (models.S3Bucket, error) {
	bucketName := *bucket.Name

	// Get bucket location
	s3Client := s3.New(sess.Copy(&aws.Config{Region: aws.String("us-east-1")}))
	locationResult, err := s3Client.GetBucketLocation(&s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return models.S3Bucket{}, fmt.Errorf("failed to get bucket location: %v", err)
	}

	region := "us-east-1"
	if locationResult.LocationConstraint != nil && *locationResult.LocationConstraint != "" {
		region = *locationResult.LocationConstraint
	}

	// Create a regional S3 client
	regionalS3Client := s3.New(sess.Copy(&aws.Config{Region: aws.String(region)}))

	bucketModel := models.S3Bucket{
		Name:         bucketName,
		AccountID:    account.ID,
		AccountName:  account.Name,
		Region:       region,
		CreationDate: *bucket.CreationDate,
	}

	// Get versioning status
	versioningResult, err := regionalS3Client.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil && versioningResult.Status != nil {
		bucketModel.Versioning = *versioningResult.Status
	}

	// Get encryption configuration
	encryptionResult, err := regionalS3Client.GetBucketEncryption(&s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil && encryptionResult.ServerSideEncryptionConfiguration != nil {
		bucketModel.Encrypted = true
	}

	// Get public access block configuration
	publicAccessResult, err := regionalS3Client.GetPublicAccessBlock(&s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil && publicAccessResult.PublicAccessBlockConfiguration != nil {
		config := publicAccessResult.PublicAccessBlockConfiguration
		bucketModel.PublicAccessBlocked = config.BlockPublicAcls != nil && *config.BlockPublicAcls &&
			config.BlockPublicPolicy != nil && *config.BlockPublicPolicy &&
			config.IgnorePublicAcls != nil && *config.IgnorePublicAcls &&
			config.RestrictPublicBuckets != nil && *config.RestrictPublicBuckets
	}

	// Check if bucket is public by checking ACL and policy
	bucketModel.IsPublic = !bucketModel.PublicAccessBlocked

	// Get bucket tagging
	tagsResult, err := regionalS3Client.GetBucketTagging(&s3.GetBucketTaggingInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil && tagsResult.TagSet != nil {
		for _, tag := range tagsResult.TagSet {
			if tag.Key != nil && tag.Value != nil {
				bucketModel.Tags = append(bucketModel.Tags, models.Tag{
					Key:   *tag.Key,
					Value: *tag.Value,
				})
			}
		}
	}

	// Check for lifecycle policy
	_, err = regionalS3Client.GetBucketLifecycleConfiguration(&s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	bucketModel.HasLifecyclePolicy = err == nil

	// Check for logging
	loggingResult, err := regionalS3Client.GetBucketLogging(&s3.GetBucketLoggingInput{
		Bucket: aws.String(bucketName),
	})
	bucketModel.HasLogging = err == nil && loggingResult.LoggingEnabled != nil

	return bucketModel, nil
}

// ListS3BucketsByAccount returns all S3 buckets for a specific account
func (s *AWSService) ListS3BucketsByAccount(accountID string) ([]models.S3Bucket, error) {
	cacheKey := fmt.Sprintf("s3-buckets:%s", accountID)

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if buckets, ok := cached.([]models.S3Bucket); ok {
			return buckets, nil
		}
	}

	// Get all buckets and filter by account
	allBuckets, err := s.ListS3Buckets()
	if err != nil {
		return nil, err
	}

	var accountBuckets []models.S3Bucket
	for _, bucket := range allBuckets {
		if bucket.AccountID == accountID {
			accountBuckets = append(accountBuckets, bucket)
		}
	}

	// Cache the result
	s.cache.Set(cacheKey, accountBuckets, s.cacheTTL)

	return accountBuckets, nil
}

// DeleteS3Bucket deletes an S3 bucket (bucket must be empty)
func (s *AWSService) DeleteS3Bucket(accountID, region, bucketName string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
	s3Client := s3.New(regionSess)

	// Try to delete the bucket
	_, err = s3Client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket %s: %v (bucket must be empty)", bucketName, err)
	}

	// Update cache - remove the deleted bucket
	s.updateS3BucketCache(accountID, bucketName)

	return nil
}

// updateS3BucketCache removes a deleted bucket from the cache
func (s *AWSService) updateS3BucketCache(accountID, bucketName string) {
	// Update account-specific cache
	cacheKey := fmt.Sprintf("s3-buckets:%s", accountID)
	if cached, found := s.cache.Get(cacheKey); found {
		if buckets, ok := cached.([]models.S3Bucket); ok {
			var updated []models.S3Bucket
			for _, bucket := range buckets {
				if bucket.Name != bucketName {
					updated = append(updated, bucket)
				}
			}
			s.cache.Set(cacheKey, updated, s.cacheTTL)
		}
	}

	// Update global cache
	if cached, found := s.cache.Get("s3-buckets"); found {
		if buckets, ok := cached.([]models.S3Bucket); ok {
			var updated []models.S3Bucket
			for _, bucket := range buckets {
				if bucket.Name != bucketName {
					updated = append(updated, bucket)
				}
			}
			s.cache.Set("s3-buckets", updated, s.cacheTTL)
		}
	}
}

// InvalidateS3BucketsCache invalidates the S3 buckets cache
func (s *AWSService) InvalidateS3BucketsCache() {
	s.cache.Delete("s3-buckets")
	s.cache.DeletePattern("s3-buckets:")
}

// ============================================================================
