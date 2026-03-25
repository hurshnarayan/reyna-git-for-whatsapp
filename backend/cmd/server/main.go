package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/reyna-bot/reyna-backend/internal/api"
	"github.com/reyna-bot/reyna-backend/internal/config"
	"github.com/reyna-bot/reyna-backend/internal/db"
	"github.com/reyna-bot/reyna-backend/internal/gdrive"
)

func main() {
	cfg := config.Load()

	// Initialize database
	store, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	// Initialize Drive service
	drive := gdrive.New(cfg.GoogleClientID, cfg.GoogleSecret, cfg.GoogleRedirectURL, "./drive_storage")

	// Initialize API server
	server := api.NewServer(cfg, store, drive)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🤖 Reyna backend starting on %s", addr)
	log.Printf("   Frontend URL: %s", cfg.FrontendURL)
	log.Printf("   Database: %s", cfg.DatabaseURL)
	log.Printf("   WhatsApp mode: %s", cfg.WhatsAppMode)

	// Graceful shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("Shutting down...")
		store.Close()
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

