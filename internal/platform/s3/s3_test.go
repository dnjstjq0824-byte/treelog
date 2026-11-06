package s3

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"
)

func testClient(t *testing.T) *Client {
	t.Helper()
	client, err := New(Config{
		Endpoint:  "https://s3.example.com",
		Region:    "us-east-1",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client
}

func TestNewRejectsMissingCredentials(t *testing.T) {
	if _, err := New(Config{Endpoint: "http://localhost:9000", Region: "us-east-1"}); err == nil {
		t.Fatalf("anonymous config accepted")
	}
	if _, err := New(Config{Endpoint: "://bad", Region: "us-east-1", AccessKey: "a", SecretKey: "b"}); err == nil {
		t.Fatalf("bad endpoint accepted")
	}
}

func TestPresignStructure(t *testing.T) {
	client := testClient(t)
	fixed := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	client.now = func() time.Time { return fixed }

	first, err := client.PresignPUT("treeblog-raw-private", "uploads/b 1/한글.png", "image/png", 30*time.Minute)
	if err != nil {
		t.Fatalf("PresignPUT: %v", err)
	}
	second, err := client.PresignPUT("treeblog-raw-private", "uploads/b 1/한글.png", "image/png", 30*time.Minute)
	if err != nil {
		t.Fatalf("PresignPUT: %v", err)
	}
	if first != second {
		t.Fatalf("presign is not deterministic for a fixed clock")
	}

	parsed, err := url.Parse(first)
	if err != nil {
		t.Fatalf("parse presigned url: %v", err)
	}
	if parsed.Host != "s3.example.com" {
		t.Errorf("host = %s", parsed.Host)
	}
	if !strings.HasPrefix(parsed.EscapedPath(), "/treeblog-raw-private/uploads/b%201/") {
		t.Errorf("path not segment-encoded: %s", parsed.EscapedPath())
	}
	if strings.Contains(parsed.EscapedPath(), "%25") {
		t.Errorf("path double-encoded: %s", parsed.EscapedPath())
	}
	query := parsed.Query()
	if query.Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" {
		t.Errorf("missing algorithm param")
	}
	if query.Get("X-Amz-Expires") != "1800" {
		t.Errorf("expires = %s", query.Get("X-Amz-Expires"))
	}
	if matched, _ := regexp.MatchString(`^[0-9a-f]{64}$`, query.Get("X-Amz-Signature")); !matched {
		t.Errorf("bad signature: %q", query.Get("X-Amz-Signature"))
	}
	if !strings.Contains(query.Get("X-Amz-Credential"), "20260524/us-east-1/s3/aws4_request") {
		t.Errorf("bad credential scope: %q", query.Get("X-Amz-Credential"))
	}
	if _, err := client.PresignGET("b", "k", 8*24*time.Hour); err == nil {
		t.Errorf("over-long expiry accepted")
	}
}

func TestHeaderAuthRoundTrip(t *testing.T) {
	var gotAuth, gotSHA, gotDate, gotRange string
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotSHA = r.Header.Get("X-Amz-Content-Sha256")
		gotDate = r.Header.Get("X-Amz-Date")
		gotRange = r.Header.Get("Range")
		gotPath = r.URL.EscapedPath()
		body := []byte("0123456789abcdef")
		switch r.Method {
		case http.MethodHead:
			w.Header().Set("Content-Length", "16")
			w.WriteHeader(http.StatusOK)
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		default:
			if gotRange != "" {
				w.WriteHeader(http.StatusPartialContent)
				_, _ = w.Write(body[:4])
				return
			}
			_, _ = w.Write(body)
		}
	}))
	t.Cleanup(server.Close)

	client := testClient(t)
	client.base, _ = url.Parse(server.URL)
	ctx := context.Background()

	size, err := client.Head(ctx, "b", "k")
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if size != 16 {
		t.Fatalf("size = %d", size)
	}
	if !strings.HasPrefix(gotAuth, "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/") {
		t.Errorf("bad authorization: %q", gotAuth)
	}
	if !strings.Contains(gotAuth, "SignedHeaders=host;x-amz-content-sha256;x-amz-date") {
		t.Errorf("bad signed headers: %q", gotAuth)
	}
	if matched, _ := regexp.MatchString(`^[0-9a-f]{64}$`, gotSHA); gotSHA == "" || !matched {
		t.Errorf("bad payload hash header: %q", gotSHA)
	}
	if gotDate == "" || gotPath != "/b/k" {
		t.Errorf("date/path = %q %q", gotDate, gotPath)
	}

	part, err := client.GetRange(ctx, "b", "k", 0, 4)
	if err != nil {
		t.Fatalf("GetRange: %v", err)
	}
	if string(part) != "0123" || gotRange != "bytes=0-3" {
		t.Errorf("range = %q header %q", part, gotRange)
	}

	if err := client.Put(ctx, "b", "k", "image/jpeg", []byte("img")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	if _, err := client.Get(ctx, "b", "k", 4); err == nil {
		t.Fatalf("over-cap Get accepted")
	}
}
