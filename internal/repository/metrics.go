package repository

import (
	"context"
	"database/sql"
)

// UserMetrics represents the database structure
type UserMetrics struct {
	UserID      string
	Orders23d   int
	TotalSpend  float64
	LocationTag string
}

// MetricsRepository defines the operations for user data
type MetricsRepository interface {
	UpsertOrder(ctx context.Context, userID string, amount float64, location string) error
	GetMetrics(ctx context.Context, userID string) (*UserMetrics, error)
	EnsureUser(ctx context.Context, userID string) error
}

type postgresMetricsRepo struct {
	db *sql.DB
}

func NewPostgresMetricsRepository(db *sql.DB) MetricsRepository {
	return &postgresMetricsRepo{db: db}
}

func (r *postgresMetricsRepo) UpsertOrder(ctx context.Context, userID string, amount float64, location string) error {
	query := `
        INSERT INTO user_metrics (user_id, orders_23d, total_spend, location_tag, last_updated)
        VALUES ($1, 1, $2, $3, NOW())
        ON CONFLICT (user_id) DO UPDATE SET
            orders_23d = user_metrics.orders_23d + 1,
            total_spend = user_metrics.total_spend + EXCLUDED.total_spend,
            location_tag = EXCLUDED.location_tag,
            last_updated = NOW();`

	_, err := r.db.ExecContext(ctx, query, userID, amount, location)
	return err
}

func (r *postgresMetricsRepo) GetMetrics(ctx context.Context, userID string) (*UserMetrics, error) {
	m := &UserMetrics{}
	query := `SELECT user_id, orders_23d, total_spend, location_tag FROM user_metrics WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&m.UserID, &m.Orders23d, &m.TotalSpend, &m.LocationTag)
	return m, err
}

func (r *postgresMetricsRepo) EnsureUser(ctx context.Context, userID string) error {
	// We initialize with 0 orders and 0 spend
	query := `
        INSERT INTO user_metrics (user_id, orders_23d, total_spend, location_tag) 
        VALUES ($1, 0, 0.0, 'unknown') 
        ON CONFLICT (user_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
