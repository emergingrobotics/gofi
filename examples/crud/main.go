package main

import (
	"context"
	"fmt"
	"log"

	"github.com/unifi-go/gofi"
	"github.com/unifi-go/gofi/types"
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

	site := "default"

	// Create a new network
	fmt.Println("Creating new network...")
	network := &types.Network{
		Name:        "IoT Network",
		Purpose:     "corporate",
		VLANEnabled: true,
		VLAN:        20,
		IPSubnet:    "192.168.20.1/24",
		DHCPDEnabled: true,
		DHCPDStart:  "192.168.20.10",
		DHCPDStop:   "192.168.20.250",
	}

	created, err := client.Networks().Create(ctx, site, network)
	if err != nil {
		log.Fatalf("Failed to create network: %v", err)
	}

	fmt.Printf("Created network: %s (ID: %s)\n", created.Name, created.ID)

	// Create a new WLAN
	fmt.Println("\nCreating new WLAN...")
	wlan := &types.WLAN{
		Name:       "Guest WiFi",
		Enabled:    true,
		Security:   "wpapsk",
		WPAMode:    "wpa2",
		WPAEnc:     "ccmp",
		Passphrase: "guestpassword123",
		// NetworkID: created.ID, // Link to network (if field exists)
		IsGuest:    true,
	}

	createdWLAN, err := client.WLANs().Create(ctx, site, wlan)
	if err != nil {
		log.Fatalf("Failed to create WLAN: %v", err)
	}

	fmt.Printf("Created WLAN: %s (ID: %s)\n", createdWLAN.Name, createdWLAN.ID)

	// Update the WLAN
	fmt.Println("\nUpdating WLAN...")
	createdWLAN.Name = "Guest WiFi (Updated)"

	updated, err := client.WLANs().Update(ctx, site, createdWLAN)
	if err != nil {
		log.Fatalf("Failed to update WLAN: %v", err)
	}

	fmt.Printf("Updated WLAN name to: %s\n", updated.Name)

	// List all WLANs
	fmt.Println("\nListing all WLANs...")
	wlans, err := client.WLANs().List(ctx, site)
	if err != nil {
		log.Fatalf("Failed to list WLANs: %v", err)
	}

	for _, w := range wlans {
		fmt.Printf("  - %s (Security: %s, Enabled: %t)\n", w.Name, w.Security, w.Enabled)
	}

	// Cleanup: Delete the WLAN and network
	fmt.Println("\nCleaning up...")

	if err := client.WLANs().Delete(ctx, site, createdWLAN.ID); err != nil {
		log.Printf("Warning: Failed to delete WLAN: %v", err)
	} else {
		fmt.Println("Deleted WLAN")
	}

	if err := client.Networks().Delete(ctx, site, created.ID); err != nil {
		log.Printf("Warning: Failed to delete network: %v", err)
	} else {
		fmt.Println("Deleted network")
	}

	fmt.Println("\nDone!")
}
