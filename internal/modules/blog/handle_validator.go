package blog

import (
	"regexp"
	"strings"
)

var handleRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$`)

var reservedHandles = map[string]struct{}{
	"about": {}, "account": {}, "admin": {}, "administrator": {}, "ads": {},
	"affiliate": {}, "analytics": {}, "api": {}, "app": {}, "assets": {},
	"auth": {}, "backup": {}, "billing": {}, "blog": {}, "blogs": {},
	"board": {}, "bot": {}, "campaign": {}, "cdn": {}, "channel": {},
	"checkout": {}, "claim": {}, "community": {}, "config": {}, "connect": {},
	"console": {}, "contact": {}, "dashboard": {}, "database": {}, "default": {},
	"delete": {}, "dev": {}, "developer": {}, "discovery": {}, "dns": {},
	"docs": {}, "download": {}, "edge": {}, "edit": {}, "editor": {},
	"email": {}, "enterprise": {}, "event": {}, "events": {}, "explore": {},
	"export": {}, "faq": {}, "feed": {}, "feeds": {}, "file": {},
	"files": {}, "finance": {}, "forum": {}, "ftp": {}, "gateway": {},
	"graph": {}, "guest": {}, "guestbook": {}, "help": {}, "home": {},
	"image": {}, "images": {}, "imap": {}, "import": {}, "inbox": {},
	"index": {}, "info": {}, "insight": {}, "internal": {}, "invite": {},
	"job": {}, "jobs": {}, "legal": {}, "live": {}, "login": {},
	"logout": {}, "mail": {}, "master": {}, "media": {}, "member": {},
	"members": {}, "membership": {}, "message": {}, "metrics": {}, "moderation": {},
	"my": {}, "network": {}, "news": {}, "notice": {}, "notification": {},
	"notifications": {}, "ns": {}, "ns1": {}, "ns2": {}, "oauth": {},
	"official": {}, "opensearch": {}, "order": {}, "orders": {}, "page": {},
	"pages": {}, "partner": {}, "pay": {}, "payment": {}, "payments": {},
	"payout": {}, "ping": {}, "policy": {}, "pop": {}, "popular": {},
	"portal": {}, "post": {}, "posts": {}, "pricing": {}, "privacy": {},
	"profile": {}, "public": {}, "queue": {}, "rank": {}, "ranking": {},
	"ready": {}, "register": {}, "release": {}, "report": {}, "root": {},
	"router": {}, "rss": {}, "rule": {}, "rules": {}, "sandbox": {},
	"search": {}, "security": {}, "server": {}, "service": {}, "services": {},
	"session": {}, "sessions": {}, "setting": {}, "settings": {}, "signin": {},
	"signup": {}, "sitemap": {}, "smtp": {}, "ssl": {}, "staff": {},
	"staging": {}, "static": {}, "status": {}, "store": {}, "studio": {},
	"subdomain": {}, "subscribe": {}, "support": {}, "sync": {}, "system": {},
	"tag": {}, "tags": {}, "team": {}, "terms": {}, "test": {},
	"testing": {}, "theme": {}, "themes": {}, "tls": {}, "token": {},
	"topic": {}, "topics": {}, "trace": {}, "transaction": {}, "treeblog": {},
	"treescan": {}, "treewiki": {}, "trend": {}, "trending": {}, "upload": {},
	"uploads": {}, "user": {}, "users": {}, "verification": {}, "verify": {},
	"video": {}, "videos": {}, "view": {}, "views": {}, "wallet": {},
	"web": {}, "webhook": {}, "webhooks": {}, "wiki": {}, "www": {},
}

// ValidateHandle checks if a handle meets length, character, and reservation constraints.
func ValidateHandle(handle string) (bool, string) {
	handle = strings.ToLower(strings.TrimSpace(handle))
	if len(handle) < 3 || len(handle) > 30 {
		return false, "Handle must be between 3 and 30 characters"
	}
	if !handleRegex.MatchString(handle) {
		return false, "Handle can only contain lowercase letters, numbers, and hyphens (cannot start or end with hyphen)"
	}
	if _, isReserved := reservedHandles[handle]; isReserved {
		return false, "Handle is a reserved platform keyword"
	}
	return true, ""
}
