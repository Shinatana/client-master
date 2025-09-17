package client

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// Option represents a functional option that can modify the internal
// configuration used when constructing a Client.
// Typical usage:
//
//	NewHTTPClient(baseURL, WithTimeout(5*time.Second), WithLogger(logger))
type Option func(*optionList)

// optionList collects configuration provided via Option functions.
// It is an internal container used during Client construction.
//   - lg: optional structured logger; if nil, it is normalized to a no-op logger.
//   - timeout: HTTP client timeout; if zero, it is normalized to a package default.
//   - headers: initial default headers added to every request; if nil, it is normalized to an empty map.
type optionList struct {
	lg      *zerolog.Logger
	timeout time.Duration
	headers http.Header
}

// WithLogger configures a zerolog.Logger to be used by the client.
// Passing nil is allowed; it will be normalized to a no-op logger.
func WithLogger(lg *zerolog.Logger) Option {
	return func(o *optionList) {
		o.lg = lg
	}
}

// WithTimeout sets the HTTP client timeout.
// If a zero duration is provided, it will be replaced with the DefaultTimeout.
func WithTimeout(timeout time.Duration) Option {
	return func(o *optionList) {
		o.timeout = timeout
	}
}

// WithHeaders sets default Headers to be sent with every request.
// If nil is provided, it will be normalized to an empty map.
// Callers may still override or add request-specific headers later.
func WithHeaders(headers http.Header) Option {
	return func(o *optionList) {
		o.headers = headers
	}
}

// applyOptions applies all provided Option functions, then normalizes
// unset or zero values to safe defaults (logger, timeout, headers).
// It returns a fully initialized optionList ready to construct a Client.
func applyOptions(opts ...Option) optionList {
	var o optionList
	for _, opt := range opts {
		opt(&o)
	}

	o.lg = normalizeLogger(o.lg)
	o.timeout = normalizeTimeout(o.timeout)
	o.headers = normalizeHeaders(o.headers)

	return o
}

// normalizeLogger returns a usable logger.
// If lg is nil, a no-op logger is returned to avoid nil-pointer usage.
func normalizeLogger(lg *zerolog.Logger) *zerolog.Logger {
	if lg == nil {
		nop := zerolog.Nop()
		return &nop
	}

	return lg
}

// normalizeTimeout returns a valid timeout.
// If timeout is zero, the DefaultTimeout is used.
func normalizeTimeout(timeout time.Duration) time.Duration {
	if timeout == 0 {
		return DefaultTimeout
	}

	return timeout
}

// normalizeHeaders returns a non-nil headers map.
// If headers is nil, an empty map is allocated.
func normalizeHeaders(headers http.Header) http.Header {
	if headers == nil {
		return make(http.Header)
	}

	return headers
}
