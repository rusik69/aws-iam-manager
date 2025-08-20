// Package server provides the HTTP server and routing configuration
package server

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"

	"aws-iam-manager/internal/handlers"
	"aws-iam-manager/internal/services"
	"aws-iam-manager/pkg/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

type Server struct {
	config  *config.Config
	handler *handlers.Handler
}

func New(cfg *config.Config) *Server {
	awsService := services.NewAWSService(cfg)
	handler := handlers.NewHandler(awsService)

	return &Server{
		config:  cfg,
		handler: handler,
	}
}

func (s *Server) SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Custom logging middleware
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s \"%s %s %s\" %d %s \"%s\" \"%s\"\n",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"*"}
	r.Use(cors.New(corsConfig))

	// API routes
	api := r.Group("/api")
	{
		// Account and user management routes
		api.GET("/accounts", s.handler.ListAccounts)
		api.GET("/accounts/:accountId/users", s.handler.ListUsers)
		api.GET("/accounts/:accountId/users/:username", s.handler.GetUser)
		api.POST("/accounts/:accountId/users/:username/keys", s.handler.CreateAccessKey)
		api.DELETE("/accounts/:accountId/users/:username/keys/:keyId", s.handler.DeleteAccessKey)
		api.PUT("/accounts/:accountId/users/:username/keys/:keyId/rotate", s.handler.RotateAccessKey)
	}

	// Serve frontend
	s.setupFrontendRoutes(r)

	return r
}

func (s *Server) setupFrontendRoutes(r *gin.Engine) {
	frontendSubFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Printf("Warning: Failed to create frontend sub filesystem: %v", err)
		return
	}

	// Serve assets manually
	r.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		if filepath[0] == '/' {
			filepath = filepath[1:] // Remove leading slash
		}
		filepath = "assets/" + filepath
		
		file, err := frontendSubFS.Open(filepath)
		if err != nil {
			c.String(404, "File not found: %s", filepath)
			return
		}
		defer file.Close()
		
		data, err := io.ReadAll(file)
		if err != nil {
			c.String(500, "Error reading file: %v", err)
			return
		}
		
		// Set appropriate content type based on file extension
		if len(filepath) > 3 && filepath[len(filepath)-3:] == ".js" {
			c.Data(200, "application/javascript", data)
		} else if len(filepath) > 4 && filepath[len(filepath)-4:] == ".css" {
			c.Data(200, "text/css", data)
		} else {
			c.Data(200, "application/octet-stream", data)
		}
	})

	// Handle root route
	r.GET("/", func(c *gin.Context) {
		file, err := frontendSubFS.Open("index.html")
		if err != nil {
			c.String(500, "Error opening index.html: %v", err)
			return
		}
		defer file.Close()
		
		data, err := io.ReadAll(file)
		if err != nil {
			c.String(500, "Error reading index.html: %v", err)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	// Handle SPA routes (catch-all for non-API routes)
	r.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}
		// Serve index.html for all other routes (SPA routing)
		file, err := frontendSubFS.Open("index.html")
		if err != nil {
			c.String(500, "Error opening index.html: %v", err)
			return
		}
		defer file.Close()
		
		data, err := io.ReadAll(file)
		if err != nil {
			c.String(500, "Error reading index.html: %v", err)
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})
}

func (s *Server) Start() error {
	r := s.SetupRoutes()

	log.Printf("[INFO] Server starting on port %s", s.config.Port)
	fmt.Printf("Server starting on port %s\n", s.config.Port)
	return r.Run(":" + s.config.Port)
}
