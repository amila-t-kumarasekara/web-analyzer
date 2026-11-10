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

func TestDoctypeIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		data       string
		wantPublic string
		wantSystem string
	}{
		{"html", "", ""},
		{
			`html PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd"`,
			"-//W3C//DTD HTML 4.01//EN",
			"http://www.w3.org/TR/html4/strict.dtd",
		},
		{
			`html SYSTEM "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"`,
			"",
			"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.data, func(t *testing.T) {
			t.Parallel()
			pub, sys := doctypeIDs(tt.data)
			if pub != tt.wantPublic || sys != tt.wantSystem {
				t.Fatalf("doctypeIDs(%q) = (%q, %q), want (%q, %q)",
					tt.data, pub, sys, tt.wantPublic, tt.wantSystem)
			}
		})
	}
}
