// Package gofi provides a Go client library for programmatic control of UniFi UDM Pro devices.
//
// This library focuses on local control operations for Ubiquiti UniFi UDM Pro devices running
// UniFi OS 4.x/5.x with Network Application 10.x+.
//
// Basic usage:
//
//	config := &gofi.Config{
//		Host:     "192.168.1.1",
//		Username: "admin",
//		Password: "password",
//	}
//
//	client, err := gofi.New(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if err := client.Connect(context.Background()); err != nil {
//		log.Fatal(err)
//	}
//	defer client.Disconnect(context.Background())
//
//	// List devices
//	devices, err := client.Devices().List(context.Background(), "default")
//	if err != nil {
//		log.Fatal(err)
//	}
package gofi
