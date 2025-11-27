package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/esxi-manager/esxi-manager/internal/config"
	"github.com/esxi-manager/esxi-manager/internal/esxi/connection"
)

func main() {
	uri := flag.String("host", "192.168.159.147", "ESXi host URI")
	username := flag.String("username", "root", "ESXi username")
	password := flag.String("password", "", "ESXi password")
	port := flag.Int("port", 22, "SSH port")
	flag.Parse()

	if *password == "" {
		log.Fatal("password flag is required")
	}

	fmt.Printf("Testing connection to %s:%d as %s\n", *uri, *port, *username)

	// Create host config
	host := &config.ESXiHost{
		URI:      *uri,
		Username: *username,
		Password: *password,
		Port:     *port,
	}

	// Validate config
	if err := host.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}
	fmt.Println("✓ Configuration validated")

	// Create connection manager
	manager, err := connection.NewManager(host)
	if err != nil {
		log.Fatalf("Failed to create connection manager: %v", err)
	}
	fmt.Println("✓ Connection manager created")

	// Attempt connection
	fmt.Println("\nAttempting to connect...")
	if err := manager.Connect(); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	fmt.Println("✓ Connected successfully!")

	// Check connection status
	if manager.IsConnected() {
		fmt.Println("✓ Connection is active")
	} else {
		log.Fatal("Connection check failed")
	}

	// Close connection
	if err := manager.Close(); err != nil {
		log.Fatalf("Failed to close connection: %v", err)
	}
	fmt.Println("✓ Connection closed gracefully")

	fmt.Println("\n✅ All tests passed!")
}
