package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rusik69/aws-iam-manager/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock AWS service for testing
type MockAWSService struct{}

func (m *MockAWSService) ListAccounts() ([]models.Account, error) {
	return []models.Account{
		{ID: "123456789012", Name: "Test Account 1"},
		{ID: "123456789013", Name: "Test Account 2"},
	}, nil
}

func (m *MockAWSService) ListUsers(accountID string) ([]models.User, error) {
	lastUsedDate := time.Now().Add(-48 * time.Hour) // 2 days ago
	return []models.User{
		{
			Username:    "testuser1",
			UserID:      "AIDACKCEVSQ6C2EXAMPLE",
			Arn:         "arn:aws:iam::123456789012:user/testuser1",
			CreateDate:  time.Now(),
			PasswordSet: true,
			AccessKeys: []models.AccessKey{
				{
					AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
					Status:          "Active",
					CreateDate:      time.Now(),
					LastUsedDate:    &lastUsedDate,
					LastUsedService: "ec2",
					LastUsedRegion:  "us-east-1",
				},
			},
		},
	}, nil
}

func (m *MockAWSService) GetUser(accountID, username string) (*models.User, error) {
	lastUsedDate := time.Now().Add(-48 * time.Hour) // 2 days ago
	user := &models.User{
		Username:    "testuser1",
		UserID:      "AIDACKCEVSQ6C2EXAMPLE",
		Arn:         "arn:aws:iam::123456789012:user/testuser1",
		CreateDate:  time.Now(),
		PasswordSet: true,
		AccessKeys: []models.AccessKey{
			{
				AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
				Status:          "Active",
				CreateDate:      time.Now(),
				LastUsedDate:    &lastUsedDate,
				LastUsedService: "ec2",
				LastUsedRegion:  "us-east-1",
			},
		},
	}
	return user, nil
}

func (m *MockAWSService) CreateAccessKey(accountID, username string) (map[string]any, error) {
	return map[string]any{
		"access_key_id":     "AKIAIOSFODNN7EXAMPLE2",
		"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"status":            "Active",
		"create_date":       time.Now(),
	}, nil
}

func (m *MockAWSService) DeleteAccessKey(accountID, username, keyID string) error {
	return nil
}

func (m *MockAWSService) RotateAccessKey(accountID, username, keyID string) (map[string]any, error) {
	return map[string]any{
		"access_key_id":     "AKIAIOSFODNN7EXAMPLE3",
		"secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY2",
		"status":            "Active",
		"create_date":       time.Now(),
		"message":           "Access key rotated successfully",
	}, nil
}

func (m *MockAWSService) DeleteUser(accountID, username string) error {
	return nil
}

func (m *MockAWSService) ListPublicIPs() ([]models.PublicIP, error) {
	return []models.PublicIP{
		{
			IPAddress:    "54.123.45.67",
			AccountID:    "123456789012",
			AccountName:  "Test Account 1",
			Region:       "us-east-1",
			ResourceType: "EC2",
			ResourceID:   "i-1234567890abcdef0",
			ResourceName: "test-instance",
			State:        "running",
		},
		{
			IPAddress:    "52.98.76.54",
			AccountID:    "123456789013",
			AccountName:  "Test Account 2",
			Region:       "us-west-2",
			ResourceType: "application",
			ResourceID:   "arn:aws:elasticloadbalancing:us-west-2:123456789013:loadbalancer/app/test-alb/50dc6c495c0c9188",
			ResourceName: "test-alb",
			State:        "active",
		},
	}, nil
}

func (m *MockAWSService) DeleteUserPassword(accountID, username string) error {
	return nil
}

func (m *MockAWSService) RotateUserPassword(accountID, username string) (map[string]any, error) {
	return map[string]any{
		"message":      "User password rotated successfully",
		"new_password": "NewRandomPassword123!",
		"username":     username,
	}, nil
}

func (m *MockAWSService) ClearCache() {}

func (m *MockAWSService) InvalidateAccountCache(accountID string) {}

func (m *MockAWSService) InvalidateUserCache(accountID, username string) {}

func (m *MockAWSService) InvalidatePublicIPsCache() {}

func (m *MockAWSService) ListSecurityGroups() ([]models.SecurityGroup, error) {
	return []models.SecurityGroup{
		{
			GroupID:     "sg-12345678",
			GroupName:   "test-security-group",
			Description: "Test security group with open ports",
			AccountID:   "123456789012",
			AccountName: "Test Account 1",
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
				AttachedToInstances: []string{"i-1234567890abcdef0"},
				TotalAttachments:    1,
			},
		},
		{
			GroupID:     "sg-87654321",
			GroupName:   "default",
			Description: "Default security group for VPC",
			AccountID:   "123456789013",
			AccountName: "Test Account 2",
			Region:      "us-west-2",
			VpcID:       "vpc-87654321",
			IsDefault:   true,
			IngressRules: []models.SecurityGroupRule{
				{
					IpProtocol: "tcp",
					FromPort:   22,
					ToPort:     22,
					CidrIPv4:   "10.0.0.0/8",
				},
			},
			EgressRules: []models.SecurityGroupRule{
				{
					IpProtocol: "-1",
					CidrIPv4:   "0.0.0.0/0",
				},
			},
			HasOpenPorts:  false,
			OpenPortsInfo: []models.OpenPortInfo{},
			IsUnused:      true,
			UsageInfo: models.SecurityGroupUsage{
				TotalAttachments: 0,
			},
		},
	}, nil
}

func (m *MockAWSService) InvalidateSecurityGroupsCache() {}

func (m *MockAWSService) DeleteSecurityGroup(accountID, region, groupID string) error {
	// Mock implementation - simulate successful deletion for non-default groups
	if groupID == "sg-default" {
		return fmt.Errorf("cannot delete default security group")
	}
	if groupID == "sg-in-use" {
		return fmt.Errorf("security group %s is still in use (attached to 2 resources)", groupID)
	}
	if groupID == "sg-not-found" {
		return fmt.Errorf("security group %s not found", groupID)
	}
	// For all other group IDs, simulate successful deletion
	return nil
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := &MockAWSService{}
	handler := NewHandler(mockService)

	api := r.Group("/api")
	{
		api.GET("/accounts", handler.ListAccounts)
		api.GET("/accounts/:accountId/users", handler.ListUsers)
		api.GET("/accounts/:accountId/users/:username", handler.GetUser)
		api.DELETE("/accounts/:accountId/users/:username", handler.DeleteUser)
		api.POST("/accounts/:accountId/users/:username/keys", handler.CreateAccessKey)
		api.DELETE("/accounts/:accountId/users/:username/keys/:keyId", handler.DeleteAccessKey)
		api.PUT("/accounts/:accountId/users/:username/keys/:keyId/rotate", handler.RotateAccessKey)
		api.GET("/public-ips", handler.ListPublicIPs)
		api.GET("/security-groups", handler.ListSecurityGroups)
		api.DELETE("/accounts/:accountId/regions/:region/security-groups/:groupId", handler.DeleteSecurityGroup)
	}

	return r
}

func TestListAccounts(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/accounts", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var accounts []models.Account
	err := json.Unmarshal(w.Body.Bytes(), &accounts)
	assert.NoError(t, err)
	assert.Len(t, accounts, 2)
	assert.Equal(t, "Test Account 1", accounts[0].Name)
}

func TestListUsers(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/accounts/123456789012/users", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var users []models.User
	err := json.Unmarshal(w.Body.Bytes(), &users)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "testuser1", users[0].Username)
	assert.True(t, users[0].PasswordSet)
	assert.Len(t, users[0].AccessKeys, 1)
}

func TestGetUser(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/accounts/123456789012/users/testuser1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user models.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "testuser1", user.Username)
	assert.Equal(t, "AIDACKCEVSQ6C2EXAMPLE", user.UserID)
}

func TestCreateAccessKey(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/accounts/123456789012/users/testuser1/keys", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "access_key_id")
	assert.Contains(t, response, "secret_access_key")
	assert.Equal(t, "Active", response["status"])
}

func TestDeleteAccessKey(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/accounts/123456789012/users/testuser1/keys/AKIAIOSFODNN7EXAMPLE", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Access key deleted successfully", response["message"])
}

func TestRotateAccessKey(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/accounts/123456789012/users/testuser1/keys/AKIAIOSFODNN7EXAMPLE/rotate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "access_key_id")
	assert.Contains(t, response, "secret_access_key")
	assert.Equal(t, "Access key rotated successfully", response["message"])
}

func TestDeleteUser(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/accounts/123456789012/users/testuser1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "User deleted successfully", response["message"])
}

func TestListPublicIPs(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/public-ips", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var ips []models.PublicIP
	err := json.Unmarshal(w.Body.Bytes(), &ips)
	assert.NoError(t, err)
	assert.Len(t, ips, 2)
	assert.Equal(t, "54.123.45.67", ips[0].IPAddress)
	assert.Equal(t, "EC2", ips[0].ResourceType)
	assert.Equal(t, "us-east-1", ips[0].Region)
	assert.Equal(t, "52.98.76.54", ips[1].IPAddress)
	assert.Equal(t, "application", ips[1].ResourceType)
	assert.Equal(t, "us-west-2", ips[1].Region)
}

func TestListSecurityGroups(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/security-groups", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var sgs []models.SecurityGroup
	err := json.Unmarshal(w.Body.Bytes(), &sgs)
	assert.NoError(t, err)
	assert.Len(t, sgs, 2)

	// Test first security group (with open ports)
	assert.Equal(t, "sg-12345678", sgs[0].GroupID)
	assert.Equal(t, "test-security-group", sgs[0].GroupName)
	assert.Equal(t, "Test security group with open ports", sgs[0].Description)
	assert.Equal(t, "123456789012", sgs[0].AccountID)
	assert.Equal(t, "Test Account 1", sgs[0].AccountName)
	assert.Equal(t, "us-east-1", sgs[0].Region)
	assert.Equal(t, "vpc-12345678", sgs[0].VpcID)
	assert.False(t, sgs[0].IsDefault)
	assert.True(t, sgs[0].HasOpenPorts)
	assert.Len(t, sgs[0].IngressRules, 1)
	assert.Len(t, sgs[0].EgressRules, 1)
	assert.Len(t, sgs[0].OpenPortsInfo, 1)

	// Test second security group (default, no open ports)
	assert.Equal(t, "sg-87654321", sgs[1].GroupID)
	assert.Equal(t, "default", sgs[1].GroupName)
	assert.Equal(t, "Default security group for VPC", sgs[1].Description)
	assert.Equal(t, "123456789013", sgs[1].AccountID)
	assert.Equal(t, "Test Account 2", sgs[1].AccountName)
	assert.Equal(t, "us-west-2", sgs[1].Region)
	assert.Equal(t, "vpc-87654321", sgs[1].VpcID)
	assert.True(t, sgs[1].IsDefault)
	assert.False(t, sgs[1].HasOpenPorts)
	assert.Len(t, sgs[1].IngressRules, 1)
	assert.Len(t, sgs[1].EgressRules, 1)
	assert.Len(t, sgs[1].OpenPortsInfo, 0)

	// Test open ports info detail
	openPort := sgs[0].OpenPortsInfo[0]
	assert.Equal(t, "tcp", openPort.Protocol)
	assert.Equal(t, "80", openPort.PortRange)
	assert.Equal(t, "0.0.0.0/0 (IPv4 Internet)", openPort.Source)
	assert.Equal(t, "TCP traffic", openPort.Description)
}

func TestDeleteSecurityGroup(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name           string
		accountID      string
		region         string
		groupID        string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Delete valid security group",
			accountID:      "123456789012",
			region:         "us-east-1",
			groupID:        "sg-12345678",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Delete default security group",
			accountID:      "123456789012",
			region:         "us-east-1",
			groupID:        "sg-default",
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
		{
			name:           "Delete security group in use",
			accountID:      "123456789012",
			region:         "us-east-1",
			groupID:        "sg-in-use",
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
		{
			name:           "Delete non-existent security group",
			accountID:      "123456789012",
			region:         "us-east-1",
			groupID:        "sg-not-found",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "Missing parameters",
			accountID:      "",
			region:         "us-east-1",
			groupID:        "sg-12345678",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/accounts/%s/regions/%s/security-groups/%s", tt.accountID, tt.region, tt.groupID)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]any
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectError {
				assert.Contains(t, response, "error")
			} else {
				assert.Contains(t, response, "message")
				assert.Contains(t, response["message"], "deleted successfully")
			}
		})
	}
}
