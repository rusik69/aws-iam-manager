package services

import (
	"fmt"

	"github.com/rusik69/aws-iam-manager/internal/models"

	"github.com/aws/aws-sdk-go/service/organizations"
)

// ListAccounts returns all accounts in the AWS organization
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

// InvalidateAccountCache invalidates cache for a specific account
func (s *AWSService) InvalidateAccountCache(accountID string) {
	s.cache.Delete(fmt.Sprintf("users:%s", accountID))
	s.cache.Delete(fmt.Sprintf("security-groups:%s", accountID))
	s.cache.Delete(fmt.Sprintf("snapshots:%s", accountID))
	s.cache.Delete(fmt.Sprintf("ec2-instances:%s", accountID))
	s.cache.Delete(fmt.Sprintf("ebs-volumes:%s", accountID))
	s.cache.Delete(fmt.Sprintf("s3-buckets:%s", accountID))
	s.cache.Delete(fmt.Sprintf("roles:%s", accountID))
	s.cache.Delete(fmt.Sprintf("load-balancers:%s", accountID))
	s.cache.Delete(fmt.Sprintf("vpcs-%s", accountID))
	s.cache.Delete(fmt.Sprintf("nat-gateways-%s", accountID))
}
