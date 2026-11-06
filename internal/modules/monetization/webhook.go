package monetization

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// TossWebhook is the subset of Toss webhook events we reconcile.
type TossWebhook struct {
	EventType string `json:"eventType"`
	Data      struct {
		PaymentKey string `json:"paymentKey"`
		OrderID    string `json:"orderId"`
		Status     string `json:"status"`
	} `json:"data"`
}

// HandleWebhook reconciles a Toss webhook by re-reading the payment from
// Toss (webhooks carry no signature) and settling or canceling the order.
// Unknown orders are acknowledged without error so Toss stops retrying.
func (s *Service) HandleWebhook(ctx context.Context, raw json.RawMessage) error {
	if err := s.requirePool(); err != nil {
		return err
	}
	if err := s.requireToss(); err != nil {
		return err
	}
	var hook TossWebhook
	if err := json.Unmarshal(raw, &hook); err != nil {
		return &Fault{Status: http.StatusBadRequest, Code: "INVALID_WEBHOOK", Message: "Invalid webhook payload"}
	}
	if hook.Data.PaymentKey == "" || hook.Data.OrderID == "" {
		return &Fault{Status: http.StatusBadRequest, Code: "INVALID_WEBHOOK", Message: "Webhook is missing paymentKey/orderId"}
	}
	info, err := s.toss.GetPayment(ctx, hook.Data.PaymentKey)
	if err != nil {
		return &Fault{Status: http.StatusBadGateway, Code: "TOSS_LOOKUP_FAILED", Message: "Payment lookup failed: " + err.Error()}
	}
	var order Order
	err = s.pool.QueryRow(ctx, `
		SELECT id, buyer_user_id, order_type, recipient_blog_id, target_ref_id,
		       gross_amount_minor, platform_fee_minor, net_amount_minor, currency, status,
		       access_expires_at, created_at
		FROM orders WHERE id = $1`, hook.Data.OrderID).Scan(
		&order.ID, &order.BuyerUserID, &order.OrderType, &order.RecipientBlogID,
		&order.TargetRefID, &order.GrossAmountMinor, &order.PlatformFeeMinor,
		&order.NetAmountMinor, &order.Currency, &order.Status,
		&order.AccessExpiresAt, &order.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.OrderID != order.ID || info.AmountMinor != order.GrossAmountMinor {
		return &Fault{Status: http.StatusConflict, Code: "TOSS_ORDER_MISMATCH", Message: "Toss payment does not match the order"}
	}
	switch info.Status {
	case "DONE":
		if order.Status == "PAID" {
			return nil
		}
		if order.Status != "PENDING" {
			return nil
		}
		_, err = s.settleOrder(ctx, &order, info)
		return err
	case "CANCELED", "PARTIAL_CANCELED", "ABORTED", "EXPIRED":
		_, err = s.pool.Exec(ctx, `UPDATE orders SET status = 'CANCELED', updated_at = NOW()
			WHERE id = $1 AND status = 'PENDING'`, order.ID)
		return err
	default:
		// WAITING_FOR_DEPOSIT and other interim states: no local change.
		return nil
	}
}
