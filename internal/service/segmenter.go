package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/redis/go-redis/v9"
)

// Segment represents the structure in our database
type Segment struct {
	Name      string          `json:"name"`
	RuleLogic json.RawMessage `json:"rule_logic"`
	Payload   json.RawMessage `json:"payload"`
}

// RunEvaluation pulls data from DB, runs rules, and updates Redis
func RunEvaluation(ctx context.Context, db *sql.DB, rdb *redis.Client) error {
	// 1. Get Segments
	segRows, _ := db.Query("SELECT name, rule_logic FROM segments")
	type segment struct {
		Name string
		Rule json.RawMessage
	}
	var segments []segment
	for segRows.Next() {
		var s segment
		segRows.Scan(&s.Name, &s.Rule)
		segments = append(segments, s)
	}

	// 2. Get Users
	userRows, _ := db.Query("SELECT user_id, orders_23d FROM user_metrics")
	for userRows.Next() {
		var uID string
		var oCount int
		userRows.Scan(&uID, &oCount)

		userData, _ := json.Marshal(map[string]int{"orders_23d": oCount})
		var matchedSegments []string

		for _, seg := range segments {
			var result bytes.Buffer
			jsonlogic.Apply(bytes.NewReader(seg.Rule), bytes.NewReader(userData), &result)
			if strings.TrimSpace(result.String()) == "true" {
				matchedSegments = append(matchedSegments, seg.Name)
			}
		}

		// 3. Update Redis
		redisKey := "user:segments:" + uID
		rdb.Del(ctx, redisKey)
		if len(matchedSegments) > 0 {
			rdb.SAdd(ctx, redisKey, matchedSegments)
		}
	}
	return nil
}

func EvaluateSpecificUser(ctx context.Context, db *sql.DB, rdb *redis.Client, uID string) error {
	// 1. Fetch current metrics and location for THIS user
	var oCount int
	var totalSpend float64
	var location string

	err := db.QueryRow(`
		SELECT orders_23d, total_spend, location_tag 
		FROM user_metrics 
		WHERE user_id = $1`, uID).Scan(&oCount, &totalSpend, &location)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User %s not found in metrics, skipping evaluation", uID)
			return nil
		}
		return err
	}

	// Prepare data for JsonLogic evaluation
	userData := map[string]interface{}{
		"orders_23d":   oCount,
		"total_spend":  totalSpend,
		"location_tag": location,
	}
	userDataBytes, _ := json.Marshal(userData)

	// 2. Fetch all defined segments
	rows, err := db.Query("SELECT name, rule_logic, payload FROM segments")
	if err != nil {
		return err
	}
	defer rows.Close()

	var matchedSegments []string
	var mergedPayloads = make(map[string]interface{})

	for rows.Next() {
		var s Segment
		if err := rows.Scan(&s.Name, &s.RuleLogic, &s.Payload); err != nil {
			continue
		}

		// Run JsonLogic evaluation
		var result bytes.Buffer
		err := jsonlogic.Apply(bytes.NewReader(s.RuleLogic), bytes.NewReader(userDataBytes), &result)
		if err != nil {
			log.Printf("Error evaluating rule for %s: %v", s.Name, err)
			continue
		}

		// If rule matches (returns "true")
		if strings.TrimSpace(result.String()) == "true" {
			matchedSegments = append(matchedSegments, s.Name)

			// Merge the segment's payload into the final experiment config
			var p map[string]interface{}
			if err := json.Unmarshal(s.Payload, &p); err == nil {
				for k, v := range p {
					mergedPayloads[k] = v
				}
			}
		}
	}

	// 3. Update Redis atomicly
	redisKeySegments := "user:segments:" + uID
	redisKeyPayload := "user:payload:" + uID

	pipe := rdb.Pipeline()

	// Clear old state
	pipe.Del(ctx, redisKeySegments, redisKeyPayload)

	if len(matchedSegments) > 0 {
		// Store segment names for quick lookup
		pipe.SAdd(ctx, redisKeySegments, matchedSegments)

		// Store the merged JSON payload (the Banners, Tiles, etc.)
		payloadBytes, _ := json.Marshal(mergedPayloads)
		pipe.Set(ctx, redisKeyPayload, payloadBytes, 0)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("Failed to update Redis for user %s: %v", uID, err)
		return err
	}

	log.Printf("✅ Re-evaluated %s: %d segments matched", uID, len(matchedSegments))
	return nil
}
