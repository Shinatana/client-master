package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (c *Client) SendRequest(ctx context.Context, method string, params url.Values,
	headers http.Header, body io.Reader) (*Response, error) {

	req, err := c.newRequestWithParams(ctx, method, params, headers, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
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
		}, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       b,
		Headers:    resp.Header.Clone(),
	}, nil
}
