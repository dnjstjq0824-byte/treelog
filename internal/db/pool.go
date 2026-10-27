package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pools struct {
	Primary *pgxpool.Pool
	Replica *pgxpool.Pool
}

func Open(ctx context.Context, primaryURL, replicaURL string) (*Pools, error) {
	if primaryURL == "" {
		return nil, errors.New("TREEBLOG_DATABASE_URL is required")
	}
	primary, err := openPool(ctx, primaryURL)
	if err != nil {
		return nil, fmt.Errorf("primary database: %w", err)
	}
	pools := &Pools{Primary: primary}
	if replicaURL != "" {
		replica, err := openPool(ctx, replicaURL)
		if err != nil {
			primary.Close()
			return nil, fmt.Errorf("replica database: %w", err)
		}
		pools.Replica = replica
	}
	return pools, nil
}

func openPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	// Blueprint SQL migrations contain multiple statements per file.
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func (p *Pools) Close() {
	if p.Replica != nil {
		p.Replica.Close()
	}
	if p.Primary != nil {
		p.Primary.Close()
	}
}

// ReadPool pins reads to the primary after a write to ensure read-your-writes.
func (p *Pools) ReadPool(afterWrite bool) *pgxpool.Pool {
	if afterWrite || p.Replica == nil {
		return p.Primary
	}
	return p.Replica
}
