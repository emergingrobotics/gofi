package main

import (
	"context"
	"fmt"
	"log"

	"github.com/unifi-go/gofi"
)

func main() {
	// Create client configuration
	config := &gofi.Config{
		Host:     "192.168.1.1", // Your UDM Pro IP
		Port:     443,
		Username: "admin",
		Password: "your-password",

		// For self-signed certificates (dev/testing only)
		SkipTLSVerify: true,
	}

	// Create client
	client, err := gofi.New(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Connect to controller
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("Connected to UniFi controller!")

	// List all sites
	sites, err := client.Sites().List(ctx)
	if err != nil {
		log.Fatalf("Failed to list sites: %v", err)
	}

	fmt.Printf("\nFound %d site(s):\n", len(sites))
	for _, site := range sites {
		fmt.Printf("  - %s (%s)\n", site.Desc, site.Name)
	}

	// List devices on default site
	devices, err := client.Devices().List(ctx, "default")
	if err != nil {
		log.Fatalf("Failed to list devices: %v", err)
	}

	fmt.Printf("\nFound %d device(s):\n", len(devices))
	for _, device := range devices {
		fmt.Printf("  - %s (%s) - %s - State: %s\n",
			device.Name,
			device.Model,
			device.MAC,
			device.State.String(),
		)
	}

	// List networks
	networks, err := client.Networks().List(ctx, "default")
	if err != nil {
		log.Fatalf("Failed to list networks: %v", err)
	}

	fmt.Printf("\nFound %d network(s):\n", len(networks))
	for _, network := range networks {
		fmt.Printf("  - %s (VLAN Enabled: %t)\n", network.Name, network.VLANEnabled)
	}

	// Get health information
	health, err := client.Sites().Health(ctx, "default")
	if err != nil {
		log.Fatalf("Failed to get health: %v", err)
	}

	fmt.Printf("\nHealth Status:\n")
	for _, h := range health {
		fmt.Printf("  - %s: %s\n", h.Subsystem, h.Status)
	}
}
