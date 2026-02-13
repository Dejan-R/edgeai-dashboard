/*
 *	main.go - main code
 *
 *	Service management: DB, MQTT, WebSocket, HTTP
 *  Implementation of graceful shutdown
 *	Resource lifecycle management
 */
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Global - shared resource for MQTT operations
var mqttClient mqtt.Client

func main() {
	InitDB()         // Initialize connection to TimescaleDB
	defer DB.Close() // Ensure the database is closed on program exit

	log.Println("Database initialized, WebSocket service ready")

	// Start MQTT connection and background worker
	mqttClient = StartMQTT()
	defer mqttClient.Disconnect(250)

		// Start goroutine that automatically sends status to clients
	StartStatusBroadcaster()

	// Start HTTP server in a goroutine (non-blocking)
	router := SetupRouter() // Set up Gin router with endpoints
	go func() {
		log.Printf("HTTP server running on %s\n", cfg.ServerAddr)
		if err := router.Run(cfg.ServerAddr); err != nil {
			log.Fatal("HTTP server failed:", err)
		}
	}()

	// Graceful shutdown - closes faultChan (stops workers), everything else via defer (MQTT, DB).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Initiating graceful shutdown...")
		// Critical - close faultChan first so workers know to stop
	close(faultChan)
	log.Println("Shutdown complete")
}
