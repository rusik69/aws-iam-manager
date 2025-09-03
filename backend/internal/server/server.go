// Package server provides the HTTP server and routing configuration
package server

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"aws-iam-manager/internal/handlers"
	"aws-iam-manager/internal/services"
	"aws-iam-manager/pkg/config"
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
	r.Use(cors.Default())

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

	r.StaticFS("/assets", http.FS(frontendSubFS))

	// Serve index.html for root and SPA routes
	indexHandler := func(c *gin.Context) {
		file, err := frontendSubFS.Open("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error opening index.html: %v", err)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading index.html: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
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
