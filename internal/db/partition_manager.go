package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var PartitionedTables = []string{
	"notifications",
	"search_click_events",
	"ad_events",
	"audit_logs",
}

// EnsureMonthlyPartitions ensures that partitions exist for the given year and month.
func EnsureMonthlyPartitions(ctx context.Context, pool *pgxpool.Pool, t time.Time) error {
	year, month, _ := t.Date()
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	partitionSuffix := fmt.Sprintf("%04d_%02d", year, int(month))

	for _, table := range PartitionedTables {
		partitionName := fmt.Sprintf("%s_%s", table, partitionSuffix)
		query := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s PARTITION OF %s
			FOR VALUES FROM ('%s') TO ('%s')
		`, partitionName, table, startOfMonth.Format("2006-01-02 15:04:05+00"), endOfMonth.Format("2006-01-02 15:04:05+00"))

		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed creating partition %s for table %s: %w", partitionName, table, err)
		}
	}
	return nil
}

// RunPartitionMaintenance creates partitions for the current month and the following month.
func RunPartitionMaintenance(ctx context.Context, pool *pgxpool.Pool) error {
	now := time.Now().UTC()
	// Current month
	if err := EnsureMonthlyPartitions(ctx, pool, now); err != nil {
		return err
	}
	// Next month
	nextMonth := now.AddDate(0, 1, 0)
	return EnsureMonthlyPartitions(ctx, pool, nextMonth)
}
