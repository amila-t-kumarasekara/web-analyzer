// Package linkcheck checks URL reachability concurrently.
package linkcheck

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
)

// Result is the outcome of checking every submitted URL (or stopping early).
type Result struct {
	Outcomes   []Outcome
	Total      int
	Incomplete bool // true when the context ended before every URL was checked
}

// Outcome is the result of checking one URL.
type Outcome struct {
	URL        string
	StatusCode int    // 0 if no response
	Reason     string // set when inaccessible
	OK         bool
}

// Checker checks links with a bounded worker pool.
type Checker struct {
	http    *http.Client
	workers int
}

// New returns a Checker.
func New(hc *http.Client, workers int) *Checker {
	return &Checker{http: hc, workers: workers}
}

// Check returns one Outcome per URL that was checked.
func (c *Checker) Check(ctx context.Context, urls []string) Result {
	if len(urls) == 0 {
		return Result{}
	}

	workers := c.workers
	if workers < 1 {
		workers = 1
	}
	if workers > len(urls) {
		workers = len(urls)
	}

	jobs := make(chan string)
	results := make(chan Outcome, len(urls))

	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for u := range jobs {
				results <- c.checkOne(ctx, u)
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, u := range urls {
			select {
			case <-ctx.Done():
				return
			case jobs <- u:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]Outcome, 0, len(urls))
	for o := range results {
		out = append(out, o)
	}
	return Result{
		Outcomes:   out,
		Total:      len(urls),
		Incomplete: len(out) < len(urls),
	}
}

// checkOne performs HEAD with GET fallback.
func (c *Checker) checkOne(ctx context.Context, u string) Outcome {
	if o := c.request(ctx, http.MethodHead, u); o.Reason == "" && !headFallbackToGet(o.StatusCode) {
		return o
	}
	return c.request(ctx, http.MethodGet, u)
}

func headFallbackToGet(status int) bool {
	return status == http.StatusMethodNotAllowed || status == http.StatusNotImplemented
}

func (c *Checker) request(ctx context.Context, method, u string) Outcome {
	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return Outcome{URL: u, Reason: friendlyCheckError(err)}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return Outcome{URL: u, Reason: friendlyCheckError(err)}
	}
	defer resp.Body.Close()
	if method == http.MethodGet {
		_, _ = io.Copy(io.Discard, resp.Body)
	}

	code := resp.StatusCode
	o := Outcome{
		URL:        u,
		StatusCode: code,
		OK:         code >= 200 && code < 400,
	}
	if !o.OK {
		o.Reason = http.StatusText(code)
	}
	return o
}

func friendlyCheckError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "Request timed out"
	}
	if errors.Is(err, context.Canceled) {
		return "Check cancelled"
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return "Request timed out"
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		return "Could not resolve host"
	}
	return "Could not reach server"
}
