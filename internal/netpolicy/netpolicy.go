// Package netpolicy blocks outbound requests to non-public addresses (SSRF).
package netpolicy

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// ErrBlocked is returned when a URL or resolved address is not allowed.
var ErrBlocked = errors.New("blocked target")

// insecureSkip disables blocking; for integration tests with loopback httptest only.
var insecureSkip bool

// SetInsecureSkipForTesting enables loopback/private fetches in tests.
func SetInsecureSkipForTesting(skip bool) { insecureSkip = skip }

func policyActive() bool { return !insecureSkip }

var blockedHostnames = map[string]struct{}{
	"localhost":                {},
	"localhost.localdomain":    {},
	"metadata":                 {},
	"metadata.google.internal": {},
}

// BlockedIP reports whether ip is loopback, private, link local, or otherwise non-public.
func BlockedIP(ip net.IP) bool {
	if !policyActive() {
		return false
	}
	if ip == nil && len(ip) == 0 {
		return true
	}
	ip = ip.To16()
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		// 169.254.0.0/16 is link-local; IsLinkLocalUnicast covers it on supported Go versions.
		return false
	}
	// Unique local addresses fc00::/7
	return ip[0]&0xfe == 0xfc
}

func blockedHostname(host string) bool {
	if !policyActive() {
		return false
	}
	host = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(host)), "[")
	host = strings.TrimSuffix(host, "]")
	if _, ok := blockedHostnames[host]; ok {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return BlockedIP(ip)
	}
	return false
}

// CheckHost resolves host and rejects if any address is blocked or the name is disallowed.
func CheckHost(ctx context.Context, host string) error {
	if host == "" {
		return ErrBlocked
	}
	if blockedHostname(host) {
		return ErrBlocked
	}
	if ip := net.ParseIP(host); ip != nil {
		if BlockedIP(ip) {
			return ErrBlocked
		}
		return nil
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return err
	}
	if len(addrs) == 0 {
		return ErrBlocked
	}
	for _, a := range addrs {
		if BlockedIP(a.IP) {
			return ErrBlocked
		}
	}
	return nil
}

// CheckURL validates scheme and host before a page fetch or redirect hop.
func CheckURL(ctx context.Context, u *url.URL) error {
	if u == nil {
		return ErrBlocked
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrBlocked
	}
	return CheckHost(ctx, u.Hostname())
}

// CheckRedirect is suitable for http.Client.CheckRedirect.
func CheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("too many redirects")
	}
	return CheckURL(req.Context(), req.URL)
}
