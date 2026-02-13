/*
 * db.go - TimescaleDB operations for EdgeAI system
 *
 * Stores and retrieves fault data.
 * Uses pgx connection pool for efficient connection management.
 */
package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB connection pool
var DB *pgxpool.Pool

// FaultEvent represents a fault or heartbeat
type FaultEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Score     float64   `json:"score"`     // anomaly distance
	Threshold float64   `json:"threshold"` 
	Status    string    `json:"status"`    // "OK" or "FAULT"
}

// InitDB connects to TimescaleDB and tests the connection
func InitDB() {
	var err error
	DB, err = pgxpool.New(context.Background(), cfg.PostgresURL)
	if err != nil {
		log.Fatal("DB connection error:", err)
	}
	if err = DB.Ping(context.Background()); err != nil {
		log.Fatal("DB ping error:", err)
	}
	log.Println("Database connected successfully")
}

// InsertFaultEvent saves a new fault or heartbeat (heartbeat not implemented)
func InsertFaultEvent(score, threshold float64, status string) (FaultEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var event FaultEvent
	err := DB.QueryRow(ctx,
		`INSERT INTO fault_events (score, threshold, status)
         VALUES ($1, $2, $3)
         RETURNING timestamp`,
		score, threshold, status,
	).Scan(&event.Timestamp)
	if err != nil {
		log.Println("DB insert error:", err)
		event.Timestamp = time.Now()
		return event, err
	}
	event.Score = score
	event.Threshold = threshold
	event.Status = status
	return event, nil
}

// GetFaultEvents returns the last 50 records for the dashboard
func GetFaultEvents() []map[string]interface{} {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    rows, err := DB.Query(ctx,
        `SELECT timestamp, score, threshold, status
         FROM fault_events
         ORDER BY timestamp DESC
         LIMIT 50`,
    )
    if err != nil {
        log.Println("DB query error:", err)
        return nil
    }
    defer rows.Close()

    events := make([]map[string]interface{}, 0)
    for rows.Next() {
        var ts time.Time
        var score, threshold float64
        var status string
        if err := rows.Scan(&ts, &score, &threshold, &status); err != nil {
            log.Println("DB scan error:", err)
            continue
        }
        events = append(events, map[string]interface{}{
            "timestamp": ts.UTC().Format(time.RFC3339), 
            "score":     score,
            "threshold": threshold,
            "status":    status,
        })
    }
    return events
}

