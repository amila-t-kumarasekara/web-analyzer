package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/fetcher"
	"github.com/amila-t-kumarasekara/web-analyzer/internal/linkcheck"
)

type stubFetch struct {
	page *fetcher.Page
	err  error
}

func (s stubFetch) Fetch(context.Context, *url.URL) (*fetcher.Page, error) {
	return s.page, s.err
}

type stubCheck struct {
	res linkcheck.Result
}

func (s stubCheck) Check(context.Context, []string) linkcheck.Result {
	return s.res
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	h := NewHandler(stubFetch{}, stubCheck{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rr.Code != http.StatusOK || rr.Body.String() != "ok" {
		t.Fatalf("got %d body %q", rr.Code, rr.Body.String())
	}
}

func TestIndex(t *testing.T) {
	t.Parallel()

	h := NewHandler(stubFetch{}, stubCheck{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Web Page Analyzer") {
		t.Fatal("expected index HTML")
	}
}

func TestAnalyze_invalidURL(t *testing.T) {
	t.Parallel()

	h := NewHandler(stubFetch{}, stubCheck{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader("url=not+a+url"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "valid http or https URL") {
		t.Fatalf("body = %q", body)
	}
}

func TestAnalyze_success(t *testing.T) {
	t.Parallel()

	final, _ := url.Parse("https://example.com/")
	html := []byte(`<!DOCTYPE html><html><head><title>Hello</title></head><body><h1>x</h1></body></html>`)
	sf := stubFetch{page: &fetcher.Page{FinalURL: final, Body: html, Status: http.StatusOK}}

	h := NewHandler(sf, stubCheck{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader("url=https://example.com"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Hello") {
		t.Fatalf("missing title in %q", body)
	}
}

func TestAnalyze_fetchError(t *testing.T) {
	t.Parallel()

	sf := stubFetch{err: &fetcher.Error{StatusCode: http.StatusNotFound, Reason: "Not Found"}}
	h := NewHandler(sf, stubCheck{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/analyze", strings.NewReader("url=https://example.com/missing"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "Not Found") {
		t.Fatal("expected fetch error in page")
	}
}

func TestHeadingRows(t *testing.T) {
	t.Parallel()

	rows := headingRows([6]int{2, 0, 1, 0, 0, 0})
	if len(rows) != 2 {
		t.Fatalf("len = %d", len(rows))
	}
	if rows[0].Level != "h1" || rows[0].Count != 2 {
		t.Fatalf("first row = %+v", rows[0])
	}
	if rows[1].Level != "h3" || rows[1].Count != 1 {
		t.Fatalf("second row = %+v", rows[1])
	}
}

func TestPageFromFetchErr(t *testing.T) {
	t.Parallel()

	h := &Handler{}
	view := h.pageFromFetchErr("https://x.com", &fetcher.Error{StatusCode: 418, Reason: "teapot"})
	if view.Err == nil || view.Err.StatusCode != 418 || view.Err.Description != "teapot" {
		t.Fatalf("view = %+v", view.Err)
	}
}
