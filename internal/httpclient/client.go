// Package httpclient builds outbound HTTP clients for page fetch and link checks.
package httpclient

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/netpolicy"
	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
)

// NewPageClient returns a client for fetching HTML pages.
func NewPageClient() *http.Client {
	return &http.Client{
		Timeout:       consts.PageHTTPTimeout,
		Transport:     newTransport(2),
		CheckRedirect: netpolicy.CheckRedirect,
	}
}

// NewLinkClient returns a client tuned for concurrent link checks.
func NewLinkClient() *http.Client {
	return &http.Client{
		Timeout:       consts.LinkHTTPTimeout,
		Transport:     newTransport(consts.DefaultLinkWorkers),
		CheckRedirect: netpolicy.CheckRedirect,
	}
}

func newTransport(idlePerHost int) *http.Transport {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DialContext = dialIPv4First
	tr.TLSHandshakeTimeout = consts.TLSHandshakeTimeout
	tr.ResponseHeaderTimeout = consts.ResponseHeaderTimeout
	tr.ExpectContinueTimeout = time.Second
	if idlePerHost > 0 {
		tr.MaxIdleConns = idlePerHost * 4
		tr.MaxIdleConnsPerHost = idlePerHost
	}
	return tr
}

// dialIPv4First dials hostnames by trying IPv4 addresses before IPv6, so a
// black-holed IPv6 path does not burn the full connect timeout before IPv4.
func dialIPv4First(ctx context.Context, network, address string) (net.Conn, error) {
	d := &net.Dialer{
		Timeout:       consts.ConnectTimeout,
		KeepAlive:     30 * time.Second,
		FallbackDelay: 300 * time.Millisecond,
	}

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if ip := net.ParseIP(host); ip != nil {
		if netpolicy.BlockedIP(ip) {
			return nil, netpolicy.ErrBlocked
		}
		return d.DialContext(ctx, network, address)
	}

	if err := netpolicy.CheckHost(ctx, host); err != nil {
		return nil, err
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	var ipv4, ipv6 []net.IPAddr
	for _, a := range addrs {
		if netpolicy.BlockedIP(a.IP) {
			continue
		}
		if a.IP.To4() != nil {
			ipv4 = append(ipv4, a)
		} else if a.IP.To16() != nil {
			ipv6 = append(ipv6, a)
		}
	}
	ordered := append(ipv4, ipv6...)
	if len(ordered) == 0 {
		return nil, netpolicy.ErrBlocked
	}

	var lastErr error
	for _, a := range ordered {
		addr := net.JoinHostPort(a.IP.String(), port)
		conn, err := d.DialContext(ctx, network, addr)
		if err == nil {
			return conn, nil
		}
		lastErr = err
		if errors.Is(err, netpolicy.ErrBlocked) {
			return nil, err
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, netpolicy.ErrBlocked
}
