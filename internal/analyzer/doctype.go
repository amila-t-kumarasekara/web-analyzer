package analyzer

import (
	"strings"

	"github.com/amila-t-kumarasekara/web-analyzer/pkg/doctype"
)

// htmlVersion maps DOCTYPE identifiers to a version name.
func htmlVersion(public, system string, hasDoctype bool) string {
	if !hasDoctype {
		return "Unknown"
	}

	pub := strings.TrimSpace(public)
	sys := strings.TrimSpace(system)
	if pub == "" && sys == "" {
		return "HTML5"
	}

	if pub != "" {
		if v, ok := doctype.ByPublic[strings.ToLower(pub)]; ok {
			return v
		}
	}
	if sys != "" {
		if v, ok := doctype.BySystem[strings.ToLower(sys)]; ok {
			return v
		}
	}
	return "Unknown"
}
