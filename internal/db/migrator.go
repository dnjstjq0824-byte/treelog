package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var migrationName = regexp.MustCompile(`^([0-9]{4})_[a-z0-9_]+\.up\.sql$`)

type Migrator struct{ Pool *pgxpool.Pool }

func migrationFiles(files fs.FS) ([]string, error) {
	names, err := fs.Glob(files, "*.up.sql")
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, errors.New("no migrations found")
	}
	for _, name := range names {
		if !migrationName.MatchString(filepath.Base(name)) {
			return nil, fmt.Errorf("invalid migration filename: %s", name)
		}
	}
	sort.Strings(names)
	return names, nil
}

func (m Migrator) withLock(ctx context.Context, fn func(*pgxpool.Conn) error) error {
	if m.Pool == nil {
		return errors.New("migration pool is required")
	}
	conn, err := m.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock(2017011101)"); err != nil {
		return err
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = conn.Exec(unlockCtx, "SELECT pg_advisory_unlock(2017011101)")
	}()
	_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(128) PRIMARY KEY,
		checksum CHAR(64) NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	if err != nil {
		return err
	}
	return fn(conn)
}

func (m Migrator) Up(ctx context.Context, files fs.FS) error {
	names, err := migrationFiles(files)
	if err != nil {
		return err
	}
	return m.withLock(ctx, func(conn *pgxpool.Conn) error {
		for _, name := range names {
			body, err := fs.ReadFile(files, name)
			if err != nil {
				return err
			}
			hash := sha256.Sum256(body)
			checksum := hex.EncodeToString(hash[:])
			var existing string
			err = conn.QueryRow(ctx, "SELECT checksum FROM schema_migrations WHERE version=$1", name).Scan(&existing)
			if err == nil {
				if existing != checksum {
					return fmt.Errorf("migration changed after application: %s", name)
				}
				continue
			}
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			tx, err := conn.Begin(ctx)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, string(body)); err != nil {
				_ = tx.Rollback(ctx)
				return fmt.Errorf("apply %s: %w", name, err)
			}
			if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations(version, checksum) VALUES($1,$2)", name, checksum); err != nil {
				_ = tx.Rollback(ctx)
				return err
			}
			if err := tx.Commit(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m Migrator) DownOne(ctx context.Context, files fs.FS) error {
	return m.withLock(ctx, func(conn *pgxpool.Conn) error {
		var version string
		err := conn.QueryRow(ctx, "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1").Scan(&version)
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("no applied migrations")
		}
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(version, ".up.sql") + ".down.sql"
		body, err := fs.ReadFile(files, name)
		if err != nil {
			return err
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("revert %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version=$1", version); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		return tx.Commit(ctx)
	})
}
