package services

import (
	"testing"
	"time"

	"aws-iam-manager/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestModelsCreation(t *testing.T) {
	// Test Account struct
	account := models.Account{
		ID:   "123456789012",
		Name: "Test Account",
	}
	assert.Equal(t, "123456789012", account.ID)
	assert.Equal(t, "Test Account", account.Name)

	// Test User struct
	createDate := time.Now()
	user := models.User{
		Username:    "testuser",
		UserID:      "AIDACKCEVSQ6C2EXAMPLE",
		Arn:         "arn:aws:iam::123456789012:user/testuser",
		CreateDate:  createDate,
		PasswordSet: true,
		AccessKeys: []models.AccessKey{
			{
				AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
				Status:      "Active",
				CreateDate:  createDate,
			},
		},
	}

	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "AIDACKCEVSQ6C2EXAMPLE", user.UserID)
	assert.Equal(t, "arn:aws:iam::123456789012:user/testuser", user.Arn)
	assert.Equal(t, createDate, user.CreateDate)
	assert.True(t, user.PasswordSet)
	assert.Len(t, user.AccessKeys, 1)

	// Test AccessKey struct
	accessKey := models.AccessKey{
		AccessKeyID: "AKIAIOSFODNN7EXAMPLE",
		Status:      "Active",
		CreateDate:  createDate,
	}

	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", accessKey.AccessKeyID)
	assert.Equal(t, "Active", accessKey.Status)
	assert.Equal(t, createDate, accessKey.CreateDate)
}
