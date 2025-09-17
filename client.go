package client

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultTimeout = 10
	DefaultTimeout = 10 * time.Second
)

type Client struct {
	Headers    Headers
	baseUrl    string
	httpClient http.Client
	lg         *zerolog.Logger
}

func NewHTTPClient(baseUrl string, opts ...Option) *Client {
	suppliedOptions := applyOptions(opts...)

	return &Client{
		Headers: suppliedOptions.headers,
		baseUrl: baseUrl,
		httpClient: http.Client{
			Timeout: suppliedOptions.timeout,
		},
		lg: suppliedOptions.lg,
	}
}

func (client *Client) AddHeader(key, val string) {
	if client.Headers == nil {
		client.Headers = make(Headers, 1)
	}

	client.Headers[key] = val
}

func (client *Client) AddHeaders(headers Headers) {
	if headers == nil {
		return
	}
	if client.Headers == nil {
		client.Headers = make(Headers, len(headers))
	}
	for k, v := range headers {
		client.Headers[k] = v
	}
}

func (client *Client) ReplaceHeaders(headers Headers) {
	if headers == nil {
		client.Headers = make(Headers)
		return
	}
	cpy := make(Headers, len(headers))
	for k, v := range headers {
		cpy[k] = v
	}
	client.Headers = cpy
}
