package services

import (
	"fmt"
	"sync"

	"github.com/rusik69/aws-iam-manager/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// ============================================================================
// EC2 INSTANCE MANAGEMENT
// ============================================================================

func (s *AWSService) ListEC2Instances() ([]models.EC2Instance, error) {
	const cacheKey = "ec2-instances"

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if instances, ok := cached.([]models.EC2Instance); ok {
			return instances, nil
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
		return []models.EC2Instance{}, nil
	}

	// Channel to collect results from goroutines
	type accountResult struct {
		instances []models.EC2Instance
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

			instances, err := s.getEC2InstancesForAccount(acc)
			resultChan <- accountResult{
				instances: instances,
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
	var allInstances []models.EC2Instance
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("[WARNING] Failed to get EC2 instances for account %s: %v\n", result.accountID, result.err)
			continue
		}
		allInstances = append(allInstances, result.instances...)
	}

	// Cache the result
	s.cache.Set(cacheKey, allInstances, s.cacheTTL)

	return allInstances, nil
}

// getEC2InstancesForAccount returns all EC2 instances for a specific account across all regions
func (s *AWSService) getEC2InstancesForAccount(account models.Account) ([]models.EC2Instance, error) {
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
		return []models.EC2Instance{}, nil
	}

	// Channel to collect results from region goroutines
	type regionResult struct {
		instances  []models.EC2Instance
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
			regionInstances, err := s.getEC2InstancesForRegion(regionSess, account, regionName)
			if err != nil {
				fmt.Printf("[WARNING] Failed to get instances in region %s for account %s: %v\n", regionName, account.ID, err)
				regionInstances = []models.EC2Instance{}
			}

			resultChan <- regionResult{
				instances:  regionInstances,
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
	var allInstances []models.EC2Instance
	for result := range resultChan {
		allInstances = append(allInstances, result.instances...)
	}

	return allInstances, nil
}

// getEC2InstancesForRegion returns all EC2 instances for a specific region
func (s *AWSService) getEC2InstancesForRegion(sess *session.Session, account models.Account, region string) ([]models.EC2Instance, error) {
	ec2Client := ec2.New(sess)
	var instances []models.EC2Instance
	var nextToken *string

	// Paginate through all instances (handle large instance counts)
	for {
		input := &ec2.DescribeInstancesInput{}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		result, err := ec2Client.DescribeInstances(input)
		if err != nil {
			return nil, err
		}

		// Process instances from this page
		for _, reservation := range result.Reservations {
			for _, instance := range reservation.Instances {
				// Extract instance name from tags
				instanceName := ""
				for _, tag := range instance.Tags {
					if *tag.Key == "Name" {
						instanceName = *tag.Value
						break
					}
				}

				// Convert tags
				var tags []models.Tag
				for _, tag := range instance.Tags {
					tags = append(tags, models.Tag{
						Key:   aws.StringValue(tag.Key),
						Value: aws.StringValue(tag.Value),
					})
				}

				ec2Instance := models.EC2Instance{
					InstanceID:   aws.StringValue(instance.InstanceId),
					Name:         instanceName,
					AccountID:    account.ID,
					AccountName:  account.Name,
					Region:       region,
					InstanceType: aws.StringValue(instance.InstanceType),
					State:        aws.StringValue(instance.State.Name),
					Tags:         tags,
				}

				if instance.LaunchTime != nil {
					ec2Instance.LaunchTime = *instance.LaunchTime
				}

				instances = append(instances, ec2Instance)
			}
		}

		// Check if there are more pages
		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	return instances, nil
}

// InvalidateEC2InstancesCache invalidates the EC2 instances cache
func (s *AWSService) InvalidateEC2InstancesCache() {
	s.cache.Delete("ec2-instances")
}

// updateEC2InstanceStateInCache updates a specific instance's state in the cache
func (s *AWSService) updateEC2InstanceStateInCache(instanceID, newState string) {
	const cacheKey = "ec2-instances"

	if cached, found := s.cache.Get(cacheKey); found {
		if instances, ok := cached.([]models.EC2Instance); ok {
			// Update the specific instance's state
			for i := range instances {
				if instances[i].InstanceID == instanceID {
					instances[i].State = newState
					break
				}
			}
			// Update the cache with the modified list
			s.cache.Set(cacheKey, instances, s.cacheTTL)
		}
	}
}

// StopEC2Instance stops an EC2 instance
func (s *AWSService) StopEC2Instance(accountID, region, instanceID string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	// Create session for the specific region
	regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
	ec2Client := ec2.New(regionSess)

	// Stop the instance
	_, err = ec2Client.StopInstances(&ec2.StopInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	})
	if err != nil {
		return fmt.Errorf("failed to stop instance %s: %v", instanceID, err)
	}

	// Update the instance state in cache instead of invalidating
	s.updateEC2InstanceStateInCache(instanceID, "stopping")

	return nil
}

// TerminateEC2Instance terminates an EC2 instance
func (s *AWSService) TerminateEC2Instance(accountID, region, instanceID string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	// Create session for the specific region
	regionSess := sess.Copy(&aws.Config{Region: aws.String(region)})
	ec2Client := ec2.New(regionSess)

	// Terminate the instance
	_, err = ec2Client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(instanceID)},
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instance %s: %v", instanceID, err)
	}

	// Update the instance state in cache instead of invalidating
	s.updateEC2InstanceStateInCache(instanceID, "terminating")

	return nil
}

// ============================================================================
// EBS VOLUME MANAGEMENT
// ============================================================================

// ListEBSVolumes returns all EBS volumes from all accessible accounts
