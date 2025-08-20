// Package handlers provides HTTP request handlers for the AWS IAM manager
package handlers

import (
	"log"
	"net/http"
	"time"

	"aws-iam-manager/internal/services"

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
	start := time.Now()
	log.Printf("[INFO] ListAccounts request from %s", c.ClientIP())
	
	accounts, err := h.awsService.ListAccounts()
	if err != nil {
		log.Printf("[ERROR] ListAccounts failed: %v (took %v)", err, time.Since(start))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("[INFO] ListAccounts successful: returned %d accounts (took %v)", len(accounts), time.Since(start))
	c.JSON(http.StatusOK, accounts)
}

func (h *Handler) ListUsers(c *gin.Context) {
	start := time.Now()
	accountID := c.Param("accountId")
	log.Printf("[INFO] ListUsers request for account %s from %s", accountID, c.ClientIP())

	users, err := h.awsService.ListUsers(accountID)
	if err != nil {
		log.Printf("[ERROR] ListUsers failed for account %s: %v (took %v)", accountID, err, time.Since(start))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("[INFO] ListUsers successful for account %s: returned %d users (took %v)", accountID, len(users), time.Since(start))
	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetUser(c *gin.Context) {
	start := time.Now()
	accountID := c.Param("accountId")
	username := c.Param("username")
	log.Printf("[INFO] GetUser request for user %s in account %s from %s", username, accountID, c.ClientIP())

	user, err := h.awsService.GetUser(accountID, username)
	if err != nil {
		log.Printf("[ERROR] GetUser failed for user %s in account %s: %v (took %v)", username, accountID, err, time.Since(start))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("[INFO] GetUser successful for user %s in account %s (took %v)", username, accountID, time.Since(start))
	c.JSON(http.StatusOK, user)
}

func (h *Handler) CreateAccessKey(c *gin.Context) {
	start := time.Now()
	accountID := c.Param("accountId")
	username := c.Param("username")
	log.Printf("[INFO] CreateAccessKey request for user %s in account %s from %s", username, accountID, c.ClientIP())

	response, err := h.awsService.CreateAccessKey(accountID, username)
	if err != nil {
		log.Printf("[ERROR] CreateAccessKey failed for user %s in account %s: %v (took %v)", username, accountID, err, time.Since(start))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("[INFO] CreateAccessKey successful for user %s in account %s (took %v)", username, accountID, time.Since(start))
	c.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteAccessKey(c *gin.Context) {
	start := time.Now()
	accountID := c.Param("accountId")
	username := c.Param("username")
	keyID := c.Param("keyId")
	log.Printf("[INFO] DeleteAccessKey request for key %s, user %s in account %s from %s", keyID, username, accountID, c.ClientIP())

	err := h.awsService.DeleteAccessKey(accountID, username, keyID)
	if err != nil {
		log.Printf("[ERROR] DeleteAccessKey failed for key %s, user %s in account %s: %v (took %v)", keyID, username, accountID, err, time.Since(start))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("[INFO] DeleteAccessKey successful for key %s, user %s in account %s (took %v)", keyID, username, accountID, time.Since(start))
	c.JSON(http.StatusOK, gin.H{"message": "Access key deleted successfully"})
}

func (h *Handler) RotateAccessKey(c *gin.Context) {
	start := time.Now()
	accountID := c.Param("accountId")
	username := c.Param("username")
	keyID := c.Param("keyId")
	log.Printf("[INFO] RotateAccessKey request for key %s, user %s in account %s from %s", keyID, username, accountID, c.ClientIP())

	response, err := h.awsService.RotateAccessKey(accountID, username, keyID)
	if err != nil {
		log.Printf("[ERROR] RotateAccessKey failed for key %s, user %s in account %s: %v (took %v)", keyID, username, accountID, err, time.Since(start))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("[INFO] RotateAccessKey successful for key %s, user %s in account %s (took %v)", keyID, username, accountID, time.Since(start))
	c.JSON(http.StatusOK, response)
}
