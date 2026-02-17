package main

import (
	"context"
	"encoding/json"
	"log"

	"daffodil-experimentation-platform/internal/repository"
	"daffodil-experimentation-platform/internal/service" // Ensure this path is correct
	"daffodil-experimentation-platform/pkg/database"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// Updated to include the InstantSync flag and Location from our previous discussion
type OrderEvent struct {
	UserID      string  `json:"user_id"`
	Amount      float64 `json:"amount"`
	Location    string  `json:"location"`
	InstantSync bool    `json:"instant_sync"`
}

func main() {
	ctx := context.Background()

	// 1. Setup Postgres
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

	// 2. Setup Redis (NEW: Needed for the worker to update the cache)
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	// 3. Setup Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"127.0.0.1:9092"},
		Topic:    "order_events",
		GroupID:  "metrics-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Println("🚀 Worker started: Listening for order events...")

	metricsRepo := repository.NewPostgresMetricsRepository(db)

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var event OrderEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			continue
		}

		log.Printf("📥 Received Event: User=%s, InstantSync=%v", event.UserID, event.InstantSync)

		err = metricsRepo.UpsertOrder(ctx, event.UserID, event.Amount, event.Location)
		if err != nil {
			log.Printf("Repo Error: %v", err)
			continue
		}

		// // 4. Update Database (Matches your new schema)
		// query := `
		//     INSERT INTO user_metrics (user_id, orders_23d, total_spend, location_tag)
		//     VALUES ($1, 1, $2, $3)
		//     ON CONFLICT (user_id) DO UPDATE SET
		//         orders_23d = user_metrics.orders_23d + 1,
		//         total_spend = user_metrics.total_spend + EXCLUDED.total_spend,
		//         location_tag = EXCLUDED.location_tag,
		//         last_updated = NOW();
		// `
		// _, err = db.Exec(query, event.UserID, event.Amount, event.Location)

		// if err != nil {
		// 	log.Printf("Failed to update DB: %v", err)
		// 	continue
		// }
		log.Printf("Updated metrics for user: %s", event.UserID)

		// 5. TRIGGER EVALUATION (The Missing Link)
		// If the event is marked for InstantSync (Hot Path), we re-calculate segments immediately
		if event.InstantSync {
			log.Printf("⚡ [HOT PATH] Re-evaluating segments for: %s", event.UserID)
			err = service.EvaluateSpecificUser(ctx, db, rdb, event.UserID)
			if err != nil {
				log.Printf("Error in evaluation: %v", err)
			}
		}
	}
}
