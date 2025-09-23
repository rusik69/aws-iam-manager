package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rusik69/aws-iam-manager/internal/config"
	"github.com/rusik69/aws-iam-manager/internal/handlers"
	"github.com/rusik69/aws-iam-manager/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Frontend files are served from filesystem
// TODO: Add embed support for production builds

type Server struct {
	config  config.Config
	handler *handlers.Handler
}

func NewServer(cfg config.Config) *Server {
	awsService := services.NewAWSService(cfg)
	handler := handlers.NewHandler(awsService)

	return &Server{
		config:  cfg,
		handler: handler,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())

	// Health check endpoints
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	r.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Authentication routes
		api.GET("/auth/user", s.handler.GetCurrentUser)

		// Account and user management routes
		api.GET("/accounts", s.handler.ListAccounts)
		api.GET("/accounts/:accountId/users", s.handler.ListUsers)
		api.GET("/accounts/:accountId/users/:username", s.handler.GetUser)
		api.DELETE("/accounts/:accountId/users/:username", s.handler.DeleteUser)
		api.DELETE("/accounts/:accountId/users/:username/password", s.handler.DeleteUserPassword)
		api.POST("/accounts/:accountId/users/:username/password/rotate", s.handler.RotateUserPassword)
		api.POST("/accounts/:accountId/users/:username/keys", s.handler.CreateAccessKey)
		api.DELETE("/accounts/:accountId/users/:username/keys/:keyId", s.handler.DeleteAccessKey)
		api.PUT("/accounts/:accountId/users/:username/keys/:keyId/rotate", s.handler.RotateAccessKey)
		
		// IP management routes
		api.GET("/public-ips", s.handler.ListPublicIPs)

		// Security groups routes
		api.GET("/security-groups", s.handler.ListSecurityGroups)
		api.GET("/accounts/:accountId/security-groups", s.handler.ListSecurityGroupsByAccount)
		api.GET("/accounts/:accountId/regions/:region/security-groups/:groupId", s.handler.GetSecurityGroup)
		api.DELETE("/accounts/:accountId/regions/:region/security-groups/:groupId", s.handler.DeleteSecurityGroup)

		// Cache management routes
		api.POST("/cache/clear", s.handler.ClearCache)
		api.POST("/cache/accounts/:accountId/invalidate", s.handler.InvalidateAccountCache)
		api.POST("/cache/accounts/:accountId/users/:username/invalidate", s.handler.InvalidateUserCache)
		api.POST("/cache/public-ips/invalidate", s.handler.InvalidatePublicIPsCache)
		api.POST("/cache/security-groups/invalidate", s.handler.InvalidateSecurityGroupsCache)
		api.POST("/cache/accounts/:accountId/security-groups/invalidate", s.handler.InvalidateAccountSecurityGroupsCache)
	}

	// Serve frontend
	s.setupFrontendRoutes(r)

	return r
}

func (s *Server) setupFrontendRoutes(r *gin.Engine) {
	log.Printf("[INFO] Serving frontend files from ./frontend/dist/")

	// Serve static assets
	r.Static("/assets", "./frontend/dist/assets")

	// Serve index.html for root and SPA routes
	indexHandler := func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	}

	r.GET("/", indexHandler)
	r.NoRoute(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		indexHandler(c)
	})
}

func (s *Server) Start() error {
	r := s.SetupRoutes()

	log.Printf("[INFO] Server starting on port %s", s.config.Port)
	fmt.Printf("Server starting on port %s\n", s.config.Port)
	return r.Run(":" + s.config.Port)
}
