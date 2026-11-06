// Package mailer delivers TreeBlog transactional email over SMTP.
package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// Config describes the SMTP relay. Mailpit-compatible defaults let local
// Compose stacks work without credentials.
type Config struct {
	Host      string
	Port      int
	Username  string
	Password  string
	From      string
	FromName  string
	UseTLS    bool
	AppBase   string
	Timeout   time.Duration
}

// DefaultConfig returns local-development settings (Mailpit on localhost).
func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     1025,
		From:     "no-reply@treeblog.local",
		FromName: "TreeBlog",
		AppBase:  "http://localhost:3000",
		Timeout:  10 * time.Second,
	}
}

// Mailer sends verification and password-reset messages.
type Mailer struct {
	cfg Config
}

// New creates a Mailer. Empty host/port fall back to DefaultConfig values.
func New(cfg Config) *Mailer {
	def := DefaultConfig()
	if cfg.Host == "" {
		cfg.Host = def.Host
	}
	if cfg.Port == 0 {
		cfg.Port = def.Port
	}
	if cfg.From == "" {
		cfg.From = def.From
	}
	if cfg.FromName == "" {
		cfg.FromName = def.FromName
	}
	if cfg.AppBase == "" {
		cfg.AppBase = def.AppBase
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = def.Timeout
	}
	return &Mailer{cfg: cfg}
}

// SendVerification delivers the 6-digit email verification code.
func (m *Mailer) SendVerification(ctx context.Context, email, code string) error {
	subject := "TreeBlog 이메일 인증 코드"
	text := fmt.Sprintf("안녕하세요, TreeBlog입니다.\n\n이메일 인증 코드: %s\n\n10분 이내에 입력해주세요.\n", code)
	html := fmt.Sprintf(`<p>안녕하세요, TreeBlog입니다.</p><p>이메일 인증 코드: <strong style="font-size:20px">%s</strong></p><p>10분 이내에 입력해주세요.</p>`, code)
	return m.send(ctx, email, subject, text, html)
}

// SendPasswordReset delivers the password-reset link.
func (m *Mailer) SendPasswordReset(ctx context.Context, email, token string) error {
	subject := "TreeBlog 비밀번호 재설정"
	link := strings.TrimRight(m.cfg.AppBase, "/") + "/reset-password?token=" + token
	text := fmt.Sprintf("안녕하세요, TreeBlog입니다.\n\n아래 링크에서 1시간 이내에 비밀번호를 재설정해주세요.\n%s\n\n본인이 요청하지 않았다면 이 메일을 무시하세요.\n", link)
	html := fmt.Sprintf(`<p>안녕하세요, TreeBlog입니다.</p><p><a href="%s">비밀번호 재설정하기</a> (1시간 유효)</p><p>본인이 요청하지 않았다면 이 메일을 무시하세요.</p>`, link)
	return m.send(ctx, email, subject, text, html)
}

func (m *Mailer) send(ctx context.Context, to, subject, text, html string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	addr := net.JoinHostPort(m.cfg.Host, strconv.Itoa(m.cfg.Port))
	dialer := &net.Dialer{Timeout: m.cfg.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial %s: %w", addr, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp handshake: %w", err)
	}
	defer client.Close()

	if m.cfg.UseTLS {
		tlsCfg := &tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}
	if m.cfg.Username != "" {
		auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(m.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	message := buildMessage(m.cfg.From, m.cfg.FromName, to, subject, text, html)
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp commit: %w", err)
	}
	return client.Quit()
}

func buildMessage(from, fromName, to, subject, text, html string) string {
	boundary := "treeblog-mail-boundary"
	var sb strings.Builder
	fmt.Fprintf(&sb, "From: %s <%s>\r\n", fromName, from)
	fmt.Fprintf(&sb, "To: <%s>\r\n", to)
	fmt.Fprintf(&sb, "Subject: %s\r\n", subject)
	sb.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&sb, "Content-Type: multipart/alternative; boundary=%s\r\n", boundary)
	sb.WriteString("\r\n")
	fmt.Fprintf(&sb, "--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", boundary, text)
	fmt.Fprintf(&sb, "--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n", boundary, html)
	fmt.Fprintf(&sb, "--%s--\r\n", boundary)
	return sb.String()
}
