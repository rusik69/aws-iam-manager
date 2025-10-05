// Package services provides business logic and external service integrations
package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	"github.com/rusik69/aws-iam-manager/internal/models"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
)

// AzureService handles Azure AD operations
type AzureService struct {
	client   *msgraphsdk.GraphServiceClient
	cache    *Cache
	cacheTTL time.Duration
}

// NewAzureService creates a new Azure service instance
func NewAzureService() (*AzureService, error) {
	// Get Azure credentials from environment variables
	tenantID := os.Getenv("AZURE_TENANT_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

	if tenantID == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("Azure credentials not configured. Set AZURE_TENANT_ID, AZURE_CLIENT_ID, and AZURE_CLIENT_SECRET environment variables")
	}

	// Create credential
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	// Create Microsoft Graph client
	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("failed to create Graph client: %w", err)
	}

	return &AzureService{
		client:   client,
		cache:    NewCache(),
		cacheTTL: 5 * time.Minute,
	}, nil
}

// ListEnterpriseApplications lists all enterprise applications (service principals) in Azure AD
func (s *AzureService) ListEnterpriseApplications(ctx context.Context) ([]models.AzureEnterpriseApplication, error) {
	const cacheKey = "azure-enterprise-apps"

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if apps, ok := cached.([]models.AzureEnterpriseApplication); ok {
			return apps, nil
		}
	}

	// Get all service principals (enterprise applications)
	result, err := s.client.ServicePrincipals().Get(ctx, &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
		QueryParameters: &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
			Select: []string{"id", "appId", "displayName", "accountEnabled", "appOwnerOrganizationId", "appRoleAssignmentRequired", "servicePrincipalType", "tags", "homepage", "replyUrls"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list enterprise applications: %w", err)
	}

	var apps []models.AzureEnterpriseApplication
	servicePrincipals := result.GetValue()

	for _, sp := range servicePrincipals {
		app := models.AzureEnterpriseApplication{
			ID:                        getStringValue(sp.GetId()),
			AppID:                     getStringValue(sp.GetAppId()),
			DisplayName:               getStringValue(sp.GetDisplayName()),
			AccountEnabled:            getBoolValue(sp.GetAccountEnabled()),
			AppOwnerOrgID:             getUUIDStringValue(sp.GetAppOwnerOrganizationId()),
			AppRoleAssignmentRequired: getBoolValue(sp.GetAppRoleAssignmentRequired()),
			ServicePrincipalType:      getStringValue(sp.GetServicePrincipalType()),
			Tags:                      sp.GetTags(),
			Homepage:                  getStringValue(sp.GetHomepage()),
			ReplyUrls:                 sp.GetReplyUrls(),
			CreatedDateTime:           time.Now(), // Default to now since this field is not available
		}

		apps = append(apps, app)
	}

	// Cache the result
	s.cache.Set(cacheKey, apps, s.cacheTTL)

	return apps, nil
}

// GetEnterpriseApplication gets a specific enterprise application by ID
func (s *AzureService) GetEnterpriseApplication(ctx context.Context, appID string) (*models.AzureEnterpriseApplication, error) {
	cacheKey := fmt.Sprintf("azure-enterprise-app:%s", appID)

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if app, ok := cached.(*models.AzureEnterpriseApplication); ok {
			return app, nil
		}
	}

	// Get the service principal
	sp, err := s.client.ServicePrincipals().ByServicePrincipalId(appID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get enterprise application: %w", err)
	}

	app := &models.AzureEnterpriseApplication{
		ID:                        getStringValue(sp.GetId()),
		AppID:                     getStringValue(sp.GetAppId()),
		DisplayName:               getStringValue(sp.GetDisplayName()),
		AccountEnabled:            getBoolValue(sp.GetAccountEnabled()),
		AppOwnerOrgID:             getUUIDStringValue(sp.GetAppOwnerOrganizationId()),
		AppRoleAssignmentRequired: getBoolValue(sp.GetAppRoleAssignmentRequired()),
		ServicePrincipalType:      getStringValue(sp.GetServicePrincipalType()),
		Tags:                      sp.GetTags(),
		Homepage:                  getStringValue(sp.GetHomepage()),
		ReplyUrls:                 sp.GetReplyUrls(),
		CreatedDateTime:           time.Now(), // Default to now since this field is not available
	}

	// Cache the result
	s.cache.Set(cacheKey, app, s.cacheTTL)

	return app, nil
}

// DeleteEnterpriseApplication deletes an enterprise application (service principal) by ID
func (s *AzureService) DeleteEnterpriseApplication(ctx context.Context, appID string) error {
	// Delete the service principal
	err := s.client.ServicePrincipals().ByServicePrincipalId(appID).Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete enterprise application: %w", err)
	}

	// Invalidate cache
	s.cache.Delete(fmt.Sprintf("azure-enterprise-app:%s", appID))
	s.cache.Delete("azure-enterprise-apps")

	return nil
}

// ClearCache clears all cached data
func (s *AzureService) ClearCache() {
	s.cache.Clear()
}

// InvalidateEnterpriseApplicationsCache invalidates the enterprise applications cache
func (s *AzureService) InvalidateEnterpriseApplicationsCache() {
	s.cache.Delete("azure-enterprise-apps")
}

// InvalidateEnterpriseApplicationCache invalidates cache for a specific enterprise application
func (s *AzureService) InvalidateEnterpriseApplicationCache(appID string) {
	s.cache.Delete(fmt.Sprintf("azure-enterprise-app:%s", appID))
	s.cache.Delete("azure-enterprise-apps")
}

// Helper functions to safely get values from pointers
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func getBoolValue(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

func getUUIDStringValue(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}
