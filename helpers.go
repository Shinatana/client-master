package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// validateMethod returns a canonical uppercase HTTP method if supported,
// or an error if the method is empty or not one of the standard methods.
func validateMethod(method string) (string, error) {
	if method == "" {
		return "", errors.New("http method is empty")
	}
	m := strings.ToUpper(method)
	switch m {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodConnect:
		return m, nil
	default:
		return "", fmt.Errorf("unsupported http method %q", method)
	}
}

// buildURL constructs a URL by copying base, optionally joining extraPath to the
// base path, and appending params as the encoded query string. It returns an
// error if base is nil.
func buildURL(base *url.URL, extraPath string, params url.Values) (*url.URL, error) {
	if base == nil {
		return nil, errors.New("base URL is nil")
	}

	u := *base

	if extraPath != "" {
		u.Path = path.Join(u.Path, extraPath)
	}

	if len(params) > 0 {
		q := u.Query()
		for k, vals := range params {
			for _, v := range vals {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}

	return &u, nil
}

// mergeHeaders merges two header maps into a new http.Header.
// Values from both inputs are appended using Add semantics.
func mergeHeaders(base, extra http.Header) http.Header {
	merged := make(http.Header)

	for k, vals := range base {
		for _, v := range vals {
			merged.Add(k, v)
		}
	}

	for k, vals := range extra {
		for _, v := range vals {
			merged.Add(k, v)
		}
	}

	return merged
}

// newRequestWithParams creates a new *http.Request using the client's base URL,
// joining path, encoding params as the query string, merging default and
// request-specific headers, and attaching body. It validates the HTTP method
// and binds the request to the provided context.
func (c *Client) newRequestWithParams(ctx context.Context, method string, path string,
	params url.Values, headers http.Header, body io.Reader) (*http.Request, error) {
	u, err := buildURL(c.base, path, params)
	if err != nil {
		return nil, err
	}

	m, err := validateMethod(method)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, m, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header = mergeHeaders(c.headers, headers)

	return req, nil
}
