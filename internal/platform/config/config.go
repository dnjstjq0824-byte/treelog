package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr          string
	DatabaseURL       string
	ReplicaURL        string
	ShutdownTimeout   time.Duration
	JWTPrivateKeyPath string
	JWTIssuer         string
	JWTAudience       string
	CodePepper        string
	SMTPHost          string
	SMTPPort          int
	SMTPUsername      string
	SMTPPassword      string
	SMTPFrom          string
	SMTPUseTLS        bool
	AppBaseURL        string
	S3Endpoint        string
	S3Region          string
	S3AccessKey       string
	S3SecretKey       string
	S3BucketRaw       string
	S3BucketPublic    string
	S3PublicBaseURL   string
	OpenSearchURL     string
	OpenSearchPrefix  string
	OAuthRedirectBase string
	OAuthGoogleID     string
	OAuthGoogleSecret string
	OAuthAppleID      string
	OAuthAppleTeamID  string
	OAuthAppleKeyID   string
	OAuthAppleKeyPath string
	OAuthGitHubID     string
	OAuthGitHubSecret string
	OAuthNaverID      string
	OAuthNaverSecret  string
	OAuthKakaoID      string
	OAuthKakaoSecret  string
	OAuthTreescanID     string
	OAuthTreescanSecret string
	OAuthTreescanIssuer string
	TreescanAPIKeys     string
	TossSecretKey       string
	TossClientKey       string
}

func Load() (Config, error) {
	addr := os.Getenv("TREEBLOG_HTTP_ADDR")
	if addr == "" {
		return Config{}, errors.New("TREEBLOG_HTTP_ADDR is required")
	}
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return Config{}, fmt.Errorf("invalid TREEBLOG_HTTP_ADDR: %w", err)
	}
	databaseURL := os.Getenv("TREEBLOG_DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("TREEBLOG_DATABASE_URL is required")
	}
	jwtKeyPath := os.Getenv("TREEBLOG_JWT_PRIVATE_KEY_PATH")
	if jwtKeyPath == "" {
		return Config{}, errors.New("TREEBLOG_JWT_PRIVATE_KEY_PATH is required")
	}
	jwtIssuer := os.Getenv("TREEBLOG_JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = "https://treeblog.local"
	}
	jwtAudience := os.Getenv("TREEBLOG_JWT_AUDIENCE")
	if jwtAudience == "" {
		jwtAudience = "treeblog-web"
	}
	codePepper := os.Getenv("TREEBLOG_CODE_PEPPER")
	if codePepper == "" {
		return Config{}, errors.New("TREEBLOG_CODE_PEPPER is required")
	}
	seconds := 15
	if raw := os.Getenv("TREEBLOG_SHUTDOWN_SECONDS"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 300 {
			return Config{}, errors.New("TREEBLOG_SHUTDOWN_SECONDS must be between 1 and 300")
		}
		seconds = value
	}
	smtpPort := 1025
	if raw := os.Getenv("TREEBLOG_SMTP_PORT"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 65535 {
			return Config{}, errors.New("TREEBLOG_SMTP_PORT must be between 1 and 65535")
		}
		smtpPort = value
	}
	return Config{
		HTTPAddr:          addr,
		DatabaseURL:       databaseURL,
		ReplicaURL:        os.Getenv("TREEBLOG_REPLICA_URL"),
		ShutdownTimeout:   time.Duration(seconds) * time.Second,
		JWTPrivateKeyPath: jwtKeyPath,
		JWTIssuer:         jwtIssuer,
		JWTAudience:       jwtAudience,
		CodePepper:        codePepper,
		SMTPHost:          envOr("TREEBLOG_SMTP_HOST", "localhost"),
		SMTPPort:          smtpPort,
		SMTPUsername:      os.Getenv("TREEBLOG_SMTP_USERNAME"),
		SMTPPassword:      os.Getenv("TREEBLOG_SMTP_PASSWORD"),
		SMTPFrom:          envOr("TREEBLOG_SMTP_FROM", "no-reply@treeblog.local"),
		SMTPUseTLS:        os.Getenv("TREEBLOG_SMTP_USE_TLS") == "true",
		AppBaseURL:        envOr("TREEBLOG_APP_BASE_URL", "http://localhost:3000"),
		S3Endpoint:        envOr("TREEBLOG_S3_ENDPOINT", "http://localhost:9000"),
		S3Region:          envOr("TREEBLOG_S3_REGION", "us-east-1"),
		S3AccessKey:       os.Getenv("TREEBLOG_S3_ACCESS_KEY"),
		S3SecretKey:       os.Getenv("TREEBLOG_S3_SECRET_KEY"),
		S3BucketRaw:       envOr("TREEBLOG_S3_BUCKET_RAW", "treeblog-raw-private"),
		S3BucketPublic:    envOr("TREEBLOG_S3_BUCKET_PUBLIC", "treeblog-media-public"),
		S3PublicBaseURL:   os.Getenv("TREEBLOG_S3_PUBLIC_BASE_URL"),
		OpenSearchURL:       os.Getenv("TREEBLOG_OPENSEARCH_URL"),
		OpenSearchPrefix:    envOr("TREEBLOG_OPENSEARCH_PREFIX", "treeblog"),
		OAuthRedirectBase:   envOr("TREEBLOG_OAUTH_REDIRECT_BASE", "http://localhost:8080"),
		OAuthGoogleID:       os.Getenv("TREEBLOG_OAUTH_GOOGLE_CLIENT_ID"),
		OAuthGoogleSecret:   os.Getenv("TREEBLOG_OAUTH_GOOGLE_CLIENT_SECRET"),
		OAuthAppleID:        os.Getenv("TREEBLOG_OAUTH_APPLE_CLIENT_ID"),
		OAuthAppleTeamID:    os.Getenv("TREEBLOG_OAUTH_APPLE_TEAM_ID"),
		OAuthAppleKeyID:     os.Getenv("TREEBLOG_OAUTH_APPLE_KEY_ID"),
		OAuthAppleKeyPath:   os.Getenv("TREEBLOG_OAUTH_APPLE_KEY_PATH"),
		OAuthGitHubID:       os.Getenv("TREEBLOG_OAUTH_GITHUB_CLIENT_ID"),
		OAuthGitHubSecret:   os.Getenv("TREEBLOG_OAUTH_GITHUB_CLIENT_SECRET"),
		OAuthNaverID:        os.Getenv("TREEBLOG_OAUTH_NAVER_CLIENT_ID"),
		OAuthNaverSecret:    os.Getenv("TREEBLOG_OAUTH_NAVER_CLIENT_SECRET"),
		OAuthKakaoID:        os.Getenv("TREEBLOG_OAUTH_KAKAO_CLIENT_ID"),
		OAuthKakaoSecret:    os.Getenv("TREEBLOG_OAUTH_KAKAO_CLIENT_SECRET"),
		OAuthTreescanID:     os.Getenv("TREEBLOG_OAUTH_TREESCAN_CLIENT_ID"),
		OAuthTreescanSecret: os.Getenv("TREEBLOG_OAUTH_TREESCAN_CLIENT_SECRET"),
		OAuthTreescanIssuer: os.Getenv("TREEBLOG_OAUTH_TREESCAN_ISSUER"),
		TreescanAPIKeys:     os.Getenv("TREEBLOG_TREESCAN_API_KEYS"),
		TossSecretKey:       os.Getenv("TREEBLOG_TOSS_SECRET_KEY"),
		TossClientKey:       os.Getenv("TREEBLOG_TOSS_CLIENT_KEY"),
	}, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
