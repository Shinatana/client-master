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
	// Deprecated: use base instead.
	Headers Headers
	// Deprecated: use base instead.
	baseUrl    string
	base       *url.URL
	headers    http.Header
	httpClient http.Client
	lg         *zerolog.Logger
}

func NewHTTPClient(baseUrl string, opts ...Option) (*Client, error) {
	baseParsedURL, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid base url '%s': %w", baseUrl, err)
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

// todo refine headers funcs

func (c *Client) AddHeader(key, val string) {
	if c.Headers == nil {
		c.Headers = make(Headers, 1)
	}

	c.Headers[key] = val
}

func (c *Client) AddHeaders(headers Headers) {
	if headers == nil {
		return
	}
	if c.Headers == nil {
		c.Headers = make(Headers, len(headers))
	}
	for k, v := range headers {
		c.Headers[k] = v
	}
}

func (c *Client) ReplaceHeaders(headers Headers) {
	if headers == nil {
		c.Headers = make(Headers)
		return
	}
	cpy := make(Headers, len(headers))
	for k, v := range headers {
		cpy[k] = v
	}
	c.Headers = cpy
}
