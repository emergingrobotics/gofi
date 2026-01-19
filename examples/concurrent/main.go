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

	// List all devices to get their IDs
	devices, err := client.Devices().List(ctx, site)
	if err != nil {
		log.Fatalf("Failed to list devices: %v", err)
	}

	if len(devices) == 0 {
		log.Fatal("No devices found for batch operations demo")
	}

	// Extract device IDs
	deviceIDs := make([]string, 0, len(devices))
	for _, device := range devices {
		deviceIDs = append(deviceIDs, device.ID)
	}

	fmt.Printf("Fetching %d devices concurrently...\n", len(deviceIDs))

	// Batch get devices
	results := gofi.BatchGet(ctx, deviceIDs, func(ctx context.Context, id string) (*types.Device, error) {
		return client.Devices().Get(ctx, site, id)
	})

	// Process results
	successCount := 0
	errorCount := 0

	for _, result := range results {
		if result.Error != nil {
			errorCount++
			fmt.Printf("  [ERROR] Index %d: %v\n", result.Index, result.Error)
		} else {
			successCount++
			fmt.Printf("  [OK] %s (%s)\n", result.Item.Name, result.Item.MAC)
		}
	}

	fmt.Printf("\nBatch operation complete: %d successful, %d errors\n", successCount, errorCount)

	// Demonstrate concurrent device commands
	if len(devices) > 0 && devices[0].MAC != "" {
		fmt.Println("\nExample: Concurrent device operations (commented out for safety)")
		fmt.Println("// To locate multiple devices:")
		fmt.Println("// macs := []string{\"aa:bb:cc:dd:ee:f1\", \"aa:bb:cc:dd:ee:f2\"}")
		fmt.Println("// errors := gofi.BatchDelete(ctx, macs, func(ctx context.Context, mac string) error {")
		fmt.Println("//     return client.Devices().Locate(ctx, site, mac)")
		fmt.Println("// })")
	}
}
