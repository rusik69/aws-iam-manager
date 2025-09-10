package main

import (
	"log"

	"github.com/rusik69/aws-iam-manager/internal/config"
	"github.com/rusik69/aws-iam-manager/internal/server"
)

func main() {
	log.Printf("[INFO] Starting AWS IAM Manager server...")
	log.Printf("[INFO] Application version: v1.0.0")

	cfg := config.LoadConfig()
	log.Printf("[INFO] Configuration loaded successfully")

	srv := server.NewServer(cfg)
	log.Printf("[INFO] Server instance created")

	if err := srv.Start(); err != nil {
		log.Fatalf("[FATAL] Failed to start server: %v", err)
	}
}