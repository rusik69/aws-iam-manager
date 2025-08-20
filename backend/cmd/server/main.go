package main

import (
	"log"

	"aws-iam-manager/internal/server"
	"aws-iam-manager/pkg/config"
)

func main() {
	log.Printf("[INFO] Starting AWS IAM Manager server...")
	log.Printf("[INFO] Application version: v1.0.0")
	
	cfg := config.Load()
	log.Printf("[INFO] Configuration loaded successfully")
	
	srv := server.New(cfg)
	log.Printf("[INFO] Server instance created")

	if err := srv.Start(); err != nil {
		log.Fatalf("[FATAL] Failed to start server: %v", err)
	}
}
