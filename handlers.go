// Package handlers provides HTTP request handlers for the AWS IAM manager
package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	awsService AWSServiceInterface
}

func NewHandler(awsService AWSServiceInterface) *Handler {
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
