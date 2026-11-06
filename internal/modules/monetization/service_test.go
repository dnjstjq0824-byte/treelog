package monetization

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"testing"

	"treeblog/internal/platform/tosspayments"
)

func testPepper(t *testing.T) []byte {
	t.Helper()
	pepper := make([]byte, 32)
	if _, err := rand.Read(pepper); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return pepper
}

func TestNewServiceValidation(t *testing.T) {
	toss := tosspayments.New(tosspayments.Config{})
	if _, err := NewService(nil, toss, testPepper(t)); err == nil {
		t.Fatalf("nil pool accepted")
	}
}

func TestPlatformFee(t *testing.T) {
	fee, net := platformFee(15000)
	if fee != 1500 || net != 13500 {
		t.Fatalf("fee=%d net=%d", fee, net)
	}
	fee, net = platformFee(99)
	if fee != 9 || net != 90 {
		t.Fatalf("small fee=%d net=%d", fee, net)
	}
}

func TestCreateOrderValidation(t *testing.T) {
	svc, err := NewService(nil, tosspayments.New(tosspayments.Config{}), testPepper(t))
	_ = svc
	if err == nil {
		t.Fatalf("expected pool error")
	}
	bare := &Service{}
	ctx := context.Background()
	if _, err := bare.CreateOrder(ctx, "u", CreateOrderInput{OrderType: "BOGUS", Currency: "KRW", IdempotencyKey: "k"}); !isFault(err, http.StatusBadRequest, "INVALID_ORDER_TYPE") {
		t.Fatalf("bad type: %v", err)
	}
	if _, err := bare.CreateOrder(ctx, "u", CreateOrderInput{OrderType: "READER_DONATION", Currency: "USD", IdempotencyKey: "k"}); !isFault(err, http.StatusBadRequest, "UNSUPPORTED_CURRENCY") {
		t.Fatalf("bad currency: %v", err)
	}
	if _, err := bare.CreateOrder(ctx, "u", CreateOrderInput{OrderType: "READER_DONATION", Currency: "KRW"}); !isFault(err, http.StatusBadRequest, "INVALID_IDEMPOTENCY_KEY") {
		t.Fatalf("missing key: %v", err)
	}
}

func TestAccountEncryptionRoundTrip(t *testing.T) {
	pepper := testPepper(t)
	sealed, err := encryptAccount(pepper, "123-456-789012")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if bytes.Contains(sealed, []byte("123-456")) {
		t.Fatalf("sealed payload leaks the account")
	}
	opened, err := decryptAccount(pepper, sealed)
	if err != nil || opened != "123-456-789012" {
		t.Fatalf("round trip: %q %v", opened, err)
	}
	sealed[len(sealed)-1] ^= 0xff
	if _, err := decryptAccount(pepper, sealed); err == nil {
		t.Fatalf("tampered payload accepted")
	}
	if _, err := decryptAccount(testPepper(t), sealed[:len(sealed)-1]); err == nil {
		t.Fatalf("wrong pepper accepted")
	}
}

func TestMethodCode(t *testing.T) {
	if methodCode("카드") != "CARD" || methodCode("가상계좌") != "VIRTUAL_ACCOUNT" {
		t.Fatalf("method mapping broken")
	}
	if methodCode("") != "CARD" || methodCode("토스페이") != "EASY_PAY" {
		t.Fatalf("method fallback broken")
	}
}

func isFault(err error, status int, code string) bool {
	var fault *Fault
	return errors.As(err, &fault) && fault.Status == status && fault.Code == code
}
