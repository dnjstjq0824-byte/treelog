package monetization

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"treeblog/internal/platform/tosspayments"
)

// PlatformFeePermille is the platform revenue share in permille (100 = 10%).
const PlatformFeePermille = 100

// WithholdingPermille is the creator payout withholding rate (33 = 3.3%).
const WithholdingPermille = 33

// Fault is the module error envelope.
type Fault struct {
	Status  int
	Code    string
	Message string
}

func (f *Fault) Error() string { return f.Message }

// Service implements orders, Toss confirm/webhook reconciliation,
// double-entry ledger posting, memberships, and payouts.
type Service struct {
	pool   *pgxpool.Pool
	toss   *tosspayments.Client
	pepper []byte
}

// NewService builds a monetization service. pepper must be 32+ bytes for
// bank-account encryption; toss may be disabled (no secret key) for
// order-only flows.
func NewService(pool *pgxpool.Pool, toss *tosspayments.Client, pepper []byte) (*Service, error) {
	if pool == nil {
		return nil, errors.New("monetization service requires a database pool")
	}
	if len(pepper) < 32 {
		return nil, errors.New("monetization service requires a 32-byte pepper")
	}
	return &Service{pool: pool, toss: toss, pepper: pepper}, nil
}

func (s *Service) requirePool() error {
	if s.pool == nil {
		return &Fault{Status: http.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "Database is unavailable"}
	}
	return nil
}

func (s *Service) requireToss() error {
	if s.toss == nil || !s.toss.Enabled() {
		return &Fault{Status: http.StatusNotImplemented, Code: "PAYMENTS_NOT_CONFIGURED", Message: "Payments are not configured"}
	}
	return nil
}

// platformFee splits gross into platform fee and creator net.
func platformFee(grossMinor int64) (fee, net int64) {
	fee = grossMinor * PlatformFeePermille / 1000
	return fee, grossMinor - fee
}

// Order is one order row.
type Order struct {
	ID              string     `json:"id"`
	BuyerUserID     string     `json:"buyer_user_id"`
	OrderType       string     `json:"order_type"`
	RecipientBlogID *string    `json:"recipient_blog_id,omitempty"`
	TargetRefID     *string    `json:"target_ref_id,omitempty"`
	GrossAmountMinor int64     `json:"gross_amount_minor"`
	PlatformFeeMinor int64     `json:"platform_fee_minor"`
	NetAmountMinor  int64      `json:"net_amount_minor"`
	Currency        string     `json:"currency"`
	Status          string     `json:"status"`
	AccessExpiresAt *time.Time `json:"access_expires_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateOrderInput carries the order payload.
type CreateOrderInput struct {
	OrderType       string  `json:"order_type"`
	RecipientBlogID *string `json:"recipient_blog_id,omitempty"`
	TargetRefID     *string `json:"target_ref_id,omitempty"`
	GrossAmountMinor int64  `json:"gross_amount_minor"`
	Currency        string  `json:"currency"`
	IdempotencyKey  string  `json:"idempotency_key"`
	DonationMessage string  `json:"donation_message"`
}

// CreateOrder validates pricing against catalog/target rows and inserts a
// PENDING order. Retries with the same idempotency key return the original.
func (s *Service) CreateOrder(ctx context.Context, buyerID string, input CreateOrderInput) (*Order, error) {
	orderType := strings.ToUpper(strings.TrimSpace(input.OrderType))
	switch orderType {
	case "MEMBERSHIP_SUB", "SINGLE_POST_UNLOCK", "READER_DONATION", "AD_PREPAID_TOPUP":
	default:
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_ORDER_TYPE", Message: "Unknown order type"}
	}
	if strings.ToUpper(strings.TrimSpace(input.Currency)) != "KRW" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "UNSUPPORTED_CURRENCY", Message: "Only KRW is supported"}
	}
	if strings.TrimSpace(input.IdempotencyKey) == "" || len(input.IdempotencyKey) > 128 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_IDEMPOTENCY_KEY", Message: "Idempotency key is required (max 128 chars)"}
	}
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var existing Order
	err = tx.QueryRow(ctx, `
		SELECT id, buyer_user_id, order_type, recipient_blog_id, target_ref_id,
		       gross_amount_minor, platform_fee_minor, net_amount_minor, currency, status,
		       access_expires_at, created_at
		FROM orders WHERE idempotency_key = $1 AND buyer_user_id = $2`,
		input.IdempotencyKey, buyerID).Scan(
		&existing.ID, &existing.BuyerUserID, &existing.OrderType, &existing.RecipientBlogID,
		&existing.TargetRefID, &existing.GrossAmountMinor, &existing.PlatformFeeMinor,
		&existing.NetAmountMinor, &existing.Currency, &existing.Status,
		&existing.AccessExpiresAt, &existing.CreatedAt,
	)
	if err == nil {
		_ = tx.Commit(ctx)
		return &existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	accessExpires, err := s.validatePricing(ctx, tx, orderType, input)
	if err != nil {
		return nil, err
	}
	fee, net := platformFee(input.GrossAmountMinor)
	if orderType == "AD_PREPAID_TOPUP" {
		fee, net = 0, input.GrossAmountMinor
	}
	orderID := "ord-" + uuid.NewString()
	var order Order
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (id, buyer_user_id, order_type, recipient_blog_id, target_ref_id,
		                    gross_amount_minor, platform_fee_minor, net_amount_minor, currency,
		                    status, idempotency_key, access_expires_at, donation_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'KRW', 'PENDING', $9, $10, NULLIF($11, ''))
		RETURNING id, buyer_user_id, order_type, recipient_blog_id, target_ref_id,
		          gross_amount_minor, platform_fee_minor, net_amount_minor, currency, status,
		          access_expires_at, created_at
	`, orderID, buyerID, orderType, input.RecipientBlogID, input.TargetRefID,
		input.GrossAmountMinor, fee, net, input.IdempotencyKey, accessExpires, input.DonationMessage).Scan(
		&order.ID, &order.BuyerUserID, &order.OrderType, &order.RecipientBlogID,
		&order.TargetRefID, &order.GrossAmountMinor, &order.PlatformFeeMinor,
		&order.NetAmountMinor, &order.Currency, &order.Status,
		&order.AccessExpiresAt, &order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *Service) validatePricing(ctx context.Context, tx pgx.Tx, orderType string, input CreateOrderInput) (*time.Time, error) {
	switch orderType {
	case "MEMBERSHIP_SUB":
		if input.RecipientBlogID == nil || input.TargetRefID == nil {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_ORDER_TARGET", Message: "Membership orders require recipient_blog_id and target_ref_id (membership id)"}
		}
		var blogID string
		var monthly, annual int64
		err := tx.QueryRow(ctx, `SELECT blog_id, monthly_price_minor, COALESCE(annual_price_minor, 0)
			FROM memberships WHERE id = $1 AND is_active = TRUE`, *input.TargetRefID).
			Scan(&blogID, &monthly, &annual)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "MEMBERSHIP_NOT_FOUND", Message: "Membership not found or inactive"}
		}
		if err != nil {
			return nil, err
		}
		if blogID != *input.RecipientBlogID {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "MEMBERSHIP_BLOG_MISMATCH", Message: "Membership does not belong to the recipient blog"}
		}
		now := time.Now().UTC()
		if input.GrossAmountMinor == monthly {
			expires := now.AddDate(0, 1, 0)
			return &expires, nil
		}
		if annual > 0 && input.GrossAmountMinor == annual {
			expires := now.AddDate(1, 0, 0)
			return &expires, nil
		}
		return nil, &Fault{Status: http.StatusBadRequest, Code: "PRICE_MISMATCH", Message: "Amount does not match the membership price"}
	case "SINGLE_POST_UNLOCK":
		if input.TargetRefID == nil {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_ORDER_TARGET", Message: "Post unlock orders require target_ref_id (post id)"}
		}
		var price *int64
		var blogID string
		err := tx.QueryRow(ctx, `SELECT paywall_price_minor, blog_id FROM posts
			WHERE id = $1 AND status = 'PUBLISHED'`, *input.TargetRefID).Scan(&price, &blogID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "POST_NOT_FOUND", Message: "Post not found"}
		}
		if err != nil {
			return nil, err
		}
		if price == nil || *price <= 0 {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "POST_NOT_FOR_SALE", Message: "Post is not for sale"}
		}
		if input.GrossAmountMinor != *price {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "PRICE_MISMATCH", Message: "Amount does not match the post price"}
		}
		return nil, nil
	case "READER_DONATION":
		if input.RecipientBlogID == nil {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_ORDER_TARGET", Message: "Donations require recipient_blog_id"}
		}
		if input.GrossAmountMinor < 100 {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "AMOUNT_TOO_SMALL", Message: "Minimum donation is 100 KRW"}
		}
		return nil, nil
	case "AD_PREPAID_TOPUP":
		if input.GrossAmountMinor < 10000 {
			return nil, &Fault{Status: http.StatusBadRequest, Code: "AMOUNT_TOO_SMALL", Message: "Minimum top-up is 10000 KRW"}
		}
		return nil, nil
	default:
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_ORDER_TYPE", Message: "Unknown order type"}
	}
}

// GetOrder returns one order owned by the buyer.
func (s *Service) GetOrder(ctx context.Context, buyerID, orderID string) (*Order, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	var order Order
	err := s.pool.QueryRow(ctx, `
		SELECT id, buyer_user_id, order_type, recipient_blog_id, target_ref_id,
		       gross_amount_minor, platform_fee_minor, net_amount_minor, currency, status,
		       access_expires_at, created_at
		FROM orders WHERE id = $1 AND buyer_user_id = $2`, orderID, buyerID).Scan(
		&order.ID, &order.BuyerUserID, &order.OrderType, &order.RecipientBlogID,
		&order.TargetRefID, &order.GrossAmountMinor, &order.PlatformFeeMinor,
		&order.NetAmountMinor, &order.Currency, &order.Status,
		&order.AccessExpiresAt, &order.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: http.StatusNotFound, Code: "ORDER_NOT_FOUND", Message: "Order not found"}
	}
	return &order, err
}

// ListOrders returns the buyer's orders newest-first.
func (s *Service) ListOrders(ctx context.Context, buyerID string, limit, offset int) ([]Order, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, buyer_user_id, order_type, recipient_blog_id, target_ref_id,
		       gross_amount_minor, platform_fee_minor, net_amount_minor, currency, status,
		       access_expires_at, created_at
		FROM orders WHERE buyer_user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, buyerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]Order, 0)
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.BuyerUserID, &order.OrderType, &order.RecipientBlogID,
			&order.TargetRefID, &order.GrossAmountMinor, &order.PlatformFeeMinor,
			&order.NetAmountMinor, &order.Currency, &order.Status,
			&order.AccessExpiresAt, &order.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

// ConfirmPaymentInput carries the Toss confirm payload from the client.
type ConfirmPaymentInput struct {
	PaymentKey string `json:"paymentKey"`
	OrderID    string `json:"orderId"`
	Amount     int64  `json:"amount"`
}

// ConfirmPayment verifies the local order, confirms with Toss, and posts
// payment + ledger rows atomically. Safe to retry: paid orders return the
// existing state.
func (s *Service) ConfirmPayment(ctx context.Context, buyerID string, input ConfirmPaymentInput) (*Order, error) {
	if strings.TrimSpace(input.PaymentKey) == "" || strings.TrimSpace(input.OrderID) == "" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_CONFIRM_PAYLOAD", Message: "paymentKey and orderId are required"}
	}
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if err := s.requireToss(); err != nil {
		return nil, err
	}
	order, err := s.GetOrder(ctx, buyerID, input.OrderID)
	if err != nil {
		return nil, err
	}
	if order.Status == "PAID" {
		return order, nil
	}
	if order.Status != "PENDING" {
		return nil, &Fault{Status: http.StatusConflict, Code: "ORDER_NOT_PAYABLE", Message: "Order is not payable"}
	}
	if input.Amount != order.GrossAmountMinor {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "AMOUNT_MISMATCH", Message: "Amount does not match the order"}
	}
	info, err := s.toss.ConfirmPayment(ctx, input.PaymentKey, input.OrderID, input.Amount)
	if err != nil {
		return nil, &Fault{Status: http.StatusBadGateway, Code: "PAYMENT_CONFIRM_FAILED", Message: "Payment confirmation failed: " + err.Error()}
	}
	if info.OrderID != order.ID || info.AmountMinor != order.GrossAmountMinor {
		return nil, &Fault{Status: http.StatusConflict, Code: "TOSS_ORDER_MISMATCH", Message: "Toss payment does not match the order"}
	}
	return s.settleOrder(ctx, order, info)
}

// settleOrder marks an order PAID and posts payment + ledger rows in one
// transaction. It is idempotent per order.
func (s *Service) settleOrder(ctx context.Context, order *Order, info *tosspayments.PaymentInfo) (*Order, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1 FOR UPDATE`, order.ID).Scan(&status)
	if err != nil {
		return nil, err
	}
	if status == "PAID" {
		_ = tx.Commit(ctx)
		order.Status = "PAID"
		return order, nil
	}
	if status != "PENDING" {
		return nil, &Fault{Status: http.StatusConflict, Code: "ORDER_NOT_PAYABLE", Message: "Order is not payable"}
	}
	paidAt := time.Now().UTC()
	if info.ApprovedAt != nil {
		paidAt = *info.ApprovedAt
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO payments (order_id, buyer_user_id, pg_provider, pg_transaction_id, method, amount_minor, currency, status, paid_at, pg_payload_json)
		VALUES ($1, $2, 'TOSS_PAYMENTS', $3, $4, $5, $6, 'CAPTURED', $7, $8)
		ON CONFLICT (pg_transaction_id) DO NOTHING
	`, order.ID, order.BuyerUserID, info.PaymentKey, methodCode(info.Method), info.AmountMinor, order.Currency, paidAt, string(info.Raw)); err != nil {
		return nil, err
	}
	if order.OrderType == "AD_PREPAID_TOPUP" {
		if err := postLedger(tx, ctx, order, order.BuyerUserID, nil, "PLATFORM_CLEARING", "DEBIT", order.GrossAmountMinor, "ORDER_PAYMENT"); err != nil {
			return nil, err
		}
		if err := postLedger(tx, ctx, order, order.BuyerUserID, nil, "ADVERTISER_PREPAID", "CREDIT", order.GrossAmountMinor, "ORDER_PAYMENT"); err != nil {
			return nil, err
		}
	} else {
		var ownerID string
		blogID := ""
		if order.RecipientBlogID != nil {
			blogID = *order.RecipientBlogID
			if err := tx.QueryRow(ctx, `SELECT owner_user_id FROM blogs WHERE id = $1`, blogID).Scan(&ownerID); err != nil {
				return nil, err
			}
		}
		var blogParam *string
		if blogID != "" {
			blogParam = &blogID
		}
		if err := postLedger(tx, ctx, order, order.BuyerUserID, nil, "PLATFORM_CLEARING", "DEBIT", order.GrossAmountMinor, "ORDER_PAYMENT"); err != nil {
			return nil, err
		}
		if err := postLedger(tx, ctx, order, ownerID, blogParam, "CREATOR_PENDING_ESCROW", "CREDIT", order.NetAmountMinor, "ORDER_PAYMENT"); err != nil {
			return nil, err
		}
		if order.PlatformFeeMinor > 0 {
			if err := postLedger(tx, ctx, order, "", nil, "PLATFORM_FEE_REVENUE", "CREDIT", order.PlatformFeeMinor, "ORDER_PAYMENT"); err != nil {
				return nil, err
			}
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE orders SET status = 'PAID', updated_at = NOW() WHERE id = $1`, order.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	order.Status = "PAID"
	return order, nil
}

func postLedger(tx pgx.Tx, ctx context.Context, order *Order, ownerUserID string, blogID *string, account, direction string, amount int64, entryType string) error {
	var ownerParam *string
	if ownerUserID != "" {
		ownerParam = &ownerUserID
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO ledger_entries (owner_user_id, blog_id, account_code, direction, amount_minor, currency, reference_type, reference_id, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (idempotency_key) DO NOTHING
	`, ownerParam, blogID, account, direction, amount, order.Currency, entryType, order.ID,
		"ledger-"+order.ID+"-"+account+"-"+direction)
	return err
}

func methodCode(tossMethod string) string {
	switch tossMethod {
	case "카드":
		return "CARD"
	case "가상계좌":
		return "VIRTUAL_ACCOUNT"
	case "간편결제":
		return "EASY_PAY"
	case "휴대폰":
		return "MOBILE"
	case "계좌이체":
		return "BANK_TRANSFER"
	case "문화상품권", "도서문화상품권", "게임문화상품권":
		return "GIFT_CERT"
	default:
		if tossMethod == "" {
			return "CARD"
		}
		return "EASY_PAY"
	}
}
