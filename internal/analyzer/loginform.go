package analyzer

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// looksLikeLoginForm reports whether form appears to be for signing in.
func looksLikeLoginForm(form *html.Node) bool {
	if form == nil || form.DataAtom != atom.Form {
		return false
	}

	loginHint := loginHintInForm(form)
	var hasPassword, hasTextLike bool

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type != html.ElementNode {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
			return
		}

		switch n.DataAtom {
		case atom.Input:
			typ := strings.ToLower(strings.TrimSpace(attr(n, "type")))
			switch typ {
			case "password":
				hasPassword = true
			case "", "text", "email":
				hasTextLike = true
			}
			ac := strings.ToLower(attr(n, "autocomplete"))
			if ac == "username" || ac == "current-password" {
				hasPassword = hasPassword || ac == "current-password"
				hasTextLike = hasTextLike || ac == "username"
			}
		case atom.Form:
			// Ignore nested forms.
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(form)
	if hasPassword {
		return true
	}
	return loginHint && hasTextLike
}

func loginHintInForm(form *html.Node) bool {
	for _, a := range form.Attr {
		val := strings.ToLower(a.Val)
		switch strings.ToLower(a.Key) {
		case "action", "id", "name", "class":
			if containsLoginWord(val) {
				return true
			}
		}
	}
	return false
}

func containsLoginWord(s string) bool {
	for _, w := range []string{"login", "log-in", "signin", "sign-in", "signon", "sign-on", "auth"} {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}
