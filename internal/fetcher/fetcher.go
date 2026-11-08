// Package fetcher retrieves HTML pages over HTTP.
package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
)

// Page is a successfully fetched HTML document.
type Page struct {
	FinalURL *url.URL
	Status   int
	Body     []byte
}

// Error describes a failed fetch in user readable terms.
type Error struct {
	StatusCode int    // target's status, or 502/504 for network failures
	Reason     string // human readable description
	Err        error  // underlying error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Reason != "" {
		return e.Reason
	}
	if e.StatusCode != 0 {
		return fmt.Sprintf("fetch failed: %d %s", e.StatusCode, http.StatusText(e.StatusCode))
	}
	return "fetch failed"
}

func (e *Error) Unwrap() error { return e.Err }

// Client fetches pages.
type Client struct {
	http     *http.Client
	maxBytes int64
}

// New returns a Client using hc for requests.
func New(hc *http.Client, maxBytes int64) *Client {
	return &Client{http: hc, maxBytes: maxBytes}
}

// Fetch GETs u and returns the page or an *Error.
func (c *Client) Fetch(ctx context.Context, u *url.URL) (*Page, error) {
	if u == nil {
		return nil, &Error{StatusCode: http.StatusBadRequest, Reason: "missing URL"}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, &Error{StatusCode: http.StatusBadRequest, Reason: "invalid request", Err: err}
	}
	req.Header.Set("User-Agent", consts.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, mapFetchError(err)
	}
	defer resp.Body.Close()

	finalURL := u
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &Error{
			StatusCode: resp.StatusCode,
			Reason:     http.StatusText(resp.StatusCode),
			Err:        fmt.Errorf("HTTP %d", resp.StatusCode),
		}
	}

	if !isHTMLContentType(resp.Header.Get("Content-Type")) {
		return nil, &Error{
			StatusCode: http.StatusUnsupportedMediaType,
			Reason:     "response is not HTML",
		}
	}

	limit := c.maxBytes
	if limit <= 0 {
		limit = consts.DefaultMaxPageBytes
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, mapFetchError(err)
	}
	if int64(len(body)) > limit {
		return nil, &Error{
			StatusCode: http.StatusRequestEntityTooLarge,
			Reason:     "page exceeds size limit",
		}
	}

	return &Page{
		FinalURL: finalURL,
		Status:   resp.StatusCode,
		Body:     body,
	}, nil
}

func isHTMLContentType(raw string) bool {
	media, _, err := mime.ParseMediaType(raw)
	if err != nil || media == "" {
		return false
	}
	media = strings.ToLower(media)
	return media == "text/html" || media == "application/xhtml+xml"
}

func mapFetchError(err error) *Error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &Error{StatusCode: http.StatusGatewayTimeout, Reason: "request timed out", Err: err}
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return &Error{StatusCode: http.StatusGatewayTimeout, Reason: "request timed out", Err: err}
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		return &Error{StatusCode: http.StatusBadGateway, Reason: "could not resolve host", Err: err}
	}
	return &Error{StatusCode: http.StatusBadGateway, Reason: "could not reach server", Err: err}
}
