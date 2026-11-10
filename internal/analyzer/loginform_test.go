package analyzer

import (
	"net/url"
	"strings"
	"testing"
)

func TestAnalyze_loginForm_cases(t *testing.T) {
	t.Parallel()

	base, err := url.Parse("https://example.com/")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		html string
		want bool
	}{
		{
			name: "password field",
			html: `<form><input type="password"></form>`,
			want: true,
		},
		{
			name: "hint only with text field",
			html: `<form class="login-form"><input type="text" name="user"></form>`,
			want: true,
		},
		{
			name: "hint only missing text field",
			html: `<form id="signin"></form>`,
			want: false,
		},
		{
			name: "nested form password stays in outer form",
			html: `<form>
<input type="text">
<form><input type="password"></form>
</form>`,
			want: true,
		},
		{
			name: "unclosed form flushed at eof",
			html: `<form><input type="password" name="pw">`,
			want: true,
		},
		{
			name: "unclosed hint form flushed at eof",
			html: `<form action="/login"><input type="email" name="user">`,
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, err := Analyze(strings.NewReader(tt.html), base)
			if err != nil {
				t.Fatal(err)
			}
			if res.HasLoginForm != tt.want {
				t.Fatalf("HasLoginForm = %v, want %v", res.HasLoginForm, tt.want)
			}
		})
	}
}
