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
	doc, err := html.Parse(r)
	if err != nil {
		return Result{}, err
	}

	res := Result{HTMLVersion: htmlVersion("", "", false)}
	linkBase := base
	seen := make(map[string]struct{})
	var sawDoctype bool

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.DoctypeNode:
			if !sawDoctype {
				sawDoctype = true
				pub, sys := doctypeAttrs(n)
				res.HTMLVersion = htmlVersion(pub, sys, true)
			}
		case html.ElementNode:
			switch n.DataAtom {
			case atom.Title:
				if res.Title == "" {
					res.Title = strings.TrimSpace(textContent(n))
				}
			case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
				if level, ok := headingLevel(n.DataAtom); ok {
					res.Headings[level]++
				}
			case atom.Base:
				if linkBase != nil {
					if href := attr(n, "href"); href != "" {
						if ref, err := url.Parse(href); err == nil {
							linkBase = linkBase.ResolveReference(ref)
						}
					}
				}
			case atom.A:
				if href := attr(n, "href"); href != "" {
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
				if looksLikeLoginForm(n) {
					res.HasLoginForm = true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return res, nil
}

func doctypeAttrs(n *html.Node) (public, system string) {
	for _, a := range n.Attr {
		switch strings.ToLower(a.Key) {
		case "public":
			public = a.Val
		case "system":
			system = a.Val
		}
	}
	return public, system
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(textContent(c))
	}
	return b.String()
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
