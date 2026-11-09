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
