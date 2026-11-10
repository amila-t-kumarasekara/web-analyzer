// Package doctype maps HTML DOCTYPE public and system identifiers to version names.
package doctype

// ByPublic maps normalised public identifiers to HTML version labels.
var ByPublic = map[string]string{
	"-//w3c//dtd html 4.01//en":                "HTML 4.01 Strict",
	"-//w3c//dtd html 4.01 transitional//en":   "HTML 4.01 Transitional",
	"-//w3c//dtd html 4.01 frameset//en":       "HTML 4.01 Frameset",
	"-//w3c//dtd html 4.0//en":                 "HTML 4.0 Strict",
	"-//w3c//dtd html 4.0 transitional//en":    "HTML 4.0 Transitional",
	"-//w3c//dtd html 4.0 frameset//en":        "HTML 4.0 Frameset",
	"-//w3c//dtd html 3.2 final//en":           "HTML 3.2",
	"-//w3c//dtd html 3.2//en":                 "HTML 3.2",
	"-//w3c//dtd xhtml 1.0 strict//en":         "XHTML 1.0 Strict",
	"-//w3c//dtd xhtml 1.0 transitional//en":   "XHTML 1.0 Transitional",
	"-//w3c//dtd xhtml 1.0 frameset//en":       "XHTML 1.0 Frameset",
	"-//w3c//dtd xhtml 1.1//en":                "XHTML 1.1",
	"-//ietf//dtd html//en":                    "HTML 2.0",
	"-//ietf//dtd html 2.0//en":                "HTML 2.0",
	"-//ietf//dtd html 2.0 strict//en":         "HTML 2.0",
	"-//ietf//dtd html 2.0 level 1//en":        "HTML 2.0",
	"-//ietf//dtd html 2.0 level 2//en":        "HTML 2.0",
	"-//ietf//dtd html 2.0 strict level 1//en": "HTML 2.0",
	"-//ietf//dtd html 2.0 strict level 2//en": "HTML 2.0",
}

// BySystem maps normalised system identifiers to HTML version labels.
var BySystem = map[string]string{
	"http://www.w3.org/tr/html4/strict.dtd":                   "HTML 4.01 Strict",
	"http://www.w3.org/tr/html4/loose.dtd":                    "HTML 4.01 Transitional",
	"http://www.w3.org/tr/html4/frameset.dtd":                 "HTML 4.01 Frameset",
	"http://www.w3.org/tr/xhtml1/dtd/xhtml1-strict.dtd":       "XHTML 1.0 Strict",
	"http://www.w3.org/tr/xhtml1/dtd/xhtml1-transitional.dtd": "XHTML 1.0 Transitional",
	"http://www.w3.org/tr/xhtml1/dtd/xhtml1-frameset.dtd":     "XHTML 1.0 Frameset",
	"http://www.w3.org/tr/xhtml11/dtd/xhtml11.dtd":            "XHTML 1.1",
}
