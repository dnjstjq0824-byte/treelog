package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (p *Pools) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := p.Primary.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
