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

// UserWithAccount represents an AWS IAM user with account information
type UserWithAccount struct {
	User
	AccountID   string `json:"accountId"`
	AccountName string `json:"accountName"`
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

// Snapshot represents an AWS EBS snapshot
type Snapshot struct {
	SnapshotID  string    `json:"snapshot_id"`
	VolumeID    string    `json:"volume_id"`
	VolumeSize  int64     `json:"volume_size"` // in GiB
	Description string    `json:"description"`
	State       string    `json:"state"` // pending, completed, error
	Progress    string    `json:"progress"`
	StartTime   time.Time `json:"start_time"`
	OwnerID     string    `json:"owner_id"`
	Encrypted   bool      `json:"encrypted"`
	AccountID   string    `json:"account_id"`
	AccountName string    `json:"account_name"`
	Region      string    `json:"region"`
	Tags        []Tag     `json:"tags,omitempty"`
}

// Tag represents an AWS resource tag
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// EC2Instance represents an AWS EC2 instance
type EC2Instance struct {
	InstanceID   string    `json:"instance_id"`
	Name         string    `json:"name"`          // From "Name" tag
	AccountID    string    `json:"account_id"`
	AccountName  string    `json:"account_name"`
	Region       string    `json:"region"`
	InstanceType string    `json:"instance_type"` // Flavor
	LaunchTime   time.Time `json:"launch_time"`
	State        string    `json:"state"` // running, stopped, etc.
	Tags         []Tag     `json:"tags,omitempty"`
}

// EBSVolume represents an AWS EBS volume
type EBSVolume struct {
	VolumeID         string    `json:"volume_id"`
	Name             string    `json:"name"` // From "Name" tag
	AccountID        string    `json:"account_id"`
	AccountName      string    `json:"account_name"`
	Region           string    `json:"region"`
	Size             int64     `json:"size"`              // in GiB
	VolumeType       string    `json:"volume_type"`       // gp2, gp3, io1, io2, st1, sc1, standard
	State            string    `json:"state"`             // creating, available, in-use, deleting, deleted, error
	CreateTime       time.Time `json:"create_time"`
	AvailabilityZone string    `json:"availability_zone"`
	Encrypted        bool      `json:"encrypted"`
	IOPS             int64     `json:"iops,omitempty"`
	Throughput       int64     `json:"throughput,omitempty"` // MB/s (for gp3)
	SnapshotID       string    `json:"snapshot_id,omitempty"`
	Attachments      []VolumeAttachment `json:"attachments,omitempty"`
	Tags             []Tag     `json:"tags,omitempty"`
}

// VolumeAttachment represents an EBS volume attachment
type VolumeAttachment struct {
	InstanceID string    `json:"instance_id"`
	Device     string    `json:"device"`
	State      string    `json:"state"` // attaching, attached, detaching, detached
	AttachTime time.Time `json:"attach_time"`
}

// S3Bucket represents an AWS S3 bucket
type S3Bucket struct {
	Name                 string    `json:"name"`
	AccountID            string    `json:"account_id"`
	AccountName          string    `json:"account_name"`
	Region               string    `json:"region"`
	CreationDate         time.Time `json:"creation_date"`
	Versioning           string    `json:"versioning"`            // Enabled, Suspended, or empty
	Encrypted            bool      `json:"encrypted"`
	PublicAccessBlocked  bool      `json:"public_access_blocked"` // True if all public access is blocked
	Tags                 []Tag     `json:"tags,omitempty"`
	Size                 int64     `json:"size,omitempty"`       // Total size in bytes (optional, can be expensive to calculate)
	ObjectCount          int64     `json:"object_count,omitempty"` // Total number of objects (optional)
	IsPublic             bool      `json:"is_public"`            // True if bucket has public access
	HasLifecyclePolicy   bool      `json:"has_lifecycle_policy"`
	HasLogging           bool      `json:"has_logging"`
}
