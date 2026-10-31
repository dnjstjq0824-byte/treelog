package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"treeblog/internal/db"
	"treeblog/internal/modules/blog"
	"treeblog/internal/modules/identity"
	"treeblog/internal/modules/user"
	"treeblog/internal/platform/config"
	"treeblog/internal/platform/httpserver"
)

type loggingMailer struct{ logger *slog.Logger }

func (m *loggingMailer) SendVerification(ctx context.Context, email, code string) error {
	m.logger.Info("verification email", "email", email, "code", code)
	return nil
}

func (m *loggingMailer) SendPasswordReset(ctx context.Context, email, token string) error {
	m.logger.Info("password reset email", "email", email, "token", token)
	return nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pools, err := db.Open(context.Background(), cfg.DatabaseURL, cfg.ReplicaURL)
	if err != nil {
		return err
	}
	defer pools.Close()

	signer, err := identity.LoadTokenSigner(cfg.JWTPrivateKeyPath, cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		return err
	}

	pepper, err := base64.StdEncoding.DecodeString(cfg.CodePepper)
	if err != nil || len(pepper) < 32 {
		return errors.New("TREEBLOG_CODE_PEPPER must be base64-encoded and at least 32 bytes")
	}

	mailer := &loggingMailer{logger: logger}
	identityService, err := identity.NewService(pools.Primary, signer, mailer, pepper)
	if err != nil {
		return err
	}
	userService := user.NewService(pools.Primary)
	blogService := blog.NewService(pools.Primary)

	modules := []httpserver.Module{
		identity.NewHandler(identityService),
		user.NewHandler(userService),
		blog.NewHandler(blogService),
	}

	server, err := httpserver.New(cfg.HTTPAddr, modules, pools.Primary.Ping, logger)
	if err != nil {
		return err
	}
	// Populate request context with authenticated user information.
	server.HTTP.Handler = identity.AuthenticateMiddleware(signer, pools.Primary)(server.HTTP.Handler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() { errCh <- server.HTTP.ListenAndServe() }()
	logger.Info("api listening", "addr", cfg.HTTPAddr)
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.HTTP.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}
