/*
 * websocket.go - Real-time communication for frontend (dashboard)
 *
 * Maintains WebSocket connections with clients and automatically broadcasts
 * fault events and system status.
 * Uses heartbeat (ping/pong) to keep connections alive.
 */
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Upgrader converts HTTP connection to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


// Map of active WebSocket connections
var (
	wsClients = make(map[*websocket.Conn]bool)
	wsMutex   sync.Mutex
)

// cleanupClient closes the connection and removes it from the map
func cleanupClient(conn *websocket.Conn) {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	if _, ok := wsClients[conn]; ok {
		conn.Close()
		delete(wsClients, conn)
		log.Println("WebSocket client disconnected")
	}
}

// broadcastMessage sends a message to ALL connected clients
func broadcastMessage(msgType string, data interface{}) {

    msg, err := json.Marshal(map[string]interface{}{
        "type": msgType,
        "data": data,
    })
    if err != nil {
        log.Println("Failed to marshal WS message:", err)
        return
    }

   	// Copy clients (short lock)
    wsMutex.Lock()
    clients := make([]*websocket.Conn, 0, len(wsClients))
    for c := range wsClients {
        clients = append(clients, c)
    }
    wsMutex.Unlock()

   	// Send messages WITHOUT holding lock
    for _, conn := range clients {
        if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
            cleanupClient(conn)
        }
    }
}


// WebSocketHandler handles new client connections
func WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	// Add client to map
	wsMutex.Lock()
	wsClients[conn] = true
	wsMutex.Unlock()
	log.Println("WebSocket client connected")
	defer cleanupClient(conn)

	//ping/pong (keep-alive)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	//ping ticker
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Gorilla WebSocket REQUIRES reading messages, otherwise browser closes connection
go func() {
    defer cleanupClient(conn)
    for {
        if _, _, err := conn.ReadMessage(); err != nil {
            return
        }
    }
}()


	//Ping loop
	for range pingTicker.C {
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}

// BroadcastFault sends a new fault event to the dashboard
func BroadcastFault(data interface{}) {
    broadcastMessage("NEW_FAULT", data)
}

// BroadcastStatus sends current MQTT / DB status
func BroadcastStatus() {
	status := map[string]string{
		"mqtt": "NOT CONNECTED",
		"db":   "ERROR",
	}
	// Check database
	if err := DB.Ping(context.Background()); err == nil {
		status["db"] = "OK"
	}
	//Check MQTT
	if mqttClient != nil && mqttClient.IsConnected() {
		status["mqtt"] = "CONNECTED"
	}
	broadcastMessage("STATUS", status)
}

// StartStatusBroadcaster sends system status every 2 seconds
func StartStatusBroadcaster() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			BroadcastStatus()
		}
	}()
}
