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

	// API routes
	api := r.Group("/api")
	{
		// Account and user management routes
		api.GET("/accounts", s.handler.ListAccounts)
		api.GET("/accounts/:accountId/users", s.handler.ListUsers)
		api.GET("/accounts/:accountId/users/:username", s.handler.GetUser)
		api.DELETE("/accounts/:accountId/users/:username", s.handler.DeleteUser)
		api.POST("/accounts/:accountId/users/:username/keys", s.handler.CreateAccessKey)
		api.DELETE("/accounts/:accountId/users/:username/keys/:keyId", s.handler.DeleteAccessKey)
		api.PUT("/accounts/:accountId/users/:username/keys/:keyId/rotate", s.handler.RotateAccessKey)
		
		// Public IP management routes
		api.GET("/public-ips", s.handler.ListPublicIPs)
		
		// Cache management routes
		api.POST("/cache/clear", s.handler.ClearCache)
		api.POST("/cache/accounts/:accountId/invalidate", s.handler.InvalidateAccountCache)
		api.POST("/cache/accounts/:accountId/users/:username/invalidate", s.handler.InvalidateUserCache)
		api.POST("/cache/public-ips/invalidate", s.handler.InvalidatePublicIPsCache)
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
