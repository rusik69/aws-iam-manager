package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Test default values
	os.Unsetenv("PORT")
	os.Unsetenv("AWS_REGION")

	cfg := Load()
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "us-east-1", cfg.AWSRegion)

	// Test custom values
	os.Setenv("PORT", "9000")
	os.Setenv("AWS_REGION", "us-west-2")

	cfg = Load()
	assert.Equal(t, "9000", cfg.Port)
	assert.Equal(t, "us-west-2", cfg.AWSRegion)

	// Clean up
	os.Unsetenv("PORT")
	os.Unsetenv("AWS_REGION")
}
