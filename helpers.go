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

func buildURL(base *url.URL, extraPath string, params url.Values) (*url.URL, error) {
	if base == nil {
		return nil, errors.New("base URL is nil")
	}

	u := *base

	if extraPath != "" {
		if strings.HasPrefix(extraPath, "/") {
			u.Path = path.Clean(extraPath)
		} else {
			u.Path = path.Join(u.Path, extraPath)
		}
	}

	if params != nil && len(params) > 0 {
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

func mergeHeaders(base, extra http.Header) http.Header {
	merged := make(http.Header)

	if base != nil {
		for k, vals := range base {
			for _, v := range vals {
				merged.Add(k, v)
			}
		}
	}

	if extra != nil {
		for k, vals := range extra {
			for _, v := range vals {
				merged.Add(k, v)
			}
		}
	}

	return merged
}

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
