// Package validate normalises and validates user-supplied URLs.
package validate

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
	"github.com/amila-t-kumarasekara/web-analyzer/pkg/errs"
)

// ErrInvalidURL is returned when input cannot be used as an http(s) URL.
var ErrInvalidURL = errs.ErrInvalidURL

// URL trims, normalises and validates raw input.
func URL(raw string) (*url.URL, error) {
	s := strings.TrimSpace(raw)
	if s == "" || len(s) > consts.MaxURLLen {
		return nil, ErrInvalidURL
	}

	switch {
	case strings.HasPrefix(s, "//"):
		s = "https:" + s
	case !strings.Contains(s, "://"):
		s = "https://" + s
	}

	if !consts.URLPrecheck.MatchString(s) {
		return nil, ErrInvalidURL
	}

	if i := strings.IndexByte(s, '#'); i >= 0 {
		s = s[:i]
	}

	u, err := url.ParseRequestURI(s)
	if err != nil {
		return nil, ErrInvalidURL
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, ErrInvalidURL
	}
	if u.Hostname() == "" {
		return nil, ErrInvalidURL
	}

	return u, nil
}

// Description returns a short, user-facing message for a failed URL validation.
func Description(raw string, err error) string {
	if err == nil {
		return ""
	}
	s := strings.TrimSpace(raw)
	if s == "" {
		return "Please enter a URL."
	}
	if len(s) > consts.MaxURLLen {
		return "That URL is too long (maximum " + strconv.Itoa(consts.MaxURLLen) + " characters)."
	}
	if !strings.Contains(s, "://") && !strings.HasPrefix(s, "//") {
		s = "https://" + s
	}
	if u, parseErr := url.Parse(s); parseErr == nil {
		scheme := strings.ToLower(u.Scheme)
		if scheme != "" && scheme != "http" && scheme != "https" {
			return "Only http and https URLs are supported."
		}
		if u.Hostname() == "" && !strings.Contains(s, "://") {
			return "Please include a hostname (for example, https://example.com)."
		}
	}
	return "Please enter a valid http or https URL (for example, https://example.com)."
}
