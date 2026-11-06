package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"treeblog/internal/db"
	"treeblog/internal/modules/admin"
	"treeblog/internal/modules/blog"
	"treeblog/internal/modules/community"
	"treeblog/internal/modules/domain"
	"treeblog/internal/modules/editor"
	"treeblog/internal/modules/identity"
	"treeblog/internal/modules/media"
	"treeblog/internal/modules/moderation"
	"treeblog/internal/modules/monetization"
	"treeblog/internal/modules/notification"
	"treeblog/internal/modules/post"
	"treeblog/internal/modules/search"
	"treeblog/internal/modules/seo"
	"treeblog/internal/modules/theme"
	"treeblog/internal/modules/treescan"
	"treeblog/internal/modules/user"
	"treeblog/internal/pkg/security"
	"treeblog/internal/platform/config"
	"treeblog/internal/platform/httpserver"
	"treeblog/internal/platform/mailer"
	"treeblog/internal/platform/opensearch"
	s3client "treeblog/internal/platform/s3"
	"treeblog/internal/platform/tosspayments"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

// splitAPIKeys parses a comma-separated API key list.
func splitAPIKeys(raw string) []string {
	keys := []string{}
	for _, key := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(key); trimmed != "" {
			keys = append(keys, trimmed)
		}
	}
	return keys
}

// loadOAuthConfigs builds per-provider social login configs. Providers
// without a client ID stay unconfigured and answer 501 at runtime.
func loadOAuthConfigs(cfg config.Config) (map[identity.OAuthProvider]identity.OAuthClientConfig, error) {
	configs := make(map[identity.OAuthProvider]identity.OAuthClientConfig)
	add := func(provider identity.OAuthProvider, id, secret string, extra func(*identity.OAuthClientConfig)) {
		if id == "" {
			return
		}
		oauthCfg := identity.OAuthClientConfig{
			ClientID:     id,
			ClientSecret: secret,
			RedirectBase: cfg.OAuthRedirectBase,
			AppBaseURL:   cfg.AppBaseURL,
		}
		if extra != nil {
			extra(&oauthCfg)
		}
		configs[provider] = oauthCfg
	}
	add(identity.OAuthGoogle, cfg.OAuthGoogleID, cfg.OAuthGoogleSecret, nil)
	add(identity.OAuthGitHub, cfg.OAuthGitHubID, cfg.OAuthGitHubSecret, nil)
	add(identity.OAuthNaver, cfg.OAuthNaverID, cfg.OAuthNaverSecret, nil)
	add(identity.OAuthKakao, cfg.OAuthKakaoID, cfg.OAuthKakaoSecret, nil)
	add(identity.OAuthTreescanOIDC, cfg.OAuthTreescanID, cfg.OAuthTreescanSecret, func(oauthCfg *identity.OAuthClientConfig) {
		oauthCfg.TreescanIssuer = cfg.OAuthTreescanIssuer
	})
	if cfg.OAuthAppleID != "" {
		keyPEM := []byte{}
		if cfg.OAuthAppleKeyPath != "" {
			raw, err := os.ReadFile(cfg.OAuthAppleKeyPath)
			if err != nil {
				return nil, err
			}
			keyPEM = raw
		}
		configs[identity.OAuthApple] = identity.OAuthClientConfig{
			ClientID:     cfg.OAuthAppleID,
			RedirectBase: cfg.OAuthRedirectBase,
			AppBaseURL:   cfg.AppBaseURL,
			AppleTeamID:  cfg.OAuthAppleTeamID,
			AppleKeyID:   cfg.OAuthAppleKeyID,
			AppleKeyPEM:  string(keyPEM),
		}
	}
	return configs, nil
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

	mailSender := mailer.New(mailer.Config{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUsername,
		Password: cfg.SMTPPassword,
		From:     cfg.SMTPFrom,
		FromName: "TreeBlog",
		UseTLS:   cfg.SMTPUseTLS,
		AppBase:  cfg.AppBaseURL,
	})
	identityService, err := identity.NewService(pools.Primary, signer, mailSender, pepper)
	if err != nil {
		return err
	}
	oauthConfigs, err := loadOAuthConfigs(cfg)
	if err != nil {
		return err
	}
	identityService.SetOAuthProviders(oauthConfigs)
	userService := user.NewService(pools.Primary)
	adminService, err := admin.NewService(pools.Primary)
	if err != nil {
		return err
	}
	blogService := blog.NewService(pools.Primary)
	communityService, err := community.NewService(pools.Primary)
	if err != nil {
		return err
	}
	domainService := domain.NewService(pools.Primary)
	mediaService, err := media.NewService(pools.Primary, media.StorageConfig{
		S3: s3client.Config{
			Endpoint:  cfg.S3Endpoint,
			Region:    cfg.S3Region,
			AccessKey: cfg.S3AccessKey,
			SecretKey: cfg.S3SecretKey,
		},
		BucketRaw:    cfg.S3BucketRaw,
		BucketPublic: cfg.S3BucketPublic,
	})
	if err != nil {
		return err
	}
	editorService := editor.NewService(pools.Primary)
	postService := post.NewService(pools.Primary)
	moderationService, err := moderation.NewService(pools.Primary)
	if err != nil {
		return err
	}
	monetizationService, err := monetization.NewService(pools.Primary,
		tosspayments.New(tosspayments.Config{SecretKey: cfg.TossSecretKey}), pepper)
	if err != nil {
		return err
	}
	notificationService, err := notification.NewService(pools.Primary)
	if err != nil {
		return err
	}
	themeService := theme.NewService(pools.Primary)
	searchService := search.NewService(pools.Primary, opensearch.New(opensearch.Config{
		Address:     cfg.OpenSearchURL,
		IndexPrefix: cfg.OpenSearchPrefix,
	}))
	seoService, err := seo.NewService(pools.Primary, cfg.AppBaseURL)
	if err != nil {
		return err
	}
	treescanService, err := treescan.NewService(pools.Primary)
	if err != nil {
		return err
	}

	modules := []httpserver.Module{
		identity.NewHandler(identityService),
		user.NewHandler(userService),
		admin.NewHandler(adminService),
		blog.NewHandler(blogService),
		community.NewHandler(communityService),
		domain.NewHandler(domainService),
		media.NewHandler(mediaService),
		editor.NewHandler(editorService),
		post.NewHandler(postService),
		moderation.NewHandler(moderationService),
		monetization.NewHandler(monetizationService, cfg.TossClientKey),
		notification.NewHandler(notificationService),
		theme.NewHandler(themeService),
		search.NewHandler(searchService),
		seo.NewHandler(seoService),
		treescan.NewHandler(treescanService, splitAPIKeys(cfg.TreescanAPIKeys)),
	}

	rateLimiter := security.NewMemoryRateLimiter()
	server, err := httpserver.New(cfg.HTTPAddr, modules, pools.Primary.Ping, logger,
		httpserver.WithGlobalMiddleware(
			security.SecurityHeadersMiddleware(),
			security.RateLimitMiddleware(rateLimiter, 600, time.Minute, "global"),
			security.CSRFMiddleware(
				// Pre-session endpoints: no session exists yet to protect.
				"/api/v1/auth/register",
				"/api/v1/auth/verify-email",
				"/api/v1/auth/login",
				"/api/v1/auth/password/",
				"/api/v1/auth/oauth/",
				// Third-party callbacks authenticated by signature instead.
				"/api/v1/billing/webhooks/",
			),
			// Populate request context with authenticated user information.
			identity.AuthenticateMiddleware(signer, pools.Primary),
		),
	)
	if err != nil {
		return err
	}

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
