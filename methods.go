package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrFailedToReadResponseBody = fmt.Errorf("failed to read response body")
)

func (c *Client) SendRequest(ctx context.Context, method string, params url.Values,
	headers http.Header, body io.Reader) (*Response, error) {

	req, err := c.newRequestWithParams(ctx, method, params, headers, body)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare a request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send a request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {

		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
		}, fmt.Errorf("%w: %w", ErrFailedToReadResponseBody, err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       b,
		Headers:    resp.Header.Clone(),
	}, nil
}

// Methods without a request body.

func (c *Client) Get(ctx context.Context, params url.Values, headers http.Header) (*Response, error) {
	return c.SendRequest(ctx, http.MethodGet, params, headers, nil)
}

func (c *Client) Head(ctx context.Context, params url.Values, headers http.Header) (*Response, error) {
	return c.SendRequest(ctx, http.MethodHead, params, headers, nil)
}

func (c *Client) Options(ctx context.Context, params url.Values, headers http.Header) (*Response, error) {
	return c.SendRequest(ctx, http.MethodOptions, params, headers, nil)
}

// Methods that usually include a request body.

func (c *Client) Post(ctx context.Context, params url.Values, headers http.Header, body io.Reader) (*Response, error) {
	return c.SendRequest(ctx, http.MethodPost, params, headers, body)
}

func (c *Client) Put(ctx context.Context, params url.Values, headers http.Header, body io.Reader) (*Response, error) {
	return c.SendRequest(ctx, http.MethodPut, params, headers, body)
}

func (c *Client) Patch(ctx context.Context, params url.Values, headers http.Header, body io.Reader) (*Response, error) {
	return c.SendRequest(ctx, http.MethodPatch, params, headers, body)
}

func (c *Client) Delete(ctx context.Context, params url.Values, headers http.Header, body io.Reader) (*Response, error) {
	return c.SendRequest(ctx, http.MethodDelete, params, headers, body)
}
