package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"github.com/SagheerAkram/QuickCmd/agent"
)

func main() {
	configPath := flag.String("config", "/etc/quickcmd/agent-config.yaml", "Path to configuration file")
	flag.Parse()
	
	// Load configuration
	config, err := agent.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Create server
	server, err := agent.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	// Start server in background
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()
	
	log.Println("QuickCMD Agent started successfully")
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	
	log.Println("Shutting down gracefully...")
	
	// Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	
	log.Println("Agent stopped")
}
