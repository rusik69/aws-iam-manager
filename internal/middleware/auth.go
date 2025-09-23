package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// OAuth2User represents the authenticated user information from OAuth2 Proxy
type OAuth2User struct {
	Email             string
	PreferredUsername string
	Groups            []string
	AccessToken       string
}

// OAuth2Middleware handles authentication via OAuth2 Proxy headers
func OAuth2Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for health check endpoints
		if isHealthCheckEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Check if user is authenticated by OAuth2 Proxy
		email := c.GetHeader("X-Auth-Request-Email")
		if email == "" {
			log.Printf("[WARNING] Missing X-Auth-Request-Email header for path: %s", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"message": "Please authenticate through OAuth2 proxy",
			})
			c.Abort()
			return
		}

		// Extract user information from headers
		user := OAuth2User{
			Email:             email,
			PreferredUsername: c.GetHeader("X-Auth-Request-Preferred-Username"),
			AccessToken:       c.GetHeader("X-Auth-Request-Access-Token"),
		}

		// Parse groups if available
		groupsHeader := c.GetHeader("X-Auth-Request-Groups")
		if groupsHeader != "" {
			user.Groups = strings.Split(groupsHeader, ",")
			// Trim whitespace from group names
			for i, group := range user.Groups {
				user.Groups[i] = strings.TrimSpace(group)
			}
		}

		// Log authentication for audit purposes
		log.Printf("[INFO] Authenticated user: %s (username: %s) accessing: %s %s",
			user.Email, user.PreferredUsername, c.Request.Method, c.Request.URL.Path)

		// Store user information in context for use by handlers
		c.Set("oauth2_user", user)
		c.Set("user_email", user.Email)
		c.Set("user_username", user.PreferredUsername)
		c.Set("user_groups", user.Groups)

		c.Next()
	}
}

// isHealthCheckEndpoint checks if the path is a health check endpoint that should skip auth
func isHealthCheckEndpoint(path string) bool {
	healthEndpoints := []string{
		"/ping",
		"/health",
		"/ready",
		"/healthz",
		"/livez",
		"/oauth2/callback",
		"/oauth2/start",
		"/oauth2/sign_in",
		"/oauth2/sign_out",
		"/oauth2/auth",
	}

	for _, endpoint := range healthEndpoints {
		if path == endpoint || strings.HasPrefix(path, "/oauth2/") {
			return true
		}
	}

	return false
}

// GetCurrentUser retrieves the current authenticated user from the Gin context
func GetCurrentUser(c *gin.Context) (*OAuth2User, bool) {
	user, exists := c.Get("oauth2_user")
	if !exists {
		return nil, false
	}

	oauth2User, ok := user.(OAuth2User)
	return &oauth2User, ok
}

// RequireGroup middleware ensures the user belongs to at least one of the specified groups
func RequireGroup(allowedGroups ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetCurrentUser(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check if user belongs to any of the allowed groups
		for _, userGroup := range user.Groups {
			for _, allowedGroup := range allowedGroups {
				if userGroup == allowedGroup {
					c.Next()
					return
				}
			}
		}

		log.Printf("[WARNING] User %s denied access - not in required groups: %v (user groups: %v)",
			user.Email, allowedGroups, user.Groups)

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
			"message": "You do not have permission to access this resource",
		})
		c.Abort()
	}
}

// RequireEmail middleware ensures the user's email is in the allowed list
func RequireEmail(allowedEmails ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetCurrentUser(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		// Check if user email is in the allowed list
		for _, allowedEmail := range allowedEmails {
			if user.Email == allowedEmail {
				c.Next()
				return
			}
		}

		log.Printf("[WARNING] User %s denied access - email not in allowed list: %v",
			user.Email, allowedEmails)

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
			"message": "Your email address is not authorized to access this resource",
		})
		c.Abort()
	}
}