// Package handlers provides HTTP request handlers for AWS SSO (IAM Identity Center) operations
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rusik69/aws-iam-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type SSOHandler struct {
	ssoService services.SSOServiceInterface
}

func NewSSOHandler(ssoService services.SSOServiceInterface) *SSOHandler {
	return &SSOHandler{
		ssoService: ssoService,
	}
}

// GetIdentityCenterInstance gets the Identity Center instance information
func (h *SSOHandler) GetIdentityCenterInstance(c *gin.Context) {
	instance, err := h.ssoService.GetIdentityCenterInstance()
	if err != nil {
		fmt.Printf("[ERROR] GetIdentityCenterInstance failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to get Identity Center instance. Check AWS credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, instance)
}

// ListSSOUsers lists all SSO users
func (h *SSOHandler) ListSSOUsers(c *gin.Context) {
	users, err := h.ssoService.ListSSOUsers()
	if err != nil {
		fmt.Printf("[ERROR] ListSSOUsers failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list SSO users. Check AWS credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetSSOUser gets a specific SSO user
func (h *SSOHandler) GetSSOUser(c *gin.Context) {
	userID := c.Param("userId")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	user, err := h.ssoService.GetSSOUser(userID)
	if err != nil {
		fmt.Printf("[ERROR] GetSSOUser failed for user %s: %v\n", userID, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to get SSO user %s. Check if it exists and you have permissions.", userID),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListSSOGroups lists all SSO groups
func (h *SSOHandler) ListSSOGroups(c *gin.Context) {
	groups, err := h.ssoService.ListSSOGroups()
	if err != nil {
		fmt.Printf("[ERROR] ListSSOGroups failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list SSO groups. Check AWS credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// GetSSOGroup gets a specific SSO group
func (h *SSOHandler) GetSSOGroup(c *gin.Context) {
	groupID := c.Param("groupId")

	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Group ID is required",
		})
		return
	}

	group, err := h.ssoService.GetSSOGroup(groupID)
	if err != nil {
		fmt.Printf("[ERROR] GetSSOGroup failed for group %s: %v\n", groupID, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to get SSO group %s. Check if it exists and you have permissions.", groupID),
		})
		return
	}

	c.JSON(http.StatusOK, group)
}

// ListGroupMembers lists members of a group
func (h *SSOHandler) ListGroupMembers(c *gin.Context) {
	groupID := c.Param("groupId")

	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Group ID is required",
		})
		return
	}

	members, err := h.ssoService.ListGroupMembers(groupID)
	if err != nil {
		fmt.Printf("[ERROR] ListGroupMembers failed for group %s: %v\n", groupID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to list members for group %s. Check AWS credentials and permissions.", groupID),
		})
		return
	}
	c.JSON(http.StatusOK, members)
}

// ListAccountAssignments lists account assignments for a specific account
func (h *SSOHandler) ListAccountAssignments(c *gin.Context) {
	accountID := c.Param("accountId")

	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Account ID is required",
		})
		return
	}

	assignments, err := h.ssoService.ListAccountAssignments(accountID)
	if err != nil {
		fmt.Printf("[ERROR] ListAccountAssignments failed for account %s: %v\n", accountID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to list account assignments for account %s. Check AWS credentials and permissions.", accountID),
		})
		return
	}
	c.JSON(http.StatusOK, assignments)
}

// ListUserAccountAssignments lists account assignments for a specific user
func (h *SSOHandler) ListUserAccountAssignments(c *gin.Context) {
	userID := c.Param("userId")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	assignments, err := h.ssoService.ListAccountAssignmentsForUser(userID)
	if err != nil {
		fmt.Printf("[ERROR] ListUserAccountAssignments failed for user %s: %v\n", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to list account assignments for user %s. Check AWS credentials and permissions.", userID),
		})
		return
	}
	c.JSON(http.StatusOK, assignments)
}

// ListGroupAccountAssignments lists account assignments for a specific group
func (h *SSOHandler) ListGroupAccountAssignments(c *gin.Context) {
	groupID := c.Param("groupId")

	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Group ID is required",
		})
		return
	}

	assignments, err := h.ssoService.ListAccountAssignmentsForGroup(groupID)
	if err != nil {
		fmt.Printf("[ERROR] ListGroupAccountAssignments failed for group %s: %v\n", groupID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to list account assignments for group %s. Check AWS credentials and permissions.", groupID),
		})
		return
	}
	c.JSON(http.StatusOK, assignments)
}

// ListAllUserAssignments lists all users with their account assignments
func (h *SSOHandler) ListAllUserAssignments(c *gin.Context) {
	users, err := h.ssoService.ListAllUserAssignments()
	if err != nil {
		fmt.Printf("[ERROR] ListAllUserAssignments failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list all user assignments. Check AWS credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, users)
}

// ListAllGroupAssignments lists all groups with their account assignments
func (h *SSOHandler) ListAllGroupAssignments(c *gin.Context) {
	groups, err := h.ssoService.ListAllGroupAssignments()
	if err != nil {
		fmt.Printf("[ERROR] ListAllGroupAssignments failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list all group assignments. Check AWS credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// ListAllAccountAssignments lists all accounts with their SSO assignments
func (h *SSOHandler) ListAllAccountAssignments(c *gin.Context) {
	accounts, err := h.ssoService.ListAllAccountAssignments()
	if err != nil {
		fmt.Printf("[ERROR] ListAllAccountAssignments failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list all account assignments. Check AWS credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

// Cache management handlers

// ClearSSOCache clears all SSO cache
func (h *SSOHandler) ClearSSOCache(c *gin.Context) {
	h.ssoService.ClearCache()
	c.JSON(http.StatusOK, gin.H{"message": "SSO cache cleared successfully"})
}

// InvalidateSSOUsersCache invalidates SSO users cache
func (h *SSOHandler) InvalidateSSOUsersCache(c *gin.Context) {
	h.ssoService.InvalidateSSOUsersCache()
	c.JSON(http.StatusOK, gin.H{"message": "SSO users cache invalidated successfully"})
}

// InvalidateSSOGroupsCache invalidates SSO groups cache
func (h *SSOHandler) InvalidateSSOGroupsCache(c *gin.Context) {
	h.ssoService.InvalidateSSOGroupsCache()
	c.JSON(http.StatusOK, gin.H{"message": "SSO groups cache invalidated successfully"})
}

// InvalidateSSOUserCache invalidates cache for a specific SSO user
func (h *SSOHandler) InvalidateSSOUserCache(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}
	h.ssoService.InvalidateSSOUserCache(userID)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("SSO user cache invalidated for user %s", userID)})
}

// InvalidateSSOGroupCache invalidates cache for a specific SSO group
func (h *SSOHandler) InvalidateSSOGroupCache(c *gin.Context) {
	groupID := c.Param("groupId")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Group ID is required",
		})
		return
	}
	h.ssoService.InvalidateSSOGroupCache(groupID)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("SSO group cache invalidated for group %s", groupID)})
}

// InvalidateAccountAssignmentsCache invalidates cache for account assignments
func (h *SSOHandler) InvalidateAccountAssignmentsCache(c *gin.Context) {
	accountID := c.Param("accountId")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Account ID is required",
		})
		return
	}
	h.ssoService.InvalidateAccountAssignmentsCache(accountID)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Account assignments cache invalidated for account %s", accountID)})
}
