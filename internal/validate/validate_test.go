package validate_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/validate"
	"github.com/amila-t-kumarasekara/web-analyzer/pkg/consts"
	"github.com/amila-t-kumarasekara/web-analyzer/pkg/errs"
)

func TestURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw    string
		wantOK bool
		want   string
	}{
		{"https://example.com", true, "https://example.com"},
		{"  example.com/path  ", true, "https://example.com/path"},
		{"//cdn.example.com/x", true, "https://cdn.example.com/x"},
		{"", false, ""},
		{"ftp://files.example.com", false, ""},
		{"not a url", false, ""},
		{"https://", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()
			u, err := validate.URL(tt.raw)
			if tt.wantOK {
				if err != nil {
					t.Fatalf("URL(%q) err = %v", tt.raw, err)
				}
				if u.String() != tt.want {
					t.Fatalf("got %q, want %q", u.String(), tt.want)
				}
				return
			}
			if err == nil {
				t.Fatalf("URL(%q) expected error", tt.raw)
			}
			if !errors.Is(err, errs.ErrInvalidURL) {
				t.Fatalf("err = %v, want ErrInvalidURL", err)
			}
		})
	}
}

func TestURL_stripsFragment(t *testing.T) {
	t.Parallel()

	u, err := validate.URL("https://example.com/page#section")
	if err != nil {
		t.Fatal(err)
	}
	if u.Fragment != "" {
		t.Fatalf("fragment = %q, want empty", u.Fragment)
	}
	if u.String() != "https://example.com/page" {
		t.Fatalf("got %q", u.String())
	}
}

func TestDescription(t *testing.T) {
	t.Parallel()

	if validate.Description("https://x.com", nil) != "" {
		t.Fatal("expected empty description when err is nil")
	}
	if !strings.Contains(validate.Description("", errs.ErrInvalidURL), "enter a URL") {
		t.Fatal("empty input message")
	}
	long := strings.Repeat("a", consts.MaxURLLen+1)
	if !strings.Contains(validate.Description(long, errs.ErrInvalidURL), "too long") {
		t.Fatal("long input message")
	}
	if validate.Description("ftp://x.com", errs.ErrInvalidURL) != "Only http and https URLs are supported." {
		t.Fatal("ftp scheme message")
	}
}
