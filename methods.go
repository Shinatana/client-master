package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	// ErrFailedToReadResponseBody indicates that reading the HTTP response body
	// failed. The returned error will wrap the underlying I/O error.
	ErrFailedToReadResponseBody = fmt.Errorf("failed to read response body")
	// ErrStatusCodeNotSuccess indicates that the HTTP response status code is not
	// in the 2xx success range. The returned error message includes the status code.
	ErrStatusCodeNotSuccess = fmt.Errorf("status code is not success")
)

// SendRequest builds and sends an HTTP request.
// method must be a valid HTTP method (e.g., GET, POST).
// path is joined with the client's base URL path.
// params are encoded into the query string.
// headers are merged with the client's default headers (added, not replaced).
// body is an optional request body; it may be nil for methods without a body.
// The returned Response includes the status code, headers, and body. If the status
// code is not in the 2xx range, an error wrapping ErrStatusCodeNotSuccess is returned
// alongside the Response. The provided context controls request cancellation and deadline.
func (c *Client) SendRequest(ctx context.Context, method string, path string,
	params url.Values, headers http.Header, body io.Reader) (*Response, error) {

	start := time.Now()

	req, err := c.newRequestWithParams(ctx, method, path, params, headers, body)
	if err != nil {
		c.lg.Error().Err(err).
			Str("method", method).
			Str("path", c.base.EscapedPath()).
			Str("query", params.Encode()).
			Msg("failed to prepare request")
		return nil, fmt.Errorf("failed to prepare a request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.lg.Error().Err(err).
			Str("method", method).
			Str("url", req.URL.String()).
			Dur("duration", time.Since(start)).
			Msg("failed to send request")
		return nil, fmt.Errorf("failed to send a request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.lg.Warn().Err(err).
				Str("method", method).
				Str("url", req.URL.String()).
				Msg("failed to close response body")
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		c.lg.Error().Err(err).
			Str("method", method).
			Str("url", req.URL.String()).
			Int("status", resp.StatusCode).
			Dur("duration", time.Since(start)).
			Msg("failed to read response body")
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
		}, fmt.Errorf("%w: %w", ErrFailedToReadResponseBody, err)
	}

	c.lg.Debug().
		Str("method", method).
		Str("url", req.URL.String()).
		Int("status", resp.StatusCode).
		Dur("duration", time.Since(start)).
		Int("resp_bytes", len(b)).
		Msg("request completed")

	res := &Response{
		StatusCode: resp.StatusCode,
		Body:       b,
		Headers:    resp.Header.Clone(),
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return res, fmt.Errorf("%w: %d", ErrStatusCodeNotSuccess, res.StatusCode)
	}

	return res, nil
}

// Get sends an HTTP GET request.
// path is joined with the client's base URL path.
// params are encoded into the query string.
// headers are merged with the client's default headers.
func (c *Client) Get(ctx context.Context, path string, params url.Values,
	headers http.Header) (*Response, error) {

	return c.SendRequest(ctx, http.MethodGet, path, params, headers, nil)
}

// Head sends an HTTP HEAD request.
// path is joined with the client's base URL path.
// params are encoded into the query string.
// headers are merged with the client's default headers.
func (c *Client) Head(ctx context.Context, path string, params url.Values,
	headers http.Header) (*Response, error) {

	return c.SendRequest(ctx, http.MethodHead, path, params, headers, nil)
}

// Options sends an HTTP OPTIONS request.
// path is joined with the client's base URL path.
// params are encoded into the query string.
// headers are merged with the client's default headers.
func (c *Client) Options(ctx context.Context, path string, params url.Values,
	headers http.Header) (*Response, error) {

	return c.SendRequest(ctx, http.MethodOptions, path, params, headers, nil)
}

// Post sends an HTTP POST request with an optional body.
// path is joined with the client's base URL path.
// params are encoded into the query string.
// headers are merged with the client's default headers.
// body may be nil if the endpoint does not require a body.
func (c *Client) Post(ctx context.Context, path string, params url.Values,
	headers http.Header, body io.Reader) (*Response, error) {

	return c.SendRequest(ctx, http.MethodPost, path, params, headers, body)
}

// Put sends an HTTP PUT request with an optional body.
// path is joined with the client's base URL path.
// params are encoded into the query string.
// headers are merged with the client's default headers.
// body may be nil if the endpoint does not require a body.
func (c *Client) Put(ctx context.Context, path string, params url.Values,
	headers http.Header, body io.Reader) (*Response, error) {

	return c.SendRequest(ctx, http.MethodPut, path, params, headers, body)
}

// Patch sends an HTTP PATCH request with an optional body.
// path is joined with the client's base URL path.
// params are encoded into the query string.
// headers are merged with the client's default headers.
// body may be nil if the endpoint does not require a body.
func (c *Client) Patch(ctx context.Context, path string, params url.Values,
	headers http.Header, body io.Reader) (*Response, error) {

	return c.SendRequest(ctx, http.MethodPatch, path, params, headers, body)
}

// Delete sends an HTTP DELETE request with an optional body.
// path is joined with the client's base URL path.
// params are encoded into the query string.
// headers are merged with the client's default headers.
// body may be nil if the endpoint does not require a body.
func (c *Client) Delete(ctx context.Context, path string, params url.Values,
	headers http.Header, body io.Reader) (*Response, error) {

	return c.SendRequest(ctx, http.MethodDelete, path, params, headers, body)
}
