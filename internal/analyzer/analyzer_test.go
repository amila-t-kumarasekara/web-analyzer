package analyzer_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/amila-t-kumarasekara/web-analyzer/internal/analyzer"
)

func base(t *testing.T) *url.URL {
	t.Helper()
	u, err := url.Parse("https://example.com/dir/")
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestAnalyze_basicPage(t *testing.T) {
	t.Parallel()

	html := `<!DOCTYPE html>
<html><head><title>My Site</title></head>
<body>
<h1>One</h1>
<h1>Two</h1>
<h2>Sub</h2>
</body></html>`

	res, err := analyzer.Analyze(strings.NewReader(html), base(t))
	if err != nil {
		t.Fatal(err)
	}
	if res.HTMLVersion != "HTML5" {
		t.Fatalf("version = %q, want HTML5", res.HTMLVersion)
	}
	if res.Title != "My Site" {
		t.Fatalf("title = %q", res.Title)
	}
	if res.Headings[0] != 2 || res.Headings[1] != 1 {
		t.Fatalf("headings = %v", res.Headings)
	}
}

func TestAnalyze_links(t *testing.T) {
	t.Parallel()

	html := `<html><body>
<a href="/local">in</a>
<a href="https://other.com/page">out</a>
<a href="#top">skip</a>
<a href="mailto:a@b.com">skip</a>
</body></html>`

	res, err := analyzer.Analyze(strings.NewReader(html), base(t))
	if err != nil {
		t.Fatal(err)
	}
	if res.InternalLinks != 1 || res.ExternalLinks != 1 {
		t.Fatalf("internal=%d external=%d", res.InternalLinks, res.ExternalLinks)
	}
	if len(res.CheckURLs) != 2 {
		t.Fatalf("check URLs = %v", res.CheckURLs)
	}
}

func TestAnalyze_loginForm(t *testing.T) {
	t.Parallel()

	html := `<html><body>
<form><input type="password" name="pw"></form>
</body></html>`

	res, err := analyzer.Analyze(strings.NewReader(html), base(t))
	if err != nil {
		t.Fatal(err)
	}
	if !res.HasLoginForm {
		t.Fatal("expected login form")
	}
}

func TestAnalyze_html401Doctype(t *testing.T) {
	t.Parallel()

	html := `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01//EN"
  "http://www.w3.org/TR/html4/strict.dtd">
<html><body></body></html>`

	res, err := analyzer.Analyze(strings.NewReader(html), base(t))
	if err != nil {
		t.Fatal(err)
	}
	if res.HTMLVersion != "HTML 4.01 Strict" {
		t.Fatalf("version = %q", res.HTMLVersion)
	}
}
