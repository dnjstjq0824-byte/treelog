package mailer

import (
	"bufio"
	"context"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

// fakeSMTP speaks just enough SMTP to capture one message.
type fakeSMTP struct {
	listener net.Listener
	received chan string
}

func startFakeSMTP(t *testing.T) *fakeSMTP {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	fake := &fakeSMTP{listener: listener, received: make(chan string, 1)}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)
		write := func(line string) {
			_, _ = writer.WriteString(line + "\r\n")
			_ = writer.Flush()
		}
		read := func() string {
			line, _ := reader.ReadString('\n')
			return line
		}
		write("220 fake smtp ready")
		for {
			line := read()
			upper := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(upper, "EHLO") || strings.HasPrefix(upper, "HELO"):
				write("250-fake")
				write("250 HELP")
			case strings.HasPrefix(upper, "MAIL FROM"):
				write("250 ok")
			case strings.HasPrefix(upper, "RCPT TO"):
				write("250 ok")
			case strings.HasPrefix(upper, "DATA"):
				write("354 end with .")
				var body strings.Builder
				for {
					dataLine := read()
					if strings.TrimRight(dataLine, "\r\n") == "." {
						break
					}
					body.WriteString(dataLine)
				}
				fake.received <- body.String()
				write("250 queued")
			case strings.HasPrefix(upper, "QUIT"):
				write("221 bye")
				return
			default:
				write("250 ok")
			}
		}
	}()
	t.Cleanup(func() { _ = listener.Close() })
	return fake
}

func TestSendVerification(t *testing.T) {
	fake := startFakeSMTP(t)
	host, port, _ := net.SplitHostPort(fake.listener.Addr().String())
	m := New(Config{Host: host, Port: mustPort(t, port), From: "no-reply@treeblog.local"})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.SendVerification(ctx, "user@example.com", "123456"); err != nil {
		t.Fatalf("SendVerification: %v", err)
	}
	select {
	case body := <-fake.received:
		for _, want := range []string{"123456", "multipart/alternative", "text/plain", "text/html"} {
			if !strings.Contains(body, want) {
				t.Errorf("message missing %q:\n%s", want, body)
			}
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("no message received")
	}
}

func TestSendPasswordReset(t *testing.T) {
	fake := startFakeSMTP(t)
	host, port, _ := net.SplitHostPort(fake.listener.Addr().String())
	m := New(Config{Host: host, Port: mustPort(t, port), AppBase: "https://blog.example.com"})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.SendPasswordReset(ctx, "user@example.com", "reset-token-abc"); err != nil {
		t.Fatalf("SendPasswordReset: %v", err)
	}
	select {
	case body := <-fake.received:
		if !strings.Contains(body, "https://blog.example.com/reset-password?token=reset-token-abc") {
			t.Errorf("reset link missing:\n%s", body)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("no message received")
	}
}

func TestSendFailsFastOnRefusedConnection(t *testing.T) {
	m := New(Config{Host: "127.0.0.1", Port: 1, Timeout: 2 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.SendVerification(ctx, "user@example.com", "123456"); err == nil {
		t.Fatalf("expected dial error")
	}
}

func mustPort(t *testing.T, raw string) int {
	t.Helper()
	port, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("bad port %q", raw)
	}
	return port
}
