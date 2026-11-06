package notification

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Fault is the module error envelope.
type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string { return f.Message }

// Notification is one in-app notification row.
type Notification struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ActionURL string    `json:"action_url"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// Service reads and mutates in-app notifications.
type Service struct {
	pool *pgxpool.Pool
}

// NewService builds a notification service.
func NewService(pool *pgxpool.Pool) (*Service, error) {
	if pool == nil {
		return nil, errors.New("notification service requires a database pool")
	}
	return &Service{pool: pool}, nil
}

// List returns the newest-first notification page and the next cursor.
func (s *Service) List(ctx context.Context, userID string, limit int, cursor string, unreadOnly bool) ([]Notification, string, error) {
	if s.pool == nil {
		return nil, "", &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	query := `
		SELECT id, category, title, body, action_url, is_read, created_at
		FROM notifications
		WHERE recipient_user_id = $1
	`
	args := []any{userID}
	if unreadOnly {
		query += ` AND is_read = FALSE`
	}
	if cursor != "" {
		cursorTime, cursorID, err := decodeCursor(cursor)
		if err != nil {
			return nil, "", &Fault{Status: http.StatusBadRequest, Code: "INVALID_CURSOR", Message: "Invalid pagination cursor"}
		}
		query += fmt.Sprintf(` AND (created_at, id) < ($%d, $%d)`, len(args)+1, len(args)+2)
		args = append(args, cursorTime, cursorID)
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC, id DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit+1)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	items := make([]Notification, 0, limit)
	for rows.Next() {
		var item Notification
		if err := rows.Scan(&item.ID, &item.Category, &item.Title, &item.Body, &item.ActionURL, &item.IsRead, &item.CreatedAt); err != nil {
			return nil, "", err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	nextCursor := ""
	if len(items) > limit {
		last := items[limit-1]
		items = items[:limit]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
	}
	return items, nextCursor, nil
}

// UnreadCount returns the number of unread notifications.
func (s *Service) UnreadCount(ctx context.Context, userID string) (int64, error) {
	if s.pool == nil {
		return 0, &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	var count int64
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications
		WHERE recipient_user_id = $1 AND is_read = FALSE`, userID).Scan(&count)
	return count, err
}

// MarkRead marks one notification as read.
func (s *Service) MarkRead(ctx context.Context, userID, id string) error {
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	res, err := s.pool.Exec(ctx, `UPDATE notifications
		SET is_read = TRUE, read_at = NOW()
		WHERE id = $1 AND recipient_user_id = $2 AND is_read = FALSE`, id, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		var exists bool
		if err := s.pool.QueryRow(ctx, `SELECT TRUE FROM notifications
			WHERE id = $1 AND recipient_user_id = $2`, id, userID).Scan(&exists); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return &Fault{Status: http.StatusNotFound, Code: "NOTIFICATION_NOT_FOUND", Message: "Notification not found"}
			}
			return err
		}
	}
	return nil
}

// MarkAllRead marks every notification as read.
func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	_, err := s.pool.Exec(ctx, `UPDATE notifications
		SET is_read = TRUE, read_at = NOW()
		WHERE recipient_user_id = $1 AND is_read = FALSE`, userID)
	return err
}

func encodeCursor(createdAt time.Time, id string) string {
	raw := createdAt.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", err
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 || parts[1] == "" {
		return time.Time{}, "", errors.New("invalid cursor shape")
	}
	moment, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", err
	}
	return moment, parts[1], nil
}
