package analyzer

import (
	"net"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

type linkKind int

const (
	linkSkip linkKind = iota // fragment, mailto, javascript, ...
	linkInternal
	linkExternal
)

// classify resolves href against base and reports its kind.
func classify(href string, base *url.URL) (*url.URL, linkKind) {
	href = strings.TrimSpace(href)
	if href == "" || base == nil {
		return nil, linkSkip
	}
	if strings.HasPrefix(href, "#") {
		return nil, linkSkip
	}

	lower := strings.ToLower(href)
	switch {
	case strings.HasPrefix(lower, "mailto:"),
		strings.HasPrefix(lower, "javascript:"),
		strings.HasPrefix(lower, "tel:"),
		strings.HasPrefix(lower, "data:"):
		return nil, linkSkip
	}

	ref, err := url.Parse(href)
	if err != nil {
		return nil, linkSkip
	}
	if ref.Scheme == "" && ref.Host == "" && ref.Path == "" && ref.Opaque == "" && ref.RawQuery == "" && ref.Fragment != "" {
		return nil, linkSkip
	}

	abs := base.ResolveReference(ref)
	scheme := strings.ToLower(abs.Scheme)
	if scheme != "" && scheme != "http" && scheme != "https" {
		return nil, linkSkip
	}
	if abs.Host == "" {
		return nil, linkSkip
	}

	if sameHost(abs, base) {
		return abs, linkInternal
	}
	return abs, linkExternal
}

// sameHost reports whether a and b refer to the same site. Subdomains of the
// same registrable domain (m.example.com, www.example.com, example.com) count
// as one site. Hosts without a registrable domain (IPs, localhost) must match
// exactly, including port.
func sameHost(a, b *url.URL) bool {
	if a == nil || b == nil {
		return false
	}
	da, okA := registrableDomain(a)
	db, okB := registrableDomain(b)
	if okA && okB {
		return da == db
	}
	return siteKey(a) == siteKey(b)
}

// registrableDomain returns the eTLD+1 of u's host (e.g. "example.com" for
// "developers.example.com"). ok is false for IPs, localhost and other hosts
// that have no registrable domain.
func registrableDomain(u *url.URL) (string, bool) {
	host := strings.ToLower(u.Hostname())
	if host == "" || net.ParseIP(host) != nil {
		return "", false
	}
	d, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", false
	}
	return d, true
}

func siteKey(u *url.URL) string {
	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")

	port := u.Port()
	if port == "" {
		switch strings.ToLower(u.Scheme) {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}
	if port == "80" || port == "443" {
		return host
	}
	return host + ":" + port
}
