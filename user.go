// Package models contains data structures for AWS IAM entities
package main

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
	AccessKeyID string    `json:"access_key_id"`
	Status      string    `json:"status"`
	CreateDate  time.Time `json:"create_date"`
}
