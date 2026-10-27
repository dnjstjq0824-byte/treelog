package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Event struct {
	ID             string          `json:"event_id"`
	AggregateType  string          `json:"aggregate_type"`
	AggregateID    string          `json:"aggregate_id"`
	Type           string          `json:"event_type"`
	Payload        json.RawMessage `json:"payload"`
	IdempotencyKey string          `json:"-"`
}

// PublishTx records an event in the caller's transaction. A rolled-back domain
// write cannot leave a queued event behind.
func PublishTx(ctx context.Context, tx pgx.Tx, event Event) (string, error) {
	if tx == nil {
		return "", errors.New("transaction is required")
	}
	if _, err := uuid.Parse(event.AggregateID); err != nil {
		return "", fmt.Errorf("invalid aggregate id: %w", err)
	}
	if strings.TrimSpace(event.AggregateType) == "" || len(event.AggregateType) > 40 ||
		strings.TrimSpace(event.Type) == "" || len(event.Type) > 64 ||
		strings.TrimSpace(event.IdempotencyKey) == "" || len(event.IdempotencyKey) > 160 ||
		!json.Valid(event.Payload) {
		return "", errors.New("invalid outbox event")
	}
	var id string
	err := tx.QueryRow(ctx, `INSERT INTO outbox_events
		(aggregate_type, aggregate_id, event_type, payload_json, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (idempotency_key) DO NOTHING RETURNING id`,
		event.AggregateType, event.AggregateID, event.Type, string(event.Payload), event.IdempotencyKey).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
	var existing Event
	var samePayload bool
	err = tx.QueryRow(ctx, `SELECT id, aggregate_type, aggregate_id, event_type,
		payload_json=$2::jsonb FROM outbox_events WHERE idempotency_key=$1`,
		event.IdempotencyKey, string(event.Payload)).Scan(
		&existing.ID, &existing.AggregateType, &existing.AggregateID, &existing.Type, &samePayload)
	if err != nil {
		return "", err
	}
	if existing.AggregateType != event.AggregateType || existing.AggregateID != event.AggregateID ||
		existing.Type != event.Type || !samePayload {
		return "", errors.New("idempotency key reused with different event")
	}
	return existing.ID, nil
}
