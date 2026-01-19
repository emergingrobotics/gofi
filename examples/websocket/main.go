package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/unifi-go/gofi"
)

func main() {
	// Create client
	config := &gofi.Config{
		Host:          "192.168.1.1",
		Username:      "admin",
		Password:      "your-password",
		SkipTLSVerify: true,
	}

	client, err := gofi.New(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Connected! Subscribing to events...")

	// Subscribe to events
	eventCh, errorCh, err := client.Events().Subscribe(ctx, "default")
	if err != nil {
		log.Fatalf("Failed to subscribe to events: %v", err)
	}
	defer client.Events().Close()

	fmt.Println("Listening for events... Press Ctrl+C to exit")

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Process events
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				fmt.Println("Event channel closed")
				return
			}

			fmt.Printf("[EVENT] %s: %s\n", event.Key, event.Message)

			// Print additional details based on event type
			switch event.Key {
			case "EVT_WU_Connected", "EVT_WU_Disconnected":
				fmt.Printf("        Client: %s, SSID: %s\n", event.Client, event.SSID)
			case "EVT_AP_Connected", "EVT_AP_Disconnected":
				fmt.Printf("        AP: %s (%s)\n", event.APName, event.APMAC)
			}

		case err, ok := <-errorCh:
			if !ok {
				fmt.Println("Error channel closed")
				return
			}

			fmt.Printf("[ERROR] %v\n", err)

		case <-sigCh:
			fmt.Println("\nShutting down gracefully...")
			return
		}
	}
}
