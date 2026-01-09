// Package handlers provides HTTP request handlers for Azure Resource Manager operations
package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rusik69/aws-iam-manager/internal/services"
	"github.com/gin-gonic/gin"
)

type AzureRMHandler struct {
	azureRMService services.AzureRMServiceInterface
}

func NewAzureRMHandler(azureRMService services.AzureRMServiceInterface) *AzureRMHandler {
	return &AzureRMHandler{
		azureRMService: azureRMService,
	}
}

// ListSubscriptions lists all Azure subscriptions
func (h *AzureRMHandler) ListSubscriptions(c *gin.Context) {
	subscriptions, err := h.azureRMService.ListSubscriptions(c.Request.Context())
	if err != nil {
		fmt.Printf("[ERROR] ListSubscriptions failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list Azure subscriptions. Check Azure credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, subscriptions)
}

// ListVMs lists all Azure virtual machines
func (h *AzureRMHandler) ListVMs(c *gin.Context) {
	subscriptionID := c.Query("subscription_id")
	vms, err := h.azureRMService.ListVMs(c.Request.Context(), subscriptionID)
	if err != nil {
		fmt.Printf("[ERROR] ListVMs failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list Azure virtual machines. Check Azure credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, vms)
}

// GetVM gets a specific Azure virtual machine
func (h *AzureRMHandler) GetVM(c *gin.Context) {
	subscriptionID := c.Param("subscriptionId")
	resourceGroup := c.Param("resourceGroup")
	vmName := c.Param("vmName")

	if subscriptionID == "" || resourceGroup == "" || vmName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Subscription ID, resource group, and VM name are required",
		})
		return
	}

	vm, err := h.azureRMService.GetVM(c.Request.Context(), subscriptionID, resourceGroup, vmName)
	if err != nil {
		fmt.Printf("[ERROR] GetVM failed for %s/%s/%s: %v\n", subscriptionID, resourceGroup, vmName, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to get virtual machine %s/%s/%s. Check if it exists and you have permissions.", subscriptionID, resourceGroup, vmName),
		})
		return
	}

	c.JSON(http.StatusOK, vm)
}

// StartVM starts an Azure virtual machine
func (h *AzureRMHandler) StartVM(c *gin.Context) {
	subscriptionID := c.Param("subscriptionId")
	resourceGroup := c.Param("resourceGroup")
	vmName := c.Param("vmName")

	if subscriptionID == "" || resourceGroup == "" || vmName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Subscription ID, resource group, and VM name are required",
		})
		return
	}

	err := h.azureRMService.StartVM(c.Request.Context(), subscriptionID, resourceGroup, vmName)
	if err != nil {
		fmt.Printf("[ERROR] StartVM failed for %s/%s: %v\n", resourceGroup, vmName, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "Forbidden") || strings.Contains(err.Error(), "Unauthorized") {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to start virtual machine %s/%s. Check if it exists and you have permissions.", resourceGroup, vmName),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Virtual machine %s/%s started successfully", resourceGroup, vmName),
	})
}

// StopVM stops an Azure virtual machine
func (h *AzureRMHandler) StopVM(c *gin.Context) {
	subscriptionID := c.Param("subscriptionId")
	resourceGroup := c.Param("resourceGroup")
	vmName := c.Param("vmName")

	if subscriptionID == "" || resourceGroup == "" || vmName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Subscription ID, resource group, and VM name are required",
		})
		return
	}

	err := h.azureRMService.StopVM(c.Request.Context(), subscriptionID, resourceGroup, vmName)
	if err != nil {
		fmt.Printf("[ERROR] StopVM failed for %s/%s: %v\n", resourceGroup, vmName, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "Forbidden") || strings.Contains(err.Error(), "Unauthorized") {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to stop virtual machine %s/%s. Check if it exists and you have permissions.", resourceGroup, vmName),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Virtual machine %s/%s stopped successfully", resourceGroup, vmName),
	})
}

// DeleteVM deletes an Azure virtual machine
func (h *AzureRMHandler) DeleteVM(c *gin.Context) {
	subscriptionID := c.Param("subscriptionId")
	resourceGroup := c.Param("resourceGroup")
	vmName := c.Param("vmName")

	if subscriptionID == "" || resourceGroup == "" || vmName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Subscription ID, resource group, and VM name are required",
		})
		return
	}

	err := h.azureRMService.DeleteVM(c.Request.Context(), subscriptionID, resourceGroup, vmName)
	if err != nil {
		fmt.Printf("[ERROR] DeleteVM failed for %s/%s: %v\n", resourceGroup, vmName, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "Forbidden") || strings.Contains(err.Error(), "Unauthorized") {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to delete virtual machine %s/%s. Check if it exists and you have permissions.", resourceGroup, vmName),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Virtual machine %s/%s deleted successfully", resourceGroup, vmName),
	})
}

// ListStorageAccounts lists all Azure storage accounts
func (h *AzureRMHandler) ListStorageAccounts(c *gin.Context) {
	subscriptionID := c.Query("subscription_id")
	accounts, err := h.azureRMService.ListStorageAccounts(c.Request.Context(), subscriptionID)
	if err != nil {
		fmt.Printf("[ERROR] ListStorageAccounts failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"details": "Failed to list Azure storage accounts. Check Azure credentials and permissions.",
		})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

// GetStorageAccount gets a specific Azure storage account
func (h *AzureRMHandler) GetStorageAccount(c *gin.Context) {
	subscriptionID := c.Param("subscriptionId")
	resourceGroup := c.Param("resourceGroup")
	name := c.Param("name")

	if subscriptionID == "" || resourceGroup == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Subscription ID, resource group, and storage account name are required",
		})
		return
	}

	account, err := h.azureRMService.GetStorageAccount(c.Request.Context(), subscriptionID, resourceGroup, name)
	if err != nil {
		fmt.Printf("[ERROR] GetStorageAccount failed for %s/%s: %v\n", resourceGroup, name, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to get storage account %s/%s. Check if it exists and you have permissions.", resourceGroup, name),
		})
		return
	}

	c.JSON(http.StatusOK, account)
}

// DeleteStorageAccount deletes an Azure storage account
func (h *AzureRMHandler) DeleteStorageAccount(c *gin.Context) {
	subscriptionID := c.Param("subscriptionId")
	resourceGroup := c.Param("resourceGroup")
	name := c.Param("name")

	if subscriptionID == "" || resourceGroup == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Subscription ID, resource group, and storage account name are required",
		})
		return
	}

	err := h.azureRMService.DeleteStorageAccount(c.Request.Context(), subscriptionID, resourceGroup, name)
	if err != nil {
		fmt.Printf("[ERROR] DeleteStorageAccount failed for %s/%s: %v\n", resourceGroup, name, err)

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "Forbidden") || strings.Contains(err.Error(), "Unauthorized") {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"details": fmt.Sprintf("Failed to delete storage account %s/%s. Check if it exists and you have permissions.", resourceGroup, name),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Storage account %s/%s deleted successfully", resourceGroup, name),
	})
}

// ClearAzureRMCache clears all Azure Resource Manager cache
func (h *AzureRMHandler) ClearAzureRMCache(c *gin.Context) {
	h.azureRMService.ClearCache()
	c.JSON(http.StatusOK, gin.H{"message": "Azure Resource Manager cache cleared successfully"})
}

// InvalidateVMsCache invalidates the VMs cache
func (h *AzureRMHandler) InvalidateVMsCache(c *gin.Context) {
	h.azureRMService.InvalidateVMsCache()
	c.JSON(http.StatusOK, gin.H{"message": "VMs cache invalidated successfully"})
}

// InvalidateStorageCache invalidates the storage accounts cache
func (h *AzureRMHandler) InvalidateStorageCache(c *gin.Context) {
	h.azureRMService.InvalidateStorageCache()
	c.JSON(http.StatusOK, gin.H{"message": "Storage accounts cache invalidated successfully"})
}
