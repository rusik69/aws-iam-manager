// Package handlers provides HTTP request handlers for Azure operations
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rusik69/aws-iam-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type AzureHandler struct {
	azureService services.AzureServiceInterface
}

func NewAzureHandler(azureService services.AzureServiceInterface) *AzureHandler {
	return &AzureHandler{
		azureService: azureService,
	}
}

// ListEnterpriseApplications lists all Azure AD enterprise applications
func (h *AzureHandler) ListEnterpriseApplications(c *gin.Context) {
	apps, err := h.azureService.ListEnterpriseApplications(c.Request.Context())
	if err != nil {
		fmt.Printf("[ERROR] ListEnterpriseApplications failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list Azure enterprise applications. Check Azure credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, apps)
}

// GetEnterpriseApplication gets a specific Azure AD enterprise application by ID
func (h *AzureHandler) GetEnterpriseApplication(c *gin.Context) {
	appID := c.Param("appId")

	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	app, err := h.azureService.GetEnterpriseApplication(c.Request.Context(), appID)
	if err != nil {
		fmt.Printf("[ERROR] GetEnterpriseApplication failed for app %s: %v\n", appID, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to get enterprise application %s. Check if it exists and you have permissions.", appID),
		})
		return
	}

	c.JSON(http.StatusOK, app)
}

// DeleteEnterpriseApplication deletes an Azure AD enterprise application
func (h *AzureHandler) DeleteEnterpriseApplication(c *gin.Context) {
	appID := c.Param("appId")

	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	err := h.azureService.DeleteEnterpriseApplication(c.Request.Context(), appID)
	if err != nil {
		fmt.Printf("[ERROR] DeleteEnterpriseApplication failed for app %s: %v\n", appID, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "Forbidden") || strings.Contains(err.Error(), "Unauthorized") {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to delete enterprise application %s. Check if it exists and you have permissions.", appID),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Enterprise application %s deleted successfully", appID),
	})
}

// ClearAzureCache clears all Azure-related cache
func (h *AzureHandler) ClearAzureCache(c *gin.Context) {
	h.azureService.ClearCache()
	c.JSON(http.StatusOK, gin.H{"message": "Azure cache cleared successfully"})
}

// InvalidateEnterpriseApplicationsCache invalidates the enterprise applications cache
func (h *AzureHandler) InvalidateEnterpriseApplicationsCache(c *gin.Context) {
	h.azureService.InvalidateEnterpriseApplicationsCache()
	c.JSON(http.StatusOK, gin.H{"message": "Enterprise applications cache invalidated successfully"})
}

// InvalidateEnterpriseApplicationCache invalidates cache for a specific enterprise application
func (h *AzureHandler) InvalidateEnterpriseApplicationCache(c *gin.Context) {
	appID := c.Param("appId")

	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	h.azureService.InvalidateEnterpriseApplicationCache(appID)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Cache invalidated for enterprise application %s", appID)})
}
