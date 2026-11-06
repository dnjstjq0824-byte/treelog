package tosspayments

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDisabledClient(t *testing.T) {
	client := New(Config{})
	if client.Enabled() {
		t.Fatalf("empty secret reports enabled")
	}
	if _, err := client.ConfirmPayment(context.Background(), "k", "o", 1000); err == nil {
		t.Fatalf("disabled confirm accepted")
	}
}

func TestConfirmPayment(t *testing.T) {
	var gotAuth, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		gotBody = string(raw)
		if r.URL.Path != "/v1/payments/confirm" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paymentKey":"key-1","orderId":"order-1","status":"DONE","method":"카드","currency":"KRW","totalAmount":15000,"approvedAt":"2026-05-01T12:00:00+09:00"}`))
	}))
	t.Cleanup(server.Close)

	client := New(Config{SecretKey: "test-secret", BaseURL: server.URL})
	client.HTTPClient = server.Client()
	info, err := client.ConfirmPayment(context.Background(), "key-1", "order-1", 15000)
	if err != nil {
		t.Fatalf("ConfirmPayment: %v", err)
	}
	if info.Status != "DONE" || info.AmountMinor != 15000 || info.Method != "카드" {
		t.Fatalf("info: %+v", info)
	}
	if gotAuth != "Basic "+base64.StdEncoding.EncodeToString([]byte("test-secret:")) {
		t.Fatalf("auth = %q", gotAuth)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(gotBody), &sent); err != nil {
		t.Fatalf("request not json: %s", gotBody)
	}
	if sent["paymentKey"] != "key-1" || sent["orderId"] != "order-1" || sent["amount"] != float64(15000) {
		t.Fatalf("sent = %v", sent)
	}
}

func TestConfirmSurfacesTossErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"INVALID_CARD_NUMBER","message":"카드 번호가 올바르지 않습니다."}`))
	}))
	t.Cleanup(server.Close)

	client := New(Config{SecretKey: "s", BaseURL: server.URL})
	client.HTTPClient = server.Client()
	_, err := client.ConfirmPayment(context.Background(), "k", "o", 1)
	var tossErr *Error
	if !errors.As(err, &tossErr) || tossErr.Code != "INVALID_CARD_NUMBER" {
		t.Fatalf("err = %v", err)
	}
}

func TestGetPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/payments/key-9") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"paymentKey":"key-9","orderId":"order-9","status":"CANCELED","method":"가상계좌","currency":"KRW","totalAmount":5000}`))
	}))
	t.Cleanup(server.Close)

	client := New(Config{SecretKey: "s", BaseURL: server.URL})
	client.HTTPClient = server.Client()
	info, err := client.GetPayment(context.Background(), "key-9")
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	if info.Status != "CANCELED" || info.OrderID != "order-9" {
		t.Fatalf("info: %+v", info)
	}
}
