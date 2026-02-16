package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"daffodil-experimentation-platform/internal/service"
	"daffodil-experimentation-platform/pkg/config"
	"daffodil-experimentation-platform/pkg/database"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

var (
	db          *sql.DB
	rdb         *redis.Client
	kafkaWriter *kafka.Writer
	ctx         = context.Background()
)

func main() {
	// 1. Setup DB
	cfg := config.LoadConfig()

	var err error
	db, err = database.NewPostgresConn(database.DBConfig{
		Host:     cfg.DBHost,
		Port:     5432,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 1. Connect to Redis
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	// 3. Setup Kafka Writer
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "order_events",
		Balancer: &kafka.LeastBytes{},
	}

	// 2. Define the endpoint
	http.HandleFunc("/experiments", getExperiments)
	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/place-order", handlePlaceOrder)
	http.HandleFunc("/evaluate", runEvaluation)

	log.Println("🚀 Experiment API started on :8080")
	log.Fatal(http.ListenAndServe(":"+cfg.APIPort, enableCORS(http.DefaultServeMux)))

}

func getExperiments(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	userID := r.URL.Query().Get("userId")

	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	// 3. Fetch segment names from Redis Set
	segments, err := rdb.SMembers(ctx, "user:segments:"+userID).Result()
	if err != nil {
		log.Printf("Redis Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. Response Structure
	response := map[string]interface{}{
		"user_id":  userID,
		"segments": segments,
		"features": make(map[string]interface{}),
	}

	// Apply logic based on segments found in Redis
	features := response["features"].(map[string]interface{})
	for _, s := range segments {
		// Matching the name we seeded in init.sql
		if s == "Power User" {
			features["show_pizza_tile"] = true
			features["home_banner"] = "Premium_Banner_V1"
			features["discount_pct"] = 15
		}
	}

	// 5. Log request for the demo
	log.Printf("GET /experiments?userId=%s - Found %d segments - Latency: %v", userID, len(segments), time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT user_id, orders_23d FROM user_metrics")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		users := []map[string]interface{}{}
		for rows.Next() {
			var id string
			var count int
			rows.Scan(&id, &count)
			users = append(users, map[string]interface{}{"user_id": id, "orders": count})
		}
		json.NewEncoder(w).Encode(users)

	case http.MethodPost:
		var req struct {
			UserID string `json:"user_id"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		_, err := db.Exec("INSERT INTO user_metrics (user_id, orders_23d) VALUES ($1, 0) ON CONFLICT (user_id) DO NOTHING", req.UserID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func handlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Count  int    `json:"count"` // Allow sending multiple orders at once
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Count <= 0 {
		req.Count = 1
	}

	for i := 0; i < req.Count; i++ {
		msg := map[string]interface{}{
			"user_id": req.UserID,
			"amount":  100.0,
		}
		msgBytes, _ := json.Marshal(msg)

		err := kafkaWriter.WriteMessages(ctx, kafka.Message{
			Key:   []byte(req.UserID),
			Value: msgBytes,
		})
		if err != nil {
			log.Printf("Kafka Write Error: %v", err)
			http.Error(w, "Failed to send to Kafka", 500)
			return
		}
	}

	log.Printf("✅ Produced %d orders for %s to Kafka", req.Count, req.UserID)
	w.WriteHeader(http.StatusAccepted)
}

func runEvaluation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	// Use the shared service
	err := service.RunEvaluation(ctx, db, rdb)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3001")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
