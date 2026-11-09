package httpclient_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/httpclient"
	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
)

func TestNewPageClient_fetches(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	resp, err := httpclient.NewPageClient().Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestDialIPv4First_prefersIPv4(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Skip("ipv4 listen:", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		_ = conn.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), consts.ConnectTimeout)
	defer cancel()

	tr := httpclient.NewPageClient().Transport.(*http.Transport)
	conn, err := tr.DialContext(ctx, "tcp", net.JoinHostPort("localhost", portOf(t, ln)))
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("listener never accepted")
	}
}

func portOf(t *testing.T, ln net.Listener) string {
	t.Helper()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	return port
}
