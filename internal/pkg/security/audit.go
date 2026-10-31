package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditEntry struct {
	ActorUserID *string
	ActorRole   string
	ActionCode  string
	TargetType  string
	TargetID    string
	BeforeState json.RawMessage
	AfterState  json.RawMessage
	Reason      string
	IPAddress   string
	UserAgent   string
	RequestID   string
	CreatedAt   time.Time
}

// ComputeAuditHash calculates SHA-256 entry_hash chained with prev_hash for tamper evidence.
func ComputeAuditHash(prevHash string, entry AuditEntry) string {
	actor := ""
	if entry.ActorUserID != nil {
		actor = *entry.ActorUserID
	}
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		prevHash,
		actor,
		entry.ActorRole,
		entry.ActionCode,
		entry.TargetType,
		entry.TargetID,
		string(entry.BeforeState),
		string(entry.AfterState),
		entry.Reason,
		entry.IPAddress,
		entry.UserAgent,
		entry.RequestID,
		entry.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

type AuditLogger struct {
	pool *pgxpool.Pool
}

func NewAuditLogger(pool *pgxpool.Pool) *AuditLogger {
	return &AuditLogger{pool: pool}
}

// Log records a new tamper-evident chained audit log entry.
func (a *AuditLogger) Log(ctx context.Context, entry AuditEntry) error {
	if a.pool == nil {
		return errors.New("audit logger requires database pool")
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fetch previous hash within transaction (locking the latest record)
	var prevHash string
	err = tx.QueryRow(ctx, `SELECT entry_hash FROM audit_logs ORDER BY created_at DESC, id DESC LIMIT 1 FOR UPDATE`).Scan(&prevHash)
	if errors.Is(err, pgx.ErrNoRows) {
		prevHash = "0000000000000000000000000000000000000000000000000000000000000000"
	} else if err != nil {
		return err
	}

	entryHash := ComputeAuditHash(prevHash, entry)

	_, err = tx.Exec(ctx, `INSERT INTO audit_logs (
		actor_user_id, actor_role, action_code, target_type, target_id,
		before_state_json, after_state_json, reason, ip_address, user_agent,
		request_id, prev_hash, entry_hash, created_at
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, NULLIF($9, '')::inet, $10,
		$11, $12, $13, $14
	)`,
		entry.ActorUserID, entry.ActorRole, entry.ActionCode, entry.TargetType, entry.TargetID,
		entry.BeforeState, entry.AfterState, entry.Reason, entry.IPAddress, entry.UserAgent,
		entry.RequestID, prevHash, entryHash, entry.CreatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

