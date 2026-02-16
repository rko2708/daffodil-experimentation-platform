package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"daffodil-experimentation-platform/internal/domain"
	// "daffodil-experimentation-platform/internal/ruleengine"

	"github.com/diegoholiveira/jsonlogic/v3"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Connections
	db, _ := sql.Open("postgres", "postgres://user:password@localhost:5432/daffodil?sslmode=disable")
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	log.Println("Cron Job Started: Evaluating segments...")

	// 2. Fetch all active segments/rules
	rows, _ := db.Query("SELECT id, name, rule_logic FROM segments WHERE is_active = true")
	var segments []domain.Segment
	for rows.Next() {
		var s domain.Segment
		rows.Scan(&s.ID, &s.Name, &s.RuleLogic)
		segments = append(segments, s)
	}

	// 3. Fetch users who have placed orders recently (Demo: fetch all)
	userRows, _ := db.Query("SELECT user_id, orders_23d, location_tag FROM user_metrics")

	// for userRows.Next() {
	// 	var u domain.UserMetrics
	// 	userRows.Scan(&u.UserID, &u.Orders23d, &u.LocationTag)

	// 	// Map user data for the Rule Engine
	// 	userData := map[string]interface{}{
	// 		"orders_23d": u.Orders23d,
	// 		"location":   u.LocationTag,
	// 	}

	// 	for _, segment := range segments {
	// 		// 4. Run the JSON-Logic check
	// 		isMember, _ := ruleengine.Evaluate(segment.RuleLogic, userData)

	// 		if isMember {
	// 			log.Printf("User %s added to segment: %s", u.UserID, segment.Name)
	// 			// 5. Store in Redis (Key: user:segments:{id}, Value: segment_id)
	// 			rdb.SAdd(ctx, "user:segments:"+u.UserID, segment.ID)
	// 		} else {
	// 			// Remove if they no longer qualify
	// 			rdb.SRem(ctx, "user:segments:"+u.UserID, segment.ID)
	// 		}
	// 	}
	// }

	// cmd/cron/main.go inside the userRows.Next() loop

	for userRows.Next() {
		var uID string
		var oCount int
		var location sql.NullString
		userRows.Scan(&uID, &oCount)

		log.Printf("🔍 Checking User %s: %d orders", uID, oCount)
		if err := userRows.Scan(&uID, &oCount, &location); err != nil {
			log.Printf("❌ Scan Error: %v", err)
			continue
		}
		// userRows, err := db.Query("SELECT user_id, orders_23d FROM user_metrics")
		// if err != nil {
		// 	log.Fatalf("Query failed: %v", err)
		// }
		// defer userRows.Close()

		// Now uID should not be empty!
		log.Printf("🔍 Processing User: [%s] | Orders: %d", uID, oCount)

		// 1. Create the data context
		data := map[string]interface{}{
			"orders_23d": oCount,
		}

		// 2. Convert map to JSON bytes (the safest way for the engine to read it)
		jsonData, _ := json.Marshal(data)

		for _, segment := range segments {
			// We use bytes.Buffer because the library Apply method takes an io.Writer
			var result bytes.Buffer
			ruleReader := strings.NewReader(string(segment.RuleLogic)) // segment.RuleLogic is the JSON string from DB
			dataReader := bytes.NewReader(jsonData)

			err := jsonlogic.Apply(ruleReader, dataReader, &result)
			if err != nil {
				log.Printf("❌ Engine Error: %v", err)
				continue
			}

			log.Println("result---", result.String())
			// 3. The library writes "true" or "false" to the buffer
			if strings.TrimSpace(result.String()) == "true" {
				log.Printf("✅ MATCH: User %s is a %s", uID, segment.Name)
				rdb.SAdd(ctx, "user:segments:"+uID, segment.Name)
			} else {
				log.Printf("ℹ️ SKIP: User %s (Result was: %s)", uID, result.String())
			}
		}
	}
	log.Println("Cron Job Finished.")
}
