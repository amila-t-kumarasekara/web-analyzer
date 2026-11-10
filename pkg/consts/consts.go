// Package consts holds shared limits, timeouts, and defaults for the app.
package consts

import (
	"regexp"
	"time"
)

const (
	UserAgent = "WebAnalyzer/1.0"

	MaxURLLen = 2048

	DefaultPort               = "8080"
	DefaultAdminAddr          = "127.0.0.1:6060"
	DefaultLinkWorkers        = 32
	DefaultMaxPageBytes int64 = 10 << 20 // 10 MiB

	AnalyzeTimeout   = 30 * time.Second
	LinkCheckTimeout = 20 * time.Second

	PageHTTPTimeout = 30 * time.Second
	LinkHTTPTimeout = 15 * time.Second

	ConnectTimeout        = 10 * time.Second
	TLSHandshakeTimeout   = 10 * time.Second
	ResponseHeaderTimeout = 15 * time.Second

	ShutdownTimeout = 15 * time.Second

	ReadHeaderTimeout = 5 * time.Second
	ReadTimeout       = 10 * time.Second
	WriteTimeout      = 60 * time.Second
	IdleTimeout       = 120 * time.Second
)

// URLPrecheck runs before url.ParseRequestURI (http(s), non-whitespace).
var URLPrecheck = regexp.MustCompile(`(?i)^https?://[^\s]+$`)
