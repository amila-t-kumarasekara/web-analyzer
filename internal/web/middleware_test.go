package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRequestID(t *testing.T) {
	t.Parallel()

	var ctxID string
	h := withRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID, _ = r.Context().Value(requestIDKey{}).(string)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "test-id")
	h.ServeHTTP(rr, req)

	if rr.Header().Get("X-Request-ID") != "test-id" || ctxID != "test-id" {
		t.Fatalf("header=%q ctx=%q", rr.Header().Get("X-Request-ID"), ctxID)
	}
}

func TestWithRecover(t *testing.T) {
	t.Parallel()

	h := withRecover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rr.Code)
	}
}
