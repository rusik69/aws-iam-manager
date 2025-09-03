package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"aws-iam-manager/internal/models"
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
	return []models.User{
		{
			Username:    "testuser1",
			UserID:      "AIDACKCEVSQ6C2EXAMPLE",
			Arn:         "arn:aws:iam::123456789012:user/testuser1",
			CreateDate:  time.Now(),
			PasswordSet: true,
			AccessKeys: []models.AccessKey{
				{
					AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
					Status:      "Active",
					CreateDate:  time.Now(),
				},
			},
		},
	}, nil
}

func (m *MockAWSService) GetUser(accountID, username string) (*models.User, error) {
	user := &models.User{
		Username:    "testuser1",
		UserID:      "AIDACKCEVSQ6C2EXAMPLE",
		Arn:         "arn:aws:iam::123456789012:user/testuser1",
		CreateDate:  time.Now(),
		PasswordSet: true,
		AccessKeys: []models.AccessKey{
			{
				AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
				Status:      "Active",
				CreateDate:  time.Now(),
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
		api.POST("/accounts/:accountId/users/:username/keys", handler.CreateAccessKey)
		api.DELETE("/accounts/:accountId/users/:username/keys/:keyId", handler.DeleteAccessKey)
		api.PUT("/accounts/:accountId/users/:username/keys/:keyId/rotate", handler.RotateAccessKey)
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
