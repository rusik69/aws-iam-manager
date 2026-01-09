// Package models contains data structures for Azure entities
package models

import "time"

// AzureEnterpriseApplication represents an Azure AD Enterprise Application (Service Principal)
type AzureEnterpriseApplication struct {
	ID                string    `json:"id"`
	AppID             string    `json:"app_id"`
	DisplayName       string    `json:"display_name"`
	CreatedDateTime   time.Time `json:"created_datetime"`
	AccountEnabled    bool      `json:"account_enabled"`
	AppOwnerOrgID     string    `json:"app_owner_org_id,omitempty"`
	AppRoleAssignmentRequired bool `json:"app_role_assignment_required"`
	ServicePrincipalType string `json:"service_principal_type"`
	Tags              []string  `json:"tags,omitempty"`
	Homepage          string    `json:"homepage,omitempty"`
	ReplyUrls         []string  `json:"reply_urls,omitempty"`
}

// AzureVM represents an Azure Virtual Machine
type AzureVM struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	ResourceGroup     string    `json:"resource_group"`
	Location          string    `json:"location"`
	VMSize            string    `json:"vm_size"`
	Status            string    `json:"status"`
	ProvisioningState string    `json:"provisioning_state"`
	OsType            string    `json:"os_type"`
	CreatedTime       time.Time `json:"created_time"`
	SubscriptionID    string    `json:"subscription_id"`
}

// AzureStorageAccount represents an Azure Storage Account
type AzureStorageAccount struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	ResourceGroup    string            `json:"resource_group"`
	Location         string            `json:"location"`
	Kind             string            `json:"kind"`
	Sku              string            `json:"sku"`
	CreatedTime      time.Time         `json:"created_time"`
	PrimaryEndpoints map[string]string `json:"primary_endpoints,omitempty"`
	SubscriptionID   string            `json:"subscription_id"`
}

// AzureSubscription represents an Azure Subscription
type AzureSubscription struct {
	ID             string `json:"id"`
	SubscriptionID string `json:"subscription_id"`
	DisplayName    string `json:"display_name"`
	State          string `json:"state"`
	TenantID       string `json:"tenant_id"`
}
