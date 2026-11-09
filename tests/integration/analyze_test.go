//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/fetcher"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/httpclient"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/linkcheck"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/web"
	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
)

func newAppHandler(t *testing.T) http.Handler {
	t.Helper()

	fetchClient := fetcher.New(httpclient.NewPageClient(), consts.DefaultMaxPageBytes)
	linkChecker := linkcheck.New(httpclient.NewLinkClient(), consts.DefaultLinkWorkers)
	return web.NewHandler(fetchClient, linkChecker)
}

func TestAnalyze_fullStack_success(t *testing.T) {
	t.Parallel()

	extBroken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(extBroken.Close)

	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html><head><title>Integration Page</title></head>
<body>
<h1>a</h1><h2>b</h2>
<a href="/good">int ok</a>
<a href="/bad">int bad</a>
<a href="%s">ext bad</a>
<form><input type="password" name="p"></form>
</body></html>`, extBroken.URL)
		case "/good":
			w.WriteHeader(http.StatusOK)
		case "/bad":
			w.WriteHeader(http.StatusNotFound)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(page.Close)

	h := newAppHandler(t)
	rr := httptest.NewRecorder()
	form := url.Values{"url": {page.URL}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}

	body := rr.Body.String()
	for _, want := range []string{
		"Integration Page",
		"HTML5",
		"Login form",
		">Yes<",
		">Internal</th>",
		">External</th>",
		">Inaccessible</th>",
		"/bad",
		extBroken.URL,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response missing %q", want)
		}
	}
}

func TestAnalyze_fullStack_unreachablePage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	h := newAppHandler(t)
	rr := httptest.NewRecorder()
	form := url.Values{"url": {srv.URL}}.Encode()
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "404") || !strings.Contains(body, "Not Found") {
		t.Fatalf("expected fetch error UI, got %q", body)
	}
}
