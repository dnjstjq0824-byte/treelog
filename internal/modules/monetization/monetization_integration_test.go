package monetization

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"treeblog/internal/db"
	"treeblog/internal/platform/tosspayments"
)

func setupIntegration(t *testing.T) (*Service, *tosspayments.Client, func()) {
	t.Helper()
	url := os.Getenv("TREEBLOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TREEBLOG_TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pools, err := db.Open(ctx, url, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := (db.Migrator{Pool: pools.Primary}).Up(ctx, os.DirFS("../../../migrations")); err != nil {
		pools.Close()
		t.Fatal(err)
	}
	pepper := make([]byte, 32)
	_, _ = rand.Read(pepper)
	toss := tosspayments.New(tosspayments.Config{SecretKey: "test-secret"})
	svc, err := NewService(pools.Primary, toss, pepper)
	if err != nil {
		pools.Close()
		t.Fatal(err)
	}
	return svc, toss, func() { pools.Close() }
}

// fakeToss emulates confirm + lookup for one expected order.
func fakeToss(t *testing.T, orderID string, amount int64) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost && r.URL.Path == "/v1/payments/confirm" {
			raw, _ := io.ReadAll(r.Body)
			var req struct {
				PaymentKey string `json:"paymentKey"`
				OrderID    string `json:"orderId"`
				Amount     int64  `json:"amount"`
			}
			_ = json.Unmarshal(raw, &req)
			if req.OrderID != orderID || req.Amount != amount {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"code":"INVALID_ORDER","message":"order mismatch"}`))
				return
			}
			_, _ = fmt.Fprintf(w, `{"paymentKey":%q,"orderId":%q,"status":"DONE","method":"카드","currency":"KRW","totalAmount":%d,"approvedAt":"2026-05-01T12:00:00+09:00"}`, req.PaymentKey, orderID, amount)
			return
		}
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/payments/") {
			key := strings.TrimPrefix(r.URL.Path, "/v1/payments/")
			_, _ = fmt.Fprintf(w, `{"paymentKey":%q,"orderId":%q,"status":"DONE","method":"카드","currency":"KRW","totalAmount":%d}`, key, orderID, amount)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func seedUser(t *testing.T, svc *Service, ctx context.Context, tag string) string {
	t.Helper()
	userID := uuid.NewString()
	handle := "mo-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	if _, err := svc.pool.Exec(ctx, `INSERT INTO users (id, email, email_verified, status)
		VALUES ($1, $2, TRUE, 'ACTIVE')`, userID, tag+"-"+userID[:8]+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.pool.Exec(ctx, `INSERT INTO user_profiles (user_id, username, display_name)
		VALUES ($1, $2, 'Money User')`, userID, handle); err != nil {
		t.Fatal(err)
	}
	return userID
}

func TestMonetization_Flow(t *testing.T) {
	svc, toss, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()
	buyerID := seedUser(t, svc, ctx, "mobuyer")
	ownerID := seedUser(t, svc, ctx, "moowner")

	blogID := uuid.NewString()
	blogHandle := "mob-" + strings.ReplaceAll(uuid.NewString()[:8], "-", "x")
	if _, err := svc.pool.Exec(ctx, `INSERT INTO blogs (id, owner_user_id, handle, title)
		VALUES ($1, $2, $3, 'Money Blog')`, blogID, ownerID, blogHandle); err != nil {
		t.Fatal(err)
	}

	order, err := svc.CreateOrder(ctx, buyerID, CreateOrderInput{
		OrderType:        "READER_DONATION",
		RecipientBlogID:  &blogID,
		GrossAmountMinor: 15000,
		Currency:         "KRW",
		IdempotencyKey:   "idem-" + uuid.NewString(),
		DonationMessage:  "응원합니다",
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if order.PlatformFeeMinor != 1500 || order.NetAmountMinor != 13500 || order.Status != "PENDING" {
		t.Fatalf("order: %+v", order)
	}

	server := fakeToss(t, order.ID, 15000)
	defer server.Close()
	toss.HTTPClient = server.Client()

	paid, err := svc.ConfirmPayment(ctx, buyerID, ConfirmPaymentInput{
		PaymentKey: "pay-" + uuid.NewString()[:8], OrderID: order.ID, Amount: 15000,
	})
	if err != nil {
		t.Fatalf("ConfirmPayment: %v", err)
	}
	if paid.Status != "PAID" {
		t.Fatalf("status = %s", paid.Status)
	}

	var debits, credits int64
	err = svc.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(CASE WHEN direction = 'DEBIT' THEN amount_minor END), 0),
		       COALESCE(SUM(CASE WHEN direction = 'CREDIT' THEN amount_minor END), 0)
		FROM ledger_entries WHERE reference_id = $1`, order.ID).Scan(&debits, &credits)
	if err != nil {
		t.Fatal(err)
	}
	if debits != 15000 || credits != 15000 {
		t.Fatalf("ledger unbalanced: debit=%d credit=%d", debits, credits)
	}

	// Webhook reconciliation of the same payment is idempotent.
	hook := fmt.Sprintf(`{"eventType":"PAYMENT_STATUS_CHANGED","data":{"paymentKey":"pay-x","orderId":%q,"status":"DONE"}}`, order.ID)
	if err := svc.HandleWebhook(ctx, json.RawMessage(hook)); err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}

	payout, err := svc.RequestPayout(ctx, ownerID, RequestPayoutInput{
		BlogID:           blogID,
		GrossAmountMinor: 13500,
		BankCode:         "088",
		BankAccount:      "110-123-456789",
		AccountHolder:    "홍길동",
	})
	if err != nil {
		t.Fatalf("RequestPayout: %v", err)
	}
	if payout.WithholdingTaxMinor != 445 || payout.NetAmountMinor != 13055 {
		t.Fatalf("payout: %+v", payout)
	}
	if _, _, _, err := svc.RevealPayoutAccount(ctx, payout.ID); err != nil {
		t.Fatalf("RevealPayoutAccount: %v", err)
	}
	decided, err := svc.DecidePayout(ctx, payout.ID, "APPROVE", "")
	if err != nil || decided.Status != "APPROVED" {
		t.Fatalf("DecidePayout: %+v %v", decided, err)
	}
	settled, err := svc.MarkPayoutPaid(ctx, payout.ID)
	if err != nil || settled.Status != "PAID" {
		t.Fatalf("MarkPayoutPaid: %+v %v", settled, err)
	}
	var settlement int64
	if err := svc.pool.QueryRow(ctx, `SELECT COUNT(*) FROM ledger_entries
		WHERE reference_type = 'PAYOUT_SETTLEMENT' AND reference_id = $1`, payout.ID).Scan(&settlement); err != nil {
		t.Fatal(err)
	}
	if settlement != 3 {
		t.Fatalf("settlement entries = %d", settlement)
	}
}
