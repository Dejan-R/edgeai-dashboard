/*
 * api.go - Routes and authentication for EdgeAI dashboard
 *
 * Gin router with Basic Auth for the dashboard (/), while the API (/api/faults)
 * and WebSocket (/ws) are public for MCU and clients.
 * Passwords are verified using bcrypt from configuration.
 */
package main

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// BcryptAuthMiddleware implements Basic Auth with bcrypt password verification
func BcryptAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Basic ") {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatus(401)
			return
		}
		// Decode Base64 and validate credentials
		payload, err := base64.StdEncoding.DecodeString(auth[len("Basic "):])
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		parts := strings.SplitN(string(payload), ":", 2)
		if len(parts) != 2 {
			c.AbortWithStatus(401)
			return
		}
		username, password := parts[0], parts[1]
		// Validate username and password
		if username != cfg.DashboardUser {
			c.AbortWithStatus(401)
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(cfg.DashboardPassHash), []byte(password)) != nil {
			c.AbortWithStatus(401)
			return
		}
		c.Next()
	}
}

// SetupRouter configures all backend routes
func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")   // HTML templates for dashboard
	r.Static("/static", "./static") // CSS/JS files

// Protected routes (dashboard)
	authorized := r.Group("/", BcryptAuthMiddleware())
	authorized.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Public routes
	r.GET("/api/faults", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetFaultEvents())
	})

	r.GET("/ws", WebSocketHandler)
	return r
}