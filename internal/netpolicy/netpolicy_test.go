package netpolicy_test

import (
	"context"
	"net"
	"testing"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/netpolicy"
)

func TestBlockedIP(t *testing.T) {
	t.Parallel()

	cases := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true},
		{"::1", true},
		{"fe80::1", true},
		{"8.8.8.8", false},
		{"1.1.1.1", false},
	}
	for _, tc := range cases {
		ip := net.ParseIP(tc.ip)
		if ip == nil {
			t.Fatalf("parse %q", tc.ip)
		}
		if got := netpolicy.BlockedIP(ip); got != tc.want {
			t.Errorf("BlockedIP(%s) = %v, want %v", tc.ip, got, tc.want)
		}
	}
}

func TestCheckHost_blockedLiterals(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	for _, host := range []string{"127.0.0.1", "localhost", "169.254.169.254"} {
		if err := netpolicy.CheckHost(ctx, host); err != netpolicy.ErrBlocked {
			t.Fatalf("CheckHost(%q) = %v, want ErrBlocked", host, err)
		}
	}
}
