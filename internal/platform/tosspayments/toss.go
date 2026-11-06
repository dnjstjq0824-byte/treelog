// Package tosspayments is a minimal Toss Payments client using only the
// standard library (confirm + lookup for webhook reconciliation).
package tosspayments

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"treeblog/internal/pkg/safenet"
)

// Config holds the Toss secret key.
type Config struct {
	SecretKey string
	BaseURL   string
}

// Client calls the Toss Payments API.
type Client struct {
	cfg Config
	http *http.Client
	// HTTPClient overrides the default SSRF-safe client (tests only).
	HTTPClient *http.Client
}

func (c *Client) transport() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return c.http
}

// New creates a Client.
func New(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.tosspayments.com"
	}
	return &Client{cfg: cfg, http: safenet.NewSafeHTTPClient(15 * time.Second)}
}

// Enabled reports whether live calls can be made.
func (c *Client) Enabled() bool {
	return c != nil && c.cfg.SecretKey != ""
}

// PaymentInfo is the normalized payment state.
type PaymentInfo struct {
	PaymentKey string
	OrderID    string
	Status     string
	Method     string
	AmountMinor int64
	Currency   string
	ApprovedAt *time.Time
	Raw        json.RawMessage
}

func (c *Client) authHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(c.cfg.SecretKey+":"))
}

func (c *Client) call(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	if !c.Enabled() {
		return nil, errors.New("toss payments is not configured")
	}
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, method, c.cfg.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.authHeader())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.transport().Do(req)
	if err != nil {
		return nil, fmt.Errorf("toss %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var failure struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(payload, &failure)
		if failure.Code == "" {
			failure.Code = "TOSS_REQUEST_FAILED"
		}
		return nil, &Error{Code: failure.Code, Message: failure.Message, Status: resp.StatusCode}
	}
	return json.RawMessage(payload), nil
}

// Error is a Toss API failure response.
type Error struct {
	Code    string
	Message string
	Status  int
}

func (e *Error) Error() string {
	return "toss " + e.Code + ": " + e.Message
}

// ConfirmPayment approves a payment. Toss verifies amount/orderId itself;
// callers must still compare against the local order first.
func (c *Client) ConfirmPayment(ctx context.Context, paymentKey, orderID string, amountMinor int64) (*PaymentInfo, error) {
	raw, err := c.call(ctx, http.MethodPost, "/v1/payments/confirm", map[string]any{
		"paymentKey": paymentKey,
		"orderId":    orderID,
		"amount":     amountMinor,
	})
	if err != nil {
		return nil, err
	}
	return parsePayment(raw)
}

// GetPayment fetches a payment by key (used for webhook reconciliation).
func (c *Client) GetPayment(ctx context.Context, paymentKey string) (*PaymentInfo, error) {
	raw, err := c.call(ctx, http.MethodGet, "/v1/payments/"+paymentKey, nil)
	if err != nil {
		return nil, err
	}
	return parsePayment(raw)
}

func parsePayment(raw json.RawMessage) (*PaymentInfo, error) {
	var doc struct {
		PaymentKey string `json:"paymentKey"`
		OrderID    string `json:"orderId"`
		Status     string `json:"status"`
		Method     string `json:"method"`
		Currency   string `json:"currency"`
		ApprovedAt *time.Time `json:"approvedAt"`
		TotalAmount int64  `json:"totalAmount"`
		BalanceAmount int64 `json:"balanceAmount"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("invalid toss payment payload: %w", err)
	}
	if doc.PaymentKey == "" || doc.OrderID == "" {
		return nil, errors.New("toss payment payload missing key/order")
	}
	return &PaymentInfo{
		PaymentKey: doc.PaymentKey, OrderID: doc.OrderID, Status: doc.Status,
		Method: doc.Method, AmountMinor: doc.TotalAmount, Currency: doc.Currency,
		ApprovedAt: doc.ApprovedAt, Raw: raw,
	}, nil
}
