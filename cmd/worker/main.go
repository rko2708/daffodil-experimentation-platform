package main

import (
	"context"
	"encoding/json"
	"log"

	"daffodil-experimentation-platform/pkg/database"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

type OrderEvent struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

func main() {
	// 1. Setup Postgres Connection
	cfg := database.DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "password",
		DBName:   "daffodil",
	}

	db, err := database.NewPostgresConn(cfg)
	if err != nil {
		log.Fatal("Could not connect to DB:", err)
	}
	defer db.Close()

	// 2. Setup Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"127.0.0.1:9092"},
		Topic:    "order_events",
		GroupID:  "metrics-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	log.Println("Worker started: Listening for order events...")

	for {
		log.Println("inside...")
		// Read Message
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var event OrderEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			continue
		}

		// 3. Upsert User Metrics
		// We increment order_count_total and orders_23d, and update LTV
		query := `
			INSERT INTO user_metrics (user_id, order_count_total, orders_23d, ltv, last_order_at, updated_at)
			VALUES ($1, 1, 1, $2, NOW(), NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				order_count_total = user_metrics.order_count_total + 1,
				orders_23d = user_metrics.orders_23d + 1,
				ltv = user_metrics.ltv + EXCLUDED.ltv,
				last_order_at = NOW(),
				updated_at = NOW();
		`
		_, err = db.Exec(query, event.UserID, event.Amount)
		if err != nil {
			log.Printf("Failed to update DB: %v", err)
		} else {
			log.Printf("Updated metrics for user: %s", event.UserID)
		}
	}
}
