// Package analyzer extracts facts from an HTML document.
package analyzer

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Result holds everything the UI displays about a page.
type Result struct {
	HTMLVersion   string
	Title         string
	Headings      [6]int // index 0 = h1
	InternalLinks int
	ExternalLinks int
	CheckURLs     []string // unique absolute http(s) URLs to check
	HasLoginForm  bool
}

// Analyze parses r once and fills a Result. base resolves relative links.
func Analyze(r io.Reader, base *url.URL) (Result, error) {
	res := Result{HTMLVersion: htmlVersion("", "", false)}
	linkBase := base
	seen := make(map[string]struct{})

	var sawDoctype bool
	var inTitle bool
	var titleBuf strings.Builder
	var form formState

	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				break
			}
			return Result{}, z.Err()
		}
		t := z.Token()

		switch tt {
		case html.DoctypeToken:
			if !sawDoctype {
				sawDoctype = true
				pub, sys := doctypeIDs(t.Data)
				res.HTMLVersion = htmlVersion(pub, sys, true)
			}
		case html.TextToken:
			if inTitle && res.Title == "" {
				titleBuf.WriteString(t.Data)
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			switch t.DataAtom {
			case atom.Title:
				if res.Title == "" {
					inTitle = true
					titleBuf.Reset()
				}
			case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
				if level, ok := headingLevel(t.DataAtom); ok {
					res.Headings[level]++
				}
			case atom.Base:
				if linkBase != nil {
					if href := attr(t.Attr, "href"); href != "" {
						if ref, err := url.Parse(href); err == nil {
							linkBase = linkBase.ResolveReference(ref)
						}
					}
				}
			case atom.A:
				if href := attr(t.Attr, "href"); href != "" {
					abs, kind := classify(href, linkBase)
					switch kind {
					case linkInternal:
						res.InternalLinks++
						addCheckURL(&res, seen, abs)
					case linkExternal:
						res.ExternalLinks++
						addCheckURL(&res, seen, abs)
					}
				}
			case atom.Form:
				if tt == html.StartTagToken {
					form.start(t.Attr)
				}
			case atom.Input:
				if tt == html.StartTagToken || tt == html.SelfClosingTagToken {
					form.input(t.Attr)
				}
			}
		case html.EndTagToken:
			switch t.DataAtom {
			case atom.Title:
				if inTitle {
					res.Title = strings.TrimSpace(titleBuf.String())
					inTitle = false
				}
			case atom.Form:
				if form.end() {
					res.HasLoginForm = true
				}
			}
		}
	}

	if inTitle && res.Title == "" {
		res.Title = strings.TrimSpace(titleBuf.String())
	}
	if form.flush() {
		res.HasLoginForm = true
	}

	return res, nil
}

// doctypeIDs extracts public and system identifiers from a DoctypeToken's Data,
// following the same rules as golang.org/x/net/html's parseDoctype.
func doctypeIDs(s string) (public, system string) {
	const ws = " \t\n\r\f"

	space := strings.IndexAny(s, ws)
	if space == -1 {
		space = len(s)
	}
	s = strings.TrimLeft(s[space:], ws)

	if len(s) < 6 {
		return "", ""
	}

	key := strings.ToLower(s[:6])
	s = s[6:]
	for key == "public" || key == "system" {
		s = strings.TrimLeft(s, ws)
		if s == "" {
			break
		}
		quote := s[0]
		if quote != '"' && quote != '\'' {
			break
		}
		s = s[1:]
		q := strings.IndexRune(s, rune(quote))
		var id string
		if q == -1 {
			id = s
			s = ""
		} else {
			id = s[:q]
			s = s[q+1:]
		}
		if key == "public" {
			public = id
			key = "system"
		} else {
			system = id
			key = ""
		}
	}
	return public, system
}

func attr(attrs []html.Attribute, key string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func headingLevel(a atom.Atom) (int, bool) {
	switch a {
	case atom.H1:
		return 0, true
	case atom.H2:
		return 1, true
	case atom.H3:
		return 2, true
	case atom.H4:
		return 3, true
	case atom.H5:
		return 4, true
	case atom.H6:
		return 5, true
	default:
		return 0, false
	}
}

func addCheckURL(res *Result, seen map[string]struct{}, u *url.URL) {
	if u == nil {
		return
	}
	s := u.String()
	if _, ok := seen[s]; ok {
		return
	}
	seen[s] = struct{}{}
	res.CheckURLs = append(res.CheckURLs, s)
}
