package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rusik69/aws-iam-manager/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestModelsCreation(t *testing.T) {
	// Test Account struct
	account := models.Account{
		ID:   "123456789012",
		Name: "Test Account",
	}
	assert.Equal(t, "123456789012", account.ID)
	assert.Equal(t, "Test Account", account.Name)

	// Test User struct
	createDate := time.Now()
	user := models.User{
		Username:    "testuser",
		UserID:      "AIDACKCEVSQ6C2EXAMPLE",
		Arn:         "arn:aws:iam::123456789012:user/testuser",
		CreateDate:  createDate,
		PasswordSet: true,
		AccessKeys: []models.AccessKey{
			{
				AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
				Status:      "Active",
				CreateDate:  createDate,
			},
		},
	}

	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "AIDACKCEVSQ6C2EXAMPLE", user.UserID)
	assert.Equal(t, "arn:aws:iam::123456789012:user/testuser", user.Arn)
	assert.Equal(t, createDate, user.CreateDate)
	assert.True(t, user.PasswordSet)
	assert.Len(t, user.AccessKeys, 1)

	// Test AccessKey struct
	lastUsedDate := time.Now().Add(-24 * time.Hour) // 1 day ago
	accessKey := models.AccessKey{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		Status:          "Active",
		CreateDate:      createDate,
		LastUsedDate:    &lastUsedDate,
		LastUsedService: "s3",
		LastUsedRegion:  "us-east-1",
	}

	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", accessKey.AccessKeyID)
	assert.Equal(t, "Active", accessKey.Status)
	assert.Equal(t, createDate, accessKey.CreateDate)
	assert.NotNil(t, accessKey.LastUsedDate)
	assert.Equal(t, lastUsedDate, *accessKey.LastUsedDate)
	assert.Equal(t, "s3", accessKey.LastUsedService)
	assert.Equal(t, "us-east-1", accessKey.LastUsedRegion)

	// Test AccessKey without last used info (never used key)
	neverUsedKey := models.AccessKey{
		AccessKeyID:  "AKIAIOSFODNN7EXAMPLE2",
		Status:       "Active",
		CreateDate:   createDate,
		LastUsedDate: nil,
	}

	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE2", neverUsedKey.AccessKeyID)
	assert.Equal(t, "Active", neverUsedKey.Status)
	assert.Nil(t, neverUsedKey.LastUsedDate)
	assert.Empty(t, neverUsedKey.LastUsedService)
	assert.Empty(t, neverUsedKey.LastUsedRegion)

	// Test SecurityGroup struct
	securityGroup := models.SecurityGroup{
		GroupID:     "sg-12345678",
		GroupName:   "test-security-group",
		Description: "Test security group",
		AccountID:   "123456789012",
		AccountName: "Test Account",
		Region:      "us-east-1",
		VpcID:       "vpc-12345678",
		IsDefault:   false,
		IngressRules: []models.SecurityGroupRule{
			{
				IpProtocol: "tcp",
				FromPort:   80,
				ToPort:     80,
				CidrIPv4:   "0.0.0.0/0",
			},
		},
		EgressRules: []models.SecurityGroupRule{
			{
				IpProtocol: "-1",
				CidrIPv4:   "0.0.0.0/0",
			},
		},
		HasOpenPorts: true,
		OpenPortsInfo: []models.OpenPortInfo{
			{
				Protocol:    "tcp",
				PortRange:   "80",
				Source:      "0.0.0.0/0 (IPv4 Internet)",
				Description: "TCP traffic",
			},
		},
		IsUnused: false,
		UsageInfo: models.SecurityGroupUsage{
			AttachedToInstances: []string{"i-example"},
			TotalAttachments:    1,
		},
	}

	assert.Equal(t, "sg-12345678", securityGroup.GroupID)
	assert.Equal(t, "test-security-group", securityGroup.GroupName)
	assert.Equal(t, "Test security group", securityGroup.Description)
	assert.Equal(t, "123456789012", securityGroup.AccountID)
	assert.Equal(t, "Test Account", securityGroup.AccountName)
	assert.Equal(t, "us-east-1", securityGroup.Region)
	assert.Equal(t, "vpc-12345678", securityGroup.VpcID)
	assert.False(t, securityGroup.IsDefault)
	assert.True(t, securityGroup.HasOpenPorts)
	assert.Len(t, securityGroup.IngressRules, 1)
	assert.Len(t, securityGroup.EgressRules, 1)
	assert.Len(t, securityGroup.OpenPortsInfo, 1)

	// Test SecurityGroupRule struct
	rule := models.SecurityGroupRule{
		IpProtocol: "tcp",
		FromPort:   22,
		ToPort:     22,
		CidrIPv4:   "10.0.0.0/8",
	}

	assert.Equal(t, "tcp", rule.IpProtocol)
	assert.Equal(t, int64(22), rule.FromPort)
	assert.Equal(t, int64(22), rule.ToPort)
	assert.Equal(t, "10.0.0.0/8", rule.CidrIPv4)

	// Test OpenPortInfo struct
	openPortInfo := models.OpenPortInfo{
		Protocol:    "tcp",
		PortRange:   "22",
		Source:      "0.0.0.0/0 (IPv4 Internet)",
		Description: "TCP traffic",
	}

	assert.Equal(t, "tcp", openPortInfo.Protocol)
	assert.Equal(t, "22", openPortInfo.PortRange)
	assert.Equal(t, "0.0.0.0/0 (IPv4 Internet)", openPortInfo.Source)
	assert.Equal(t, "TCP traffic", openPortInfo.Description)
}

func TestCheckForOpenPorts(t *testing.T) {
	service := &AWSService{}

	tests := []struct {
		name          string
		ingressRules  []models.SecurityGroupRule
		expectedOpen  bool
		expectedPorts int
	}{
		{
			name: "No open ports",
			ingressRules: []models.SecurityGroupRule{
				{
					IpProtocol: "tcp",
					FromPort:   22,
					ToPort:     22,
					CidrIPv4:   "10.0.0.0/8",
				},
			},
			expectedOpen:  false,
			expectedPorts: 0,
		},
		{
			name: "Open to IPv4 internet",
			ingressRules: []models.SecurityGroupRule{
				{
					IpProtocol: "tcp",
					FromPort:   80,
					ToPort:     80,
					CidrIPv4:   "0.0.0.0/0",
				},
			},
			expectedOpen:  true,
			expectedPorts: 1,
		},
		{
			name: "Open to IPv6 internet",
			ingressRules: []models.SecurityGroupRule{
				{
					IpProtocol: "tcp",
					FromPort:   443,
					ToPort:     443,
					CidrIPv6:   "::/0",
				},
			},
			expectedOpen:  true,
			expectedPorts: 1,
		},
		{
			name: "Multiple open ports",
			ingressRules: []models.SecurityGroupRule{
				{
					IpProtocol: "tcp",
					FromPort:   80,
					ToPort:     80,
					CidrIPv4:   "0.0.0.0/0",
				},
				{
					IpProtocol: "tcp",
					FromPort:   443,
					ToPort:     443,
					CidrIPv4:   "0.0.0.0/0",
				},
			},
			expectedOpen:  true,
			expectedPorts: 2,
		},
		{
			name: "All ports open",
			ingressRules: []models.SecurityGroupRule{
				{
					IpProtocol: "-1",
					CidrIPv4:   "0.0.0.0/0",
				},
			},
			expectedOpen:  true,
			expectedPorts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasOpen, openPorts := service.checkForOpenPorts(tt.ingressRules)
			assert.Equal(t, tt.expectedOpen, hasOpen)
			assert.Len(t, openPorts, tt.expectedPorts)
		})
	}
}

func TestUserWithLastUsedDates(t *testing.T) {
	// Test a user with multiple access keys having different last used dates
	oldDate := time.Now().Add(-48 * time.Hour) // 2 days ago
	recentDate := time.Now().Add(-24 * time.Hour) // 1 day ago

	user := models.User{
		Username:    "testuser",
		UserID:      "AIDACKCEVSQ6C2EXAMPLE",
		Arn:         "arn:aws:iam::123456789012:user/testuser",
		CreateDate:  time.Now(),
		PasswordSet: true,
		AccessKeys: []models.AccessKey{
			{
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE1",
				Status:          "Active",
				CreateDate:      time.Now().Add(-72 * time.Hour), // 3 days ago
				LastUsedDate:    &oldDate,
				LastUsedService: "s3",
				LastUsedRegion:  "us-east-1",
			},
			{
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE2",
				Status:          "Active",
				CreateDate:      time.Now().Add(-36 * time.Hour), // 1.5 days ago
				LastUsedDate:    &recentDate,
				LastUsedService: "ec2",
				LastUsedRegion:  "us-west-2",
			},
			{
				AccessKeyID:  "AKIAIOSFODNN7EXAMPLE3",
				Status:       "Active",
				CreateDate:   time.Now().Add(-12 * time.Hour), // 12 hours ago
				LastUsedDate: nil, // Never used
			},
		},
	}

	assert.Equal(t, "testuser", user.Username)
	assert.Len(t, user.AccessKeys, 3)

	// Verify the first key (older last used)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE1", user.AccessKeys[0].AccessKeyID)
	assert.NotNil(t, user.AccessKeys[0].LastUsedDate)
	assert.Equal(t, oldDate, *user.AccessKeys[0].LastUsedDate)
	assert.Equal(t, "s3", user.AccessKeys[0].LastUsedService)

	// Verify the second key (more recent last used)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE2", user.AccessKeys[1].AccessKeyID)
	assert.NotNil(t, user.AccessKeys[1].LastUsedDate)
	assert.Equal(t, recentDate, *user.AccessKeys[1].LastUsedDate)
	assert.Equal(t, "ec2", user.AccessKeys[1].LastUsedService)

	// Verify the third key (never used)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE3", user.AccessKeys[2].AccessKeyID)
	assert.Nil(t, user.AccessKeys[2].LastUsedDate)
	assert.Empty(t, user.AccessKeys[2].LastUsedService)
}

func TestAccessKeySearchFunctionality(t *testing.T) {
	// Test data with access keys for search testing
	users := []models.User{
		{
			Username: "alice",
			UserID:   "AIDAALICE123",
			AccessKeys: []models.AccessKey{
				{
					AccessKeyID:     "AKIAALICE001",
					Status:          "Active",
					CreateDate:      time.Now(),
					LastUsedService: "s3",
					LastUsedRegion:  "us-east-1",
				},
			},
		},
		{
			Username: "bob",
			UserID:   "AIDABOB456",
			AccessKeys: []models.AccessKey{
				{
					AccessKeyID:     "AKIABOB002",
					Status:          "Inactive",
					CreateDate:      time.Now(),
					LastUsedService: "ec2",
					LastUsedRegion:  "us-west-2",
				},
				{
					AccessKeyID:     "AKIABOB003",
					Status:          "Active",
					CreateDate:      time.Now(),
					LastUsedService: "dynamodb",
					LastUsedRegion:  "eu-west-1",
				},
			},
		},
	}

	// Test search by access key ID and other fields
	searchResults := []struct {
		query    string
		expected []string // expected usernames
	}{
		{"AKIAALICE001", []string{"alice"}},
		{"AKIABOB", []string{"bob"}}, // Should match both keys for bob
		{"002", []string{"bob"}},     // Partial match
		{"NONEXISTENT", []string{}},  // No matches
		{"alice", []string{"alice"}}, // Search by username
		{"AIDABOB456", []string{"bob"}}, // Search by user ID
	}

	for _, test := range searchResults {
		t.Run(fmt.Sprintf("Search_%s", test.query), func(t *testing.T) {
			matchedUsers := []string{} // Initialize to empty slice, not nil
			query := strings.ToLower(test.query)

			for _, user := range users {
				// Simulate the search logic from frontend components
				if strings.Contains(strings.ToLower(user.Username), query) ||
					strings.Contains(strings.ToLower(user.UserID), query) {
					matchedUsers = append(matchedUsers, user.Username)
					continue
				}

				// Check access keys
				for _, key := range user.AccessKeys {
					if strings.Contains(strings.ToLower(key.AccessKeyID), query) {
						matchedUsers = append(matchedUsers, user.Username)
						break
					}
				}
			}

			assert.Equal(t, test.expected, matchedUsers, "Search for '%s' should return users: %v, got: %v", test.query, test.expected, matchedUsers)
		})
	}

	// Test access key field searches
	t.Run("AccessKeyFieldSearch", func(t *testing.T) {
		user := users[1] // bob with multiple keys

		// Verify we can find keys by service and region
		s3Key := false
		dynamoKey := false

		for _, key := range user.AccessKeys {
			if key.LastUsedService == "ec2" && key.LastUsedRegion == "us-west-2" {
				s3Key = true
			}
			if key.LastUsedService == "dynamodb" && key.LastUsedRegion == "eu-west-1" {
				dynamoKey = true
			}
		}

		assert.True(t, s3Key, "Should find key with ec2 service in us-west-2")
		assert.True(t, dynamoKey, "Should find key with dynamodb service in eu-west-1")
	})
}
