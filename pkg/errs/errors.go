// Package errs defines sentinel errors shared across the application.
package errs

import "errors"

// ErrInvalidURL is returned when input cannot be used as an http(s) URL.
var ErrInvalidURL = errors.New("invalid url")
