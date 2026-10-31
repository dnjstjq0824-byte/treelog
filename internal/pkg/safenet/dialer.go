package safenet

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"syscall"
	"time"
)

var (
	ErrBlockedIP    = errors.New("connection to private/loopback/metadata IP is blocked")
	ErrNoAllowedIPs = errors.New("no routable public IP addresses found")
)

var (
	blockedCIDRsOnce sync.Once
	blockedIPv4Nets  []*net.IPNet
	blockedIPv6Nets  []*net.IPNet
)

func initBlockedCIDRs() {
	rawIPv4CIDRs := []string{
		"0.0.0.0/8",
		"10.0.0.0/8",
		"100.64.0.0/10",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"192.88.99.0/24",
		"192.168.0.0/16",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"224.0.0.0/4",
		"240.0.0.0/4",
		"255.255.255.255/32",
	}
	for _, cidr := range rawIPv4CIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			blockedIPv4Nets = append(blockedIPv4Nets, ipNet)
		}
	}

	rawIPv6CIDRs := []string{
		"::/128",
		"::1/128",
		"::ffff:0:0/96",
		"100::/64",
		"2001:db8::/32",
		"fc00::/7",
		"fe80::/10",
		"ff00::/8",
	}
	for _, cidr := range rawIPv6CIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			blockedIPv6Nets = append(blockedIPv6Nets, ipNet)
		}
	}
}

// IsBlockedIP checks if an IP belongs to private, loopback, multicast, or metadata ranges.
func IsBlockedIP(ip net.IP) bool {
	blockedCIDRsOnce.Do(initBlockedCIDRs)
	if ip == nil {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		for _, netRange := range blockedIPv4Nets {
			if netRange.Contains(v4) {
				return true
			}
		}
		return false
	}
	for _, netRange := range blockedIPv6Nets {
		if netRange.Contains(ip) {
			return true
		}
	}
	return false
}

// NewSafeDialer returns a net.Dialer that prevents SSRF by validating destination IPs
// both during DNS lookup and immediately before socket connection via Control hook.
func NewSafeDialer(timeout time.Duration) *net.Dialer {
	return &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if ip != nil && IsBlockedIP(ip) {
				return fmt.Errorf("%w: %s", ErrBlockedIP, host)
			}
			return nil
		},
	}
}

// NewSafeTransport creates an http.Transport configured with SSRF protection.
func NewSafeTransport(timeout time.Duration) *http.Transport {
	dialer := NewSafeDialer(timeout)
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}
			var allowedIP net.IP
			for _, ip := range ips {
				if !IsBlockedIP(ip) {
					allowedIP = ip
					break
				}
			}
			if allowedIP == nil {
				return nil, fmt.Errorf("%w for host %s", ErrNoAllowedIPs, host)
			}
			targetAddr := net.JoinHostPort(allowedIP.String(), port)
			return dialer.DialContext(ctx, network, targetAddr)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// NewSafeHTTPClient creates an http.Client with SSRF prevention and safe redirect policy.
func NewSafeHTTPClient(timeout time.Duration) *http.Client {
	transport := NewSafeTransport(timeout)
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			host := req.URL.Hostname()
			ip := net.ParseIP(host)
			if ip != nil && IsBlockedIP(ip) {
				return fmt.Errorf("%w in redirect to %s", ErrBlockedIP, host)
			}
			return nil
		},
	}
}
