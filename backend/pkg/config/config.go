// Package config provides application configuration management
package config

import "os"

// Config holds application configuration
type Config struct {
	Port      string
	AWSRegion string
	RoleName  string
}

// Load creates and returns application configuration from environment variables
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	roleName := os.Getenv("IAM_ROLE_NAME")
	if roleName == "" {
		roleName = "OrganizationAccountAccessRole"
	}

	return &Config{
		Port:      port,
		AWSRegion: region,
		RoleName:  roleName,
	}
}
