// Package handlers provides HTTP request handlers for the AWS IAM manager
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"aws-iam-manager/internal/services"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *Handler) ListUsers(c *gin.Context) {
	accountID := c.Param("accountId")
	users, err := h.awsService.ListUsers(accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetUser(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	user, err := h.awsService.GetUser(accountID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) CreateAccessKey(c *gin.Context) {
	accountID := c.Param("accountId")
	username := c.Param("username")
	response, err := h.awsService.CreateAccessKey(accountID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}
