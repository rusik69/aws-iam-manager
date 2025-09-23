// Package handlers provides HTTP request handlers for the AWS IAM manager
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rusik69/aws-iam-manager/internal/middleware"
	"github.com/rusik69/aws-iam-manager/internal/services"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	awsService services.AWSServiceInterface
}

func NewHandler(awsService services.AWSServiceInterface) *Handler {
	return &Handler{
		awsService: awsService,
	}
}

func (h *Handler) ListAccounts(c *gin.Context) {
	accounts, err := h.awsService.ListAccounts()
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("[ERROR] ListAccounts failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list AWS accounts. Check AWS credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *Handler) ListUsers(c *gin.Context) {
	accountID := c.Param("accountId")
	users, err := h.awsService.ListUsers(accountID)
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("[ERROR] ListUsers failed for account %s: %v\n", accountID, err)
		// Check if it's an access issue
		if containsAccessDenied(err.Error()) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied to account",
				"details": fmt.Sprintf("Cannot access account %s. The role may not exist or trust relationship is not configured.", accountID),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to list users for account %s. Check AWS credentials and permissions.", accountID),
		})
		return
	}
	c.JSON(http.StatusOK, users)
}

// containsAccessDenied checks if an error message indicates access denied
func containsAccessDenied(errMsg string) bool {
	return 	strings.Contains(errMsg, "AccessDenied") ||
			strings.Contains(errMsg, "assume role") ||
			strings.Contains(errMsg, "not authorized")
}

func (h *Handler) GetCurrentUser(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":     user.Email,
		"username":  user.PreferredUsername,
		"groups":    user.Groups,
		"authenticated": true,
	})
}

func (h *Handler) GetUser(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	user, err := h.awsService.GetUser(accountID, username)
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("[ERROR] GetUser failed for user %s in account %s: %v\n", username, accountID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to get user %s from account %s. Check AWS credentials and permissions.", username, accountID),
		})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) CreateAccessKey(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	response, err := h.awsService.CreateAccessKey(accountID, username)
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("[ERROR] CreateAccessKey failed for user %s in account %s: %v\n", username, accountID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to create access key for user %s in account %s. Check AWS credentials and permissions.", username, accountID),
		})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteAccessKey(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	keyID := c.Param("keyId")
	err := h.awsService.DeleteAccessKey(accountID, username, keyID)
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("[ERROR] DeleteAccessKey failed for key %s, user %s in account %s: %v\n", keyID, username, accountID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to delete access key %s for user %s in account %s. Check AWS credentials and permissions.", keyID, username, accountID),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Access key deleted successfully"})
}

func (h *Handler) RotateAccessKey(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	keyID := c.Param("keyId")
	response, err := h.awsService.RotateAccessKey(accountID, username, keyID)
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("[ERROR] RotateAccessKey failed for key %s, user %s in account %s: %v\n", keyID, username, accountID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to rotate access key %s for user %s in account %s. Check AWS credentials and permissions.", keyID, username, accountID),
		})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	err := h.awsService.DeleteUser(accountID, username)
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("[ERROR] DeleteUser failed for user %s in account %s: %v\n", username, accountID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to delete user %s from account %s. Check AWS credentials and permissions.", username, accountID),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *Handler) DeleteUserPassword(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	err := h.awsService.DeleteUserPassword(accountID, username)
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("[ERROR] DeleteUserPassword failed for user %s in account %s: %v\n", username, accountID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to delete password for user %s in account %s. %s", username, accountID, err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User password deleted successfully"})
}

func (h *Handler) RotateUserPassword(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	response, err := h.awsService.RotateUserPassword(accountID, username)
	if err != nil {
		// Log the full error for debugging
		fmt.Printf("[ERROR] RotateUserPassword failed for user %s in account %s: %v\n", username, accountID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to rotate password for user %s in account %s. %s", username, accountID, err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) ListPublicIPs(c *gin.Context) {
	ips, err := h.awsService.ListPublicIPs()
	if err != nil {
		fmt.Printf("[ERROR] ListPublicIPs failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list public IP addresses. Check AWS credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, ips)
}

func (h *Handler) ClearCache(c *gin.Context) {
	h.awsService.ClearCache()
	c.JSON(http.StatusOK, gin.H{"message": "Cache cleared successfully"})
}

func (h *Handler) InvalidateAccountCache(c *gin.Context) {
	accountID := c.Param("accountId")
	h.awsService.InvalidateAccountCache(accountID)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Cache invalidated for account %s", accountID)})
}

func (h *Handler) InvalidateUserCache(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	h.awsService.InvalidateUserCache(accountID, username)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Cache invalidated for user %s in account %s", username, accountID)})
}

func (h *Handler) InvalidatePublicIPsCache(c *gin.Context) {
	h.awsService.InvalidatePublicIPsCache()
	c.JSON(http.StatusOK, gin.H{"message": "Public IPs cache invalidated successfully"})
}

func (h *Handler) ListSecurityGroups(c *gin.Context) {
	sgs, err := h.awsService.ListSecurityGroups()
	if err != nil {
		fmt.Printf("[ERROR] ListSecurityGroups failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list security groups. Check AWS credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, sgs)
}

func (h *Handler) ListSecurityGroupsByAccount(c *gin.Context) {
	accountID := c.Param("accountId")

	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Account ID is required",
		})
		return
	}

	sgs, err := h.awsService.ListSecurityGroupsByAccount(accountID)
	if err != nil {
		fmt.Printf("[ERROR] ListSecurityGroupsByAccount failed for account %s: %v\n", accountID, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "cannot access account") {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, sgs)
}

func (h *Handler) InvalidateSecurityGroupsCache(c *gin.Context) {
	h.awsService.InvalidateSecurityGroupsCache()
	c.JSON(http.StatusOK, gin.H{"message": "Security groups cache invalidated successfully"})
}

func (h *Handler) InvalidateAccountSecurityGroupsCache(c *gin.Context) {
	accountID := c.Param("accountId")

	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Account ID is required",
		})
		return
	}

	h.awsService.InvalidateAccountSecurityGroupsCache(accountID)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Security groups cache invalidated for account %s", accountID)})
}

func (h *Handler) GetSecurityGroup(c *gin.Context) {
	accountID := c.Param("accountId")
	region := c.Param("region")
	groupID := c.Param("groupId")

	if accountID == "" || region == "" || groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Account ID, region, and group ID are required",
		})
		return
	}

	sg, err := h.awsService.GetSecurityGroup(accountID, region, groupID)
	if err != nil {
		fmt.Printf("[ERROR] GetSecurityGroup failed for group %s in account %s, region %s: %v\n", groupID, accountID, region, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "cannot access account") {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, sg)
}

func (h *Handler) DeleteSecurityGroup(c *gin.Context) {
	accountID := c.Param("accountId")
	region := c.Param("region")
	groupID := c.Param("groupId")

	if accountID == "" || region == "" || groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Account ID, region, and group ID are required",
		})
		return
	}

	err := h.awsService.DeleteSecurityGroup(accountID, region, groupID)
	if err != nil {
		fmt.Printf("[ERROR] DeleteSecurityGroup failed for group %s in account %s, region %s: %v\n", groupID, accountID, region, err)

		// Determine appropriate status code based on error
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "cannot delete default") ||
			strings.Contains(err.Error(), "still in use") {
			statusCode = http.StatusConflict
		} else if strings.Contains(err.Error(), "cannot access account") {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Security group %s deleted successfully", groupID),
	})
}

