package fetcher_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/fetcher"
)

func TestClient_Fetch_ok(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	page, err := fetcher.New(srv.Client(), 1<<20).Fetch(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
	if page.Status != http.StatusOK {
		t.Fatalf("status = %d", page.Status)
	}
	if !strings.Contains(string(page.Body), "ok") {
		t.Fatalf("body = %q", page.Body)
	}
}

func TestClient_Fetch_rejectsNonHTML(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	u, _ := url.Parse(srv.URL)
	_, err := fetcher.New(srv.Client(), 1<<20).Fetch(context.Background(), u)
	if err == nil {
		t.Fatal("expected error")
	}
	var fe *fetcher.Error
	if !errors.As(err, &fe) {
		t.Fatalf("err type %T", err)
	}
	if fe.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("code = %d", fe.StatusCode)
	}
}

func TestClient_Fetch_httpError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	u, _ := url.Parse(srv.URL)
	_, err := fetcher.New(srv.Client(), 1<<20).Fetch(context.Background(), u)
	if err == nil {
		t.Fatal("expected error")
	}
	var fe *fetcher.Error
	if !errors.As(err, &fe) || fe.StatusCode != http.StatusNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestClient_Fetch_nilURL(t *testing.T) {
	t.Parallel()

	_, err := fetcher.New(http.DefaultClient, 1024).Fetch(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_Fetch_tooLarge(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(make([]byte, 64))
	}))
	t.Cleanup(srv.Close)

	u, _ := url.Parse(srv.URL)
	_, err := fetcher.New(srv.Client(), 16).Fetch(context.Background(), u)
	if err == nil {
		t.Fatal("expected error")
	}
	var fe *fetcher.Error
	if !errors.As(err, &fe) || fe.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("err = %v", err)
	}
}

func TestError_string(t *testing.T) {
	t.Parallel()

	if (&fetcher.Error{Reason: "bad"}).Error() != "bad" {
		t.Fatal("reason preferred")
	}
}
