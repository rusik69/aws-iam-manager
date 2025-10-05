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
