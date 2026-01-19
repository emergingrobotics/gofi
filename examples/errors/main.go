package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/unifi-go/gofi"
)

func main() {
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

	// Example 1: Connection errors
	fmt.Println("=== Example 1: Handling connection errors ===")
	if err := client.Connect(ctx); err != nil {
		// Check for specific error types
		if errors.Is(err, gofi.ErrAuthenticationFailed) {
			fmt.Println("Authentication failed - check credentials")
			return
		}

		if errors.Is(err, gofi.ErrTimeout) {
			fmt.Println("Connection timed out - check network")
			return
		}

		fmt.Printf("Connection error: %v\n", err)
		return
	}
	defer client.Disconnect(ctx)

	fmt.Println("Connected successfully!")

	// Example 2: Resource not found
	fmt.Println("\n=== Example 2: Handling not found errors ===")
	_, err = client.Devices().Get(ctx, "default", "nonexistent-id")
	if err != nil {
		if errors.Is(err, gofi.ErrNotFound) {
			fmt.Println("Device not found (expected)")
		} else {
			fmt.Printf("Unexpected error: %v\n", err)
		}
	}

	// Example 3: API errors
	fmt.Println("\n=== Example 3: Handling API errors ===")
	_, err = client.Networks().Get(ctx, "default", "invalid-id")
	if err != nil {
		var apiErr *gofi.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("API Error [%d]: %s (endpoint: %s)\n",
				apiErr.StatusCode,
				apiErr.Message,
				apiErr.Endpoint,
			)
		}
	}

	// Example 4: Validation errors
	fmt.Println("\n=== Example 4: Handling validation errors ===")
	invalidConfig := &gofi.Config{
		// Missing required fields
		Username: "admin",
	}

	_, err = gofi.New(invalidConfig)
	if err != nil {
		var valErr *gofi.ValidationError
		if errors.As(err, &valErr) {
			fmt.Printf("Validation error on field '%s': %s\n", valErr.Field, valErr.Message)
		}
	}

	// Example 5: Retry on transient failures
	fmt.Println("\n=== Example 5: Automatic retry on transient failures ===")
	retryConfig := &gofi.Config{
		Host:          "192.168.1.1",
		Username:      "admin",
		Password:      "your-password",
		SkipTLSVerify: true,
		RetryConfig: &gofi.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 100,
			MaxBackoff:     5000,
		},
	}

	retryClient, _ := gofi.New(retryConfig)
	if err := retryClient.Connect(ctx); err != nil {
		// Connection will be retried automatically on transient failures
		fmt.Printf("Connection failed after retries: %v\n", err)
	} else {
		fmt.Println("Connected with retry configuration")
		retryClient.Disconnect(ctx)
	}

	fmt.Println("\n=== Error handling examples complete ===")
}
