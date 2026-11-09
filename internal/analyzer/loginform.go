package analyzer

import (
	"strings"

	"golang.org/x/net/html"
)

type formState struct {
	depth       int
	hint        bool
	hasPassword bool
	hasTextLike bool
}

func (f *formState) start(attrs []html.Attribute) {
	if f.depth > 0 {
		// Nested <form> start tags are ignored by the HTML5 parser.
		return
	}
	f.hint = loginHintInAttrs(attrs)
	f.hasPassword = false
	f.hasTextLike = false
	f.depth++
}

func (f *formState) input(attrs []html.Attribute) {
	if f.depth != 1 {
		return
	}
	typ := strings.ToLower(strings.TrimSpace(attr(attrs, "type")))
	switch typ {
	case "password":
		f.hasPassword = true
	case "", "text", "email":
		f.hasTextLike = true
	}
	ac := strings.ToLower(attr(attrs, "autocomplete"))
	if ac == "username" || ac == "current-password" {
		f.hasPassword = f.hasPassword || ac == "current-password"
		f.hasTextLike = f.hasTextLike || ac == "username"
	}
}

func (f *formState) end() bool {
	if f.depth == 0 {
		return false
	}
	login := false
	if f.depth == 1 {
		login = f.hasPassword || (f.hint && f.hasTextLike)
	}
	f.depth--
	if f.depth == 0 {
		f.hint = false
		f.hasPassword = false
		f.hasTextLike = false
	}
	return login
}

func (f *formState) flush() bool {
	if f.depth == 0 {
		return false
	}
	login := f.hasPassword || (f.hint && f.hasTextLike)
	f.depth = 0
	f.hint = false
	f.hasPassword = false
	f.hasTextLike = false
	return login
}

func loginHintInAttrs(attrs []html.Attribute) bool {
	for _, a := range attrs {
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
