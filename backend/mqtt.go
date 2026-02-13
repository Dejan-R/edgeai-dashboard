/*
 * mqtt.go - MQTT client for communication with nRF5340DK
 *
 * Receives sensor data (vibrations) and sends it to processing via the faultChan buffer.
 */

package main

import (
	"encoding/json"
	"log"
	"time"
	"strings"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// faultChan buffer for spikes received from MQTT.
// If full, messages are simply skipped.
var faultChan = make(chan FaultPayload, 100)

// FaultPayload - data structure received from MCU
type FaultPayload struct {
	Score     float64 `json:"score"`     // anomaly distance
	Threshold float64 `json:"threshold"` // 
	Status    string  `json:"status"`    // "OK" or "FAULT"
}

// faultWorker processes messages from the channel
func faultWorker() {
    for payload := range faultChan {
        if strings.ToUpper(payload.Status) != "FAULT" {
            continue
        }
        event, err := InsertFaultEvent(payload.Score, payload.Threshold, payload.Status)
        if err != nil {
            event.Timestamp = time.Now()
            event.Score = payload.Score
            event.Threshold = payload.Threshold
            event.Status = payload.Status
        }
        wsData := map[string]interface{}{
            "timestamp": event.Timestamp.Format(time.RFC3339),
            "score":     event.Score,
            "threshold": event.Threshold,
            "status":    event.Status,
        }
        // Send via WebSocket
        BroadcastFault(wsData)
    }
}


// StartMQTT connects to the broker and receives data from nRF5340
func StartMQTT() mqtt.Client {
		// Start worker goroutine in the background
	go faultWorker()

	// MQTT configuration
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID(cfg.MQTTClient)

	// authentication
	if cfg.MQTTUsername != "" && cfg.MQTTPassword != "" {
		opts.SetUsername(cfg.MQTTUsername)
		opts.SetPassword(cfg.MQTTPassword)
	}
	client := mqtt.NewClient(opts)

	// Attempt to connect to broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("MQTT connection error:", token.Error())
	}

	log.Println("Connected to MQTT broker")

// Subscribe to topic to receive messages from MCU
	// QoS = 1
	client.Subscribe(cfg.MQTTTopic, 1, func(c mqtt.Client, msg mqtt.Message) {
			var payload FaultPayload

		//Extract  JSON 
		if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
			log.Println("MQTT JSON error:", err)
			return
		}
		// Asynchronously forward into processing pipeline
		// Select pattern prevents blocking if buffer is full
		select {
		case faultChan <- payload:
				// Normal flow - message added to buffer
		default:
			log.Println("Warning: faultChan buffer full, skipped MQTT message")
		}
	})

	return client
}
