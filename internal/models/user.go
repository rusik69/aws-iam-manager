// Package models contains data structures for AWS IAM entities
package models

import "time"

// Account represents an AWS account
type Account struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Accessible bool   `json:"accessible"`
}

// User represents an AWS IAM user
type User struct {
	Username    string      `json:"username"`
	UserID      string      `json:"user_id"`
	Arn         string      `json:"arn"`
	CreateDate  time.Time   `json:"create_date"`
	PasswordSet bool        `json:"password_set"`
	AccessKeys  []AccessKey `json:"access_keys"`
}

// AccessKey represents an AWS access key
type AccessKey struct {
	AccessKeyID  string     `json:"access_key_id"`
	Status       string     `json:"status"`
	CreateDate   time.Time  `json:"create_date"`
	LastUsedDate *time.Time `json:"last_used_date,omitempty"`
	LastUsedService string  `json:"last_used_service,omitempty"`
	LastUsedRegion  string  `json:"last_used_region,omitempty"`
}

// PublicIP represents a public IP address used by AWS resources
type PublicIP struct {
	IPAddress    string `json:"ip_address"`
	AccountID    string `json:"account_id"`
	AccountName  string `json:"account_name"`
	Region       string `json:"region"`
	ResourceType string `json:"resource_type"` // "EC2", "CLB", "ALB", "NLB", "NAT"
	ResourceID   string `json:"resource_id"`
	ResourceName string `json:"resource_name,omitempty"`
	State        string `json:"state,omitempty"` // running, stopped, etc.
}

// SecurityGroup represents an AWS security group
type SecurityGroup struct {
	GroupID       string              `json:"group_id"`
	GroupName     string              `json:"group_name"`
	Description   string              `json:"description"`
	AccountID     string              `json:"account_id"`
	AccountName   string              `json:"account_name"`
	Region        string              `json:"region"`
	VpcID         string              `json:"vpc_id,omitempty"`
	IsDefault     bool                `json:"is_default"`
	IngressRules  []SecurityGroupRule `json:"ingress_rules"`
	EgressRules   []SecurityGroupRule `json:"egress_rules"`
	HasOpenPorts  bool                `json:"has_open_ports"`
	OpenPortsInfo []OpenPortInfo      `json:"open_ports_info,omitempty"`
	IsUnused      bool                `json:"is_unused"`
	UsageInfo     SecurityGroupUsage  `json:"usage_info"`
}

// SecurityGroupUsage represents usage information for a security group
type SecurityGroupUsage struct {
	AttachedToInstances       []string `json:"attached_to_instances,omitempty"`
	AttachedToNetworkInterfaces []string `json:"attached_to_network_interfaces,omitempty"`
	AttachedToLoadBalancers   []string `json:"attached_to_load_balancers,omitempty"`
	ReferencedBySecurityGroups []string `json:"referenced_by_security_groups,omitempty"`
	TotalAttachments          int      `json:"total_attachments"`
}

// SecurityGroupRule represents a security group rule
type SecurityGroupRule struct {
	IpProtocol string `json:"ip_protocol"`
	FromPort   int64  `json:"from_port,omitempty"`
	ToPort     int64  `json:"to_port,omitempty"`
	CidrIPv4   string `json:"cidr_ipv4,omitempty"`
	CidrIPv6   string `json:"cidr_ipv6,omitempty"`
	GroupID    string `json:"group_id,omitempty"`
	GroupOwner string `json:"group_owner,omitempty"`
}

// OpenPortInfo represents information about ports open to the internet
type OpenPortInfo struct {
	Protocol    string `json:"protocol"`
	PortRange   string `json:"port_range"`
	Source      string `json:"source"`
	Description string `json:"description"`
}
