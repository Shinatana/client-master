package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

// Response contains the HTTP status code, response body, and response headers
// returned by a completed HTTP request.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

const (
	// defaultTimeout deprecated
	defaultTimeout = 10
	// DefaultTimeout is the package-level default timeout for new clients when no
	// explicit timeout is provided via options.
	DefaultTimeout = 10 * time.Second
)

// Client wraps http.Client with a parsed base URL, default headers, and an optional logger.
// It is safe for concurrent use for sending requests, but mutating methods that change
// default headers (e.g., AddHeader, AddHeaders, ReplaceHeaders) are not safe for
// concurrent use and should be done during setup.
type Client struct {
	// Deprecated: use headers instead.
	Headers Headers
	// Deprecated: use base instead.
	baseUrl string

	base       *url.URL
	headers    http.Header
	httpClient http.Client
	lg         *zerolog.Logger
}

// NewHTTPClient constructs a Client from a baseURL string and optional configuration
// options. The baseURL must be a valid URL. Options control default headers,
// timeout, and logging. When no timeout or headers are provided, package defaults
// are applied.
func NewHTTPClient(baseURL string, opts ...Option) (*Client, error) {
	baseParsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url '%s': %w", baseURL, err)
	}

	suppliedOptions := applyOptions(opts...)

	return &Client{
		base:    baseParsedURL,
		headers: suppliedOptions.headers,
		httpClient: http.Client{
			Timeout: suppliedOptions.timeout,
		},
		lg: suppliedOptions.lg,
	}, nil
}

// AddHeader appends one or more values for the given header key to the client's
// default headers. Multiple values for the same key are preserved (Add semantics).
// Not safe for concurrent use with in-flight mutations.
func (c *Client) AddHeader(key string, vals []string) {
	if c.headers == nil {
		c.headers = make(http.Header, 1)
	}
	for _, v := range vals {
		c.headers.Add(key, v)
	}
}

// AddHeaders appends all provided headers to the client's default headers.
// Values are appended (Add semantics) rather than replaced. Passing a nil map
// is a no-op. Not safe for concurrent use with in-flight mutations.
func (c *Client) AddHeaders(h http.Header) {
	if h == nil {
		return
	}
	if c.headers == nil {
		c.headers = make(http.Header, len(h))
	}
	for k, vals := range h {
		for _, v := range vals {
			c.headers.Add(k, v)
		}
	}
}

// ReplaceHeaders replaces the client's default headers with a clone of the
// provided map. If h is nil, the defaults are reset to an empty map. Not safe
// for concurrent use with in-flight mutations.
func (c *Client) ReplaceHeaders(h http.Header) {
	if h == nil {
		return
	}

	c.headers = h.Clone()
}
