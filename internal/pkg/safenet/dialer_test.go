package safenet

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		ip      string
		blocked bool
	}{
		{"127.0.0.1", true},
		{"127.0.0.2", true},
		{"10.0.0.1", true},
		{"10.254.1.1", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true}, // AWS/GCP/Azure metadata IP
		{"169.254.1.1", true},
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"::1", true},
		{"fe80::1", true},
		{"fc00::1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"93.184.216.34", false}, // example.com
	}

	for _, tc := range tests {
		ip := net.ParseIP(tc.ip)
		if ip == nil {
			t.Fatalf("failed to parse IP: %s", tc.ip)
		}
		got := IsBlockedIP(ip)
		if got != tc.blocked {
			t.Errorf("IsBlockedIP(%s) = %v; want %v", tc.ip, got, tc.blocked)
		}
	}
}

func TestSafeDialer_BlocksMetadataAndLoopback(t *testing.T) {
	dialer := NewSafeDialer(100 * time.Millisecond)

	targets := []string{
		"127.0.0.1:80",
		"127.0.0.1:8080",
		"169.254.169.254:80",
		"10.0.0.1:443",
		"192.168.0.1:80",
	}

	for _, target := range targets {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		conn, err := dialer.DialContext(ctx, "tcp", target)
		cancel()
		if conn != nil {
			conn.Close()
			t.Fatalf("expected dial to %s to be blocked, but connection succeeded", target)
		}
		if err == nil {
			t.Fatalf("expected error dialing %s, got nil", target)
		}
	}
}

