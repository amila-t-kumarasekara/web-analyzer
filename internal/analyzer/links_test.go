package analyzer

import (
	"net/url"
	"testing"
)

func TestClassify(t *testing.T) {
	t.Parallel()

	base, _ := url.Parse("https://example.com/app/")

	tests := []struct {
		href string
		want linkKind
	}{
		{"/path", linkInternal},
		{"https://example.com/x", linkInternal},
		{"https://other.com/x", linkExternal},
		{"https://developers.example.com/x", linkInternal},
		{"https://example.org/x", linkExternal},
		{"https://notexample.com/x", linkExternal},
		{"#", linkInternal},
		{"#frag", linkSkip},
		{"mailto:x@y.com", linkSkip},
		{"", linkSkip},
	}

	for _, tt := range tests {
		_, kind := classify(tt.href, base)
		if kind != tt.want {
			t.Errorf("classify(%q) = %v, want %v", tt.href, kind, tt.want)
		}
	}
}

func TestSameHost_wwwIgnored(t *testing.T) {
	t.Parallel()

	a, _ := url.Parse("https://www.example.com/")
	b, _ := url.Parse("https://example.com/")
	if !sameHost(a, b) {
		t.Fatal("www should match bare host")
	}
}

func TestClassify_unwrapsRedirectShim(t *testing.T) {
	t.Parallel()

	base, _ := url.Parse("https://www.example.com/login/")

	abs, kind := classify("https://l.example.com/l.php?u=https%3A%2F%2Fother.test%2F&h=track", base)
	if kind != linkExternal {
		t.Fatalf("kind = %v, want external", kind)
	}
	if got := abs.String(); got != "https://other.test/" {
		t.Fatalf("abs = %q, want https://other.test/", got)
	}

	// Shim pointing back at the same site stays internal.
	_, kind = classify("https://lm.example.com/l.php?u=https%3A%2F%2Fwww.example.com%2Fhelp", base)
	if kind != linkInternal {
		t.Fatalf("kind = %v, want internal", kind)
	}

	// A shim without a usable target is left as-is.
	abs, _ = classify("https://l.example.com/l.php?u=javascript%3Aalert(1)", base)
	if abs.Host != "l.example.com" {
		t.Fatalf("unexpected unwrap: %q", abs)
	}

	// Host prefix other than l/lm: target is not unwrapped.
	abs, _ = classify("https://track.example.com/l.php?u=https%3A%2F%2Fother.test%2F", base)
	if abs.Host != "track.example.com" {
		t.Fatalf("should not unwrap: %q", abs)
	}
}
