package analyzer

import "testing"

func TestHTMLVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pub, sys string
		has      bool
		want     string
	}{
		{"", "", false, "Unknown"},
		{"", "", true, "HTML5"},
		{"-//W3C//DTD HTML 4.01//EN", "", true, "HTML 4.01 Strict"},
		{"", "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd", true, "XHTML 1.0 Strict"},
		{"unknown", "unknown", true, "Unknown"},
	}

	for _, tt := range tests {
		got := htmlVersion(tt.pub, tt.sys, tt.has)
		if got != tt.want {
			t.Errorf("htmlVersion(%q,%q,%v) = %q, want %q", tt.pub, tt.sys, tt.has, got, tt.want)
		}
	}
}
