package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/redis/go-redis/v9"
)

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
