// Package models contains data structures for AWS SSO (IAM Identity Center) entities
package models

// SSOUser represents an AWS SSO (IAM Identity Center) user
type SSOUser struct {
	UserID      string   `json:"user_id"`
	UserName    string   `json:"user_name"`
	DisplayName string   `json:"display_name,omitempty"`
	Emails      []string `json:"emails,omitempty"`
	Active      bool     `json:"active"`
}

// SSOGroup represents an AWS SSO (IAM Identity Center) group
type SSOGroup struct {
	GroupID      string `json:"group_id"`
	DisplayName  string `json:"display_name"`
	Description  string `json:"description,omitempty"`
}

// SSOGroupMember represents a member of an SSO group
type SSOGroupMember struct {
	MemberID   string `json:"member_id"`
	MemberType string `json:"member_type"` // "USER" or "GROUP"
	UserName   string `json:"user_name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// SSOAccountAssignment represents an account assignment in IAM Identity Center
type SSOAccountAssignment struct {
	AccountID       string `json:"account_id"`
	AccountName     string `json:"account_name,omitempty"`
	PrincipalID     string `json:"principal_id"`
	PrincipalType   string `json:"principal_type"` // "USER" or "GROUP"
	PrincipalName   string `json:"principal_name,omitempty"`
	PermissionSetArn string `json:"permission_set_arn"`
	PermissionSetName string `json:"permission_set_name,omitempty"`
}

// SSOUserWithAssignments represents an SSO user with their account assignments
type SSOUserWithAssignments struct {
	SSOUser
	AccountAssignments []SSOAccountAssignment `json:"account_assignments"`
	GroupMemberships   []string               `json:"group_memberships,omitempty"` // Group IDs
}

// SSOGroupWithAssignments represents an SSO group with their account assignments
type SSOGroupWithAssignments struct {
	SSOGroup
	AccountAssignments []SSOAccountAssignment `json:"account_assignments"`
	MemberCount        int                     `json:"member_count"`
	Members            []SSOGroupMember        `json:"members,omitempty"`
}

// SSOAccountWithAssignments represents an AWS account with SSO assignments
type SSOAccountWithAssignments struct {
	AccountID   string               `json:"account_id"`
	AccountName string               `json:"account_name"`
	Assignments []SSOAccountAssignment `json:"assignments"`
}

// SSOIdentityCenterInstance represents IAM Identity Center instance information
type SSOIdentityCenterInstance struct {
	InstanceARN      string `json:"instance_arn"`
	IdentityStoreID string `json:"identity_store_id"`
}
