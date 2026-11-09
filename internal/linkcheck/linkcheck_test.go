package linkcheck_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/linkcheck"
	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
)

func TestCheck_sendsBrowserLikeHeaders(t *testing.T) {
	t.Parallel()

	var gotUA, gotAccept, gotLang string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept")
		gotLang = r.Header.Get("Accept-Language")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := linkcheck.New(srv.Client(), 1)
	res := c.Check(context.Background(), []string{srv.URL})

	if len(res.Outcomes) != 1 || !res.Outcomes[0].OK {
		t.Fatalf("outcome = %+v", res.Outcomes)
	}
	if gotUA != consts.UserAgent {
		t.Fatalf("User-Agent = %q, want %q", gotUA, consts.UserAgent)
	}
	if gotAccept != "text/html,*/*" {
		t.Fatalf("Accept = %q", gotAccept)
	}
	if gotLang != "en-US,en;q=0.9" {
		t.Fatalf("Accept-Language = %q", gotLang)
	}
}

func TestCheck_allURLs_checked(t *testing.T) {
	t.Parallel()

	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	urls := make([]string, 20)
	for i := range urls {
		urls[i] = srv.URL
	}

	c := linkcheck.New(srv.Client(), 4)
	res := c.Check(context.Background(), urls)

	if res.Total != len(urls) {
		t.Fatalf("total = %d, want %d", res.Total, len(urls))
	}
	if len(res.Outcomes) != len(urls) {
		t.Fatalf("outcomes = %d, want %d", len(res.Outcomes), len(urls))
	}
	if res.Incomplete {
		t.Fatal("expected complete result")
	}
}

func TestCheck_contextCancel_noHang(t *testing.T) {
	t.Parallel()

	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-block:
			return
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(func() {
		close(block)
		srv.Close()
	})

	urls := make([]string, 50)
	for i := range urls {
		urls[i] = srv.URL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	c := linkcheck.New(srv.Client(), 8)
	done := make(chan struct{})
	var res linkcheck.Result
	go func() {
		res = c.Check(ctx, urls)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Check did not return after context timeout (possible goroutine leak or deadlock)")
	}

	if !res.Incomplete {
		t.Fatalf("expected incomplete, got %d outcomes for %d urls", len(res.Outcomes), res.Total)
	}
	if len(res.Outcomes) >= len(urls) {
		t.Fatalf("expected fewer than %d outcomes after cancel, got %d", len(urls), len(res.Outcomes))
	}
}
