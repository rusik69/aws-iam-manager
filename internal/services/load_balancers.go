package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/rusik69/aws-iam-manager/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

// ============================================================================
// LOAD BALANCER MANAGEMENT
// ============================================================================

// ListAllLoadBalancers returns all load balancers from all accounts
func (s *AWSService) ListAllLoadBalancers() ([]models.LoadBalancer, error) {
	const cacheKey = "load-balancers"

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if lbs, ok := cached.([]models.LoadBalancer); ok {
			return lbs, nil
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
		return []models.LoadBalancer{}, nil
	}

	// Channel to collect results from goroutines
	type accountResult struct {
		lbs       []models.LoadBalancer
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

			lbs, err := s.ListLoadBalancersByAccount(acc.ID)
			resultChan <- accountResult{
				lbs:       lbs,
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
	var allLBs []models.LoadBalancer
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("[WARNING] Failed to get load balancers for account %s: %v\n", result.accountID, result.err)
			continue
		}
		allLBs = append(allLBs, result.lbs...)
	}

	// Cache the result
	s.cache.Set(cacheKey, allLBs, s.cacheTTL)

	return allLBs, nil
}

// ListLoadBalancersByAccount returns all load balancers for a specific account
func (s *AWSService) ListLoadBalancersByAccount(accountID string) ([]models.LoadBalancer, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("load-balancers:%s", accountID)
	if cached, found := s.cache.Get(cacheKey); found {
		if lbs, ok := cached.([]models.LoadBalancer); ok {
			return lbs, nil
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

	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	// Get all regions
	regions, err := s.getRegions(sess)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %v", err)
	}

	// Channel to collect results from region goroutines
	type regionResult struct {
		lbs        []models.LoadBalancer
		err        error
		regionName string
	}

	resultChan := make(chan regionResult, len(regions))
	var wg sync.WaitGroup

	// Process each region in parallel
	for _, region := range regions {
		wg.Add(1)
		go func(regionName string) {
			defer wg.Done()

			lbs, err := s.getLoadBalancersForRegion(sess, targetAccount, regionName)
			resultChan <- regionResult{
				lbs:        lbs,
				err:        err,
				regionName: regionName,
			}
		}(region)
	}

	// Wait for all goroutines to complete and close channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var allLBs []models.LoadBalancer
	for result := range resultChan {
		if result.err != nil {
			fmt.Printf("[WARNING] Failed to get load balancers in region %s for account %s: %v\n", result.regionName, accountID, result.err)
			continue
		}
		allLBs = append(allLBs, result.lbs...)
	}

	// Cache the result
	s.cache.Set(cacheKey, allLBs, s.cacheTTL)

	return allLBs, nil
}

// getLoadBalancersForRegion fetches all load balancers (ALB, NLB, and Classic) for a specific region
func (s *AWSService) getLoadBalancersForRegion(sess *session.Session, account models.Account, region string) ([]models.LoadBalancer, error) {
	var allLBs []models.LoadBalancer

	// Get ALB/NLB (ELBv2)
	elbv2Lbs, err := s.getELBv2LoadBalancers(sess, account, region)
	if err != nil {
		fmt.Printf("[WARNING] Failed to get ALB/NLB in region %s: %v\n", region, err)
	} else {
		allLBs = append(allLBs, elbv2Lbs...)
	}

	// Get Classic ELB
	classicLbs, err := s.getClassicLoadBalancers(sess, account, region)
	if err != nil {
		fmt.Printf("[WARNING] Failed to get Classic ELB in region %s: %v\n", region, err)
	} else {
		allLBs = append(allLBs, classicLbs...)
	}

	return allLBs, nil
}

// getELBv2LoadBalancers fetches ALB and NLB load balancers
func (s *AWSService) getELBv2LoadBalancers(sess *session.Session, account models.Account, region string) ([]models.LoadBalancer, error) {
	elbv2Client := elbv2.New(sess.Copy(&aws.Config{Region: aws.String(region)}))

	result, err := elbv2Client.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}

	var lbs []models.LoadBalancer

	for _, lb := range result.LoadBalancers {
		lbType := "application"
		if lb.Type != nil {
			lbType = *lb.Type
		}

		var createdTime *time.Time
		if lb.CreatedTime != nil {
			createdTime = lb.CreatedTime
		}

		loadBalancer := models.LoadBalancer{
			LoadBalancerArn:  aws.StringValue(lb.LoadBalancerArn),
			LoadBalancerName: aws.StringValue(lb.LoadBalancerName),
			DNSName:          aws.StringValue(lb.DNSName),
			Type:             lbType,
			Scheme:           aws.StringValue(lb.Scheme),
			State:            aws.StringValue(lb.State.Code),
			VPCID:            aws.StringValue(lb.VpcId),
			AccountID:        account.ID,
			AccountName:      account.Name,
			Region:           region,
			CreatedTime:      createdTime,
		}

		// Get target groups and listeners for this load balancer
		targetCount, healthyCount, listenerCount, isUnused := s.getELBv2Details(elbv2Client, *lb.LoadBalancerArn)
		loadBalancer.TargetCount = targetCount
		loadBalancer.HealthyTargetCount = healthyCount
		loadBalancer.ListenerCount = listenerCount
		loadBalancer.IsUnused = isUnused

		lbs = append(lbs, loadBalancer)
	}

	return lbs, nil
}

// getELBv2Details gets target group and listener details for an ALB/NLB
func (s *AWSService) getELBv2Details(client *elbv2.ELBV2, lbArn string) (targetCount, healthyCount, listenerCount int, isUnused bool) {
	// Get listeners
	listenersResult, err := client.DescribeListeners(&elbv2.DescribeListenersInput{
		LoadBalancerArn: aws.String(lbArn),
	})
	if err == nil {
		listenerCount = len(listenersResult.Listeners)
	}

	// Get target groups for this load balancer
	tgResult, err := client.DescribeTargetGroups(&elbv2.DescribeTargetGroupsInput{
		LoadBalancerArn: aws.String(lbArn),
	})
	if err != nil {
		// If no target groups, consider it unused
		isUnused = true
		return
	}

	if len(tgResult.TargetGroups) == 0 {
		isUnused = true
		return
	}

	// Check each target group for healthy targets
	for _, tg := range tgResult.TargetGroups {
		healthResult, err := client.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
			TargetGroupArn: tg.TargetGroupArn,
		})
		if err != nil {
			continue
		}

		targetCount += len(healthResult.TargetHealthDescriptions)
		for _, target := range healthResult.TargetHealthDescriptions {
			if target.TargetHealth != nil && target.TargetHealth.State != nil {
				if *target.TargetHealth.State == "healthy" {
					healthyCount++
				}
			}
		}
	}

	// Consider unused if no targets or no healthy targets
	isUnused = targetCount == 0 || healthyCount == 0

	return
}

// getClassicLoadBalancers fetches Classic ELB load balancers
func (s *AWSService) getClassicLoadBalancers(sess *session.Session, account models.Account, region string) ([]models.LoadBalancer, error) {
	elbClient := elb.New(sess.Copy(&aws.Config{Region: aws.String(region)}))

	result, err := elbClient.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}

	var lbs []models.LoadBalancer

	for _, lb := range result.LoadBalancerDescriptions {
		var createdTime *time.Time
		if lb.CreatedTime != nil {
			createdTime = lb.CreatedTime
		}

		loadBalancer := models.LoadBalancer{
			LoadBalancerName: aws.StringValue(lb.LoadBalancerName),
			DNSName:          aws.StringValue(lb.DNSName),
			Type:             "classic",
			Scheme:           aws.StringValue(lb.Scheme),
			VPCID:            aws.StringValue(lb.VPCId),
			AccountID:        account.ID,
			AccountName:      account.Name,
			Region:           region,
			CreatedTime:      createdTime,
			ListenerCount:    len(lb.ListenerDescriptions),
		}

		// Get instance health for Classic ELB
		instanceCount := len(lb.Instances)
		loadBalancer.TargetCount = instanceCount

		if instanceCount > 0 {
			healthResult, err := elbClient.DescribeInstanceHealth(&elb.DescribeInstanceHealthInput{
				LoadBalancerName: lb.LoadBalancerName,
			})
			if err == nil {
				for _, instanceHealth := range healthResult.InstanceStates {
					if instanceHealth.State != nil && *instanceHealth.State == "InService" {
						loadBalancer.HealthyTargetCount++
					}
				}
			}
		}

		// Consider unused if no instances or no healthy instances
		loadBalancer.IsUnused = instanceCount == 0 || loadBalancer.HealthyTargetCount == 0

		lbs = append(lbs, loadBalancer)
	}

	return lbs, nil
}

// DeleteLoadBalancer deletes a load balancer
func (s *AWSService) DeleteLoadBalancer(accountID, region, loadBalancerArnOrName, lbType string) error {
	sess, err := s.getSessionForAccount(accountID)
	if err != nil {
		return fmt.Errorf("cannot access account %s: %w", accountID, err)
	}

	switch lbType {
	case "application", "network":
		// Use ELBv2 API for ALB/NLB
		elbv2Client := elbv2.New(sess.Copy(&aws.Config{Region: aws.String(region)}))

		_, err = elbv2Client.DeleteLoadBalancer(&elbv2.DeleteLoadBalancerInput{
			LoadBalancerArn: aws.String(loadBalancerArnOrName),
		})
		if err != nil {
			return fmt.Errorf("failed to delete %s load balancer: %v", lbType, err)
		}

	case "classic":
		// Use ELB API for Classic
		elbClient := elb.New(sess.Copy(&aws.Config{Region: aws.String(region)}))

		_, err = elbClient.DeleteLoadBalancer(&elb.DeleteLoadBalancerInput{
			LoadBalancerName: aws.String(loadBalancerArnOrName),
		})
		if err != nil {
			return fmt.Errorf("failed to delete classic load balancer: %v", err)
		}

	default:
		return fmt.Errorf("unknown load balancer type: %s", lbType)
	}

	// Invalidate cache
	s.InvalidateLoadBalancersCache(accountID)

	return nil
}

// InvalidateLoadBalancersCache invalidates load balancers cache for a specific account
func (s *AWSService) InvalidateLoadBalancersCache(accountID string) {
	s.cache.Delete(fmt.Sprintf("load-balancers:%s", accountID))
	// Also invalidate the global cache
	s.cache.Delete("load-balancers")
}

// InvalidateAllLoadBalancersCache invalidates all load balancer caches
func (s *AWSService) InvalidateAllLoadBalancersCache() {
	s.cache.Delete("load-balancers")
	s.cache.DeletePattern("load-balancers:")
}
