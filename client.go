package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

const (
	defaultTimeout = 10
	DefaultTimeout = 10 * time.Second
)

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

func (c *Client) AddHeader(key string, val string) {
	if c.headers == nil {
		c.headers = make(http.Header, 1)
	}
	c.headers.Add(key, val)
}

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

func (c *Client) ReplaceHeaders(h http.Header) {
	if h == nil {
		c.headers = make(http.Header)
		return
	}
	clone := make(http.Header, len(h))
	for k, vals := range h {
		// copy slice to avoid sharing backing arrays
		cp := make([]string, len(vals))
		copy(cp, vals)
		clone[k] = cp
	}
	c.headers = clone
}
