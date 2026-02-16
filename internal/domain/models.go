package domain

import (
	"encoding/json"
	"time"
)

// UserMetrics represents the aggregated state of a user.
type UserMetrics struct {
	UserID       string    `json:"user_id" db:"user_id"`
	OrderCount   int       `json:"order_count" db:"order_count_total"`
	Orders23d    int       `json:"orders_23d" db:"orders_23d"`
	LastOrderAt  time.Time `json:"last_order_at" db:"last_order_at"`
	LocationTag  string    `json:"location" db:"location_tag"`
	LTV          float64   `json:"ltv" db:"ltv"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Segment defines David's rules.
type Segment struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	RuleLogic json.RawMessage `json:"rule_logic"` // JSON-Logic format
}

// Experiment defines what the user gets.
type Experiment struct {
	SegmentID string          `json:"segment_id"`
	Key       string          `json:"key"` // e.g., "banners"
	Payload   json.RawMessage `json:"payload"`
}