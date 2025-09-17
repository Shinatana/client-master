package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRequest(t *testing.T) {
	t.Parallel()

	t.Run("success: builds request, merges headers, encodes params, returns body and headers", func(t *testing.T) {
		t.Parallel()

		var gotReq *http.Request
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotReq = r
			// Assert method and path/query inside handler as additional safety
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/v1/items", r.URL.Path)
			q := r.URL.Query()
			assert.ElementsMatch(t, []string{"1", "2"}, q["p"])
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, []string{"defA", "defB"}, r.Header["X-Default"])
			assert.Equal(t, []string{"reqA"}, r.Header["X-Req"])

			w.Header().Set("X-Resp", "yes")
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		defer srv.Close()

		base := mustParseURL(t, srv.URL+"/api")
		cli := mustNewClient(t, base.String(),
			WithTimeout(2*time.Second),
			WithHeaders(http.Header{
				"X-Default":      {"defA", "defB"},
				"Content-Type":   {"application/json"},
				"X-Should-Merge": {"client"},
			}),
		)

		params := url.Values{"p": {"1", "2"}}
		reqHdr := http.Header{
			"X-Req":          {"reqA"},
			"X-Should-Merge": {"request"},
		}
		body := strings.NewReader(`{"in":true}`)
		resp, err := cli.SendRequest(context.Background(), http.MethodPost, "v1/items", params, reqHdr, body)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Response assertions
		assert.Equal(t, 200, resp.StatusCode)
		require.NotNil(t, resp.Body)
		assert.JSONEq(t, `{"ok":true}`, string(resp.Body))
		require.NotNil(t, resp.Headers)
		assert.Equal(t, "yes", resp.Headers.Get("X-Resp"))

		// Ensure request actually sent had merged headers (server asserted part already).
		require.NotNil(t, gotReq)
		// Both client default and request-specific headers must be present (add semantics)
		assert.Equal(t, []string{"client", "request"}, gotReq.Header["X-Should-Merge"])
	})

	t.Run("non-2xx: returns response and wraps ErrStatusCodeNotSuccess", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Resp", "nope")
			w.WriteHeader(418) // I'm a teapot
			_, _ = w.Write([]byte("not ok"))
		}))
		defer srv.Close()

		cli := mustNewClient(t, srv.URL)

		resp, err := cli.SendRequest(context.Background(), http.MethodGet, "", nil, nil, nil)
		require.Error(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 418, resp.StatusCode)
		assert.Equal(t, "nope", resp.Headers.Get("X-Resp"))
		assert.True(t, errors.Is(err, ErrStatusCodeNotSuccess), "error should wrap ErrStatusCodeNotSuccess")
	})

	t.Run("http.Client.Do error is propagated", func(t *testing.T) {
		t.Parallel()

		cli := mustNewClient(t, "http://example.invalid")
		cli.httpClient = http.Client{ // custom transport that always errors
			Transport: roundTripperErr{err: errors.New("boom")},
		}

		resp, err := cli.SendRequest(context.Background(), http.MethodGet, "", nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to send a request")
	})

	t.Run("response body read failure returns response with headers and wraps ErrFailedToReadResponseBody", func(t *testing.T) {
		t.Parallel()

		cli := mustNewClient(t, "http://example.invalid")
		cli.httpClient = http.Client{
			Transport: roundTripperBodyErr{
				status: 200,
				header: http.Header{"X-Resp": {"hdr"}},
				err:    errors.New("read fail"),
			},
		}

		resp, err := cli.SendRequest(context.Background(), http.MethodGet, "", nil, nil, nil)
		require.Error(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "hdr", resp.Headers.Get("X-Resp"))
		assert.True(t, errors.Is(err, ErrFailedToReadResponseBody), "error should wrap ErrFailedToReadResponseBody")
	})

	t.Run("invalid method -> error from request preparation", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}))
		defer srv.Close()

		cli := mustNewClient(t, srv.URL)
		resp, err := cli.SendRequest(context.Background(), "INVALID", "", nil, nil, nil)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to prepare a request")
	})

	t.Run("URL path joining and query encoding with base path", func(t *testing.T) {
		t.Parallel()

		var gotPath string
		var gotQuery url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotQuery = r.URL.Query()
			w.WriteHeader(204)
		}))
		defer srv.Close()

		cli := mustNewClient(t, srv.URL+"/base")
		params := url.Values{"a": {"1"}, "b": {"x", "y"}}
		resp, err := cli.SendRequest(context.Background(), http.MethodGet, "v1/../v2/items", params, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 204, resp.StatusCode)

		// Even with success/non-success, the server captured the path and query.
		assert.Equal(t, "/base/v2/items", gotPath)
		assert.ElementsMatch(t, []string{"1"}, gotQuery["a"])
		assert.ElementsMatch(t, []string{"x", "y"}, gotQuery["b"])
	})
}

func TestThinWrappers(t *testing.T) {
	t.Parallel()

	type rec struct {
		method   string
		bodySize int
	}

	tests := []struct {
		name       string
		wantMethod string
		wantBody   bool
		call       func(c *Client) (*Response, error)
	}{
		{
			name:       "Get_no_body",
			wantMethod: http.MethodGet,
			wantBody:   false,
			call: func(c *Client) (*Response, error) {
				return c.Get(context.Background(), "/v1/items", nil, nil)
			},
		},
		{
			name:       "Head_no_body",
			wantMethod: http.MethodHead,
			wantBody:   false,
			call: func(c *Client) (*Response, error) {
				return c.Head(context.Background(), "/v1/items", nil, nil)
			},
		},
		{
			name:       "Options_no_body",
			wantMethod: http.MethodOptions,
			wantBody:   false,
			call: func(c *Client) (*Response, error) {
				return c.Options(context.Background(), "/v1/items", nil, nil)
			},
		},
		{
			name:       "Post_with_body",
			wantMethod: http.MethodPost,
			wantBody:   true,
			call: func(c *Client) (*Response, error) {
				return c.Post(context.Background(), "/v1/items", nil, nil, strings.NewReader("x"))
			},
		},
		{
			name:       "Put_with_body",
			wantMethod: http.MethodPut,
			wantBody:   true,
			call: func(c *Client) (*Response, error) {
				return c.Put(context.Background(), "/v1/items", nil, nil, strings.NewReader("x"))
			},
		},
		{
			name:       "Patch_with_body",
			wantMethod: http.MethodPatch,
			wantBody:   true,
			call: func(c *Client) (*Response, error) {
				return c.Patch(context.Background(), "/v1/items", nil, nil, strings.NewReader("x"))
			},
		},
		{
			name:       "Delete_with_body",
			wantMethod: http.MethodDelete,
			wantBody:   true,
			call: func(c *Client) (*Response, error) {
				return c.Delete(context.Background(), "/v1/items", nil, nil, strings.NewReader("x"))
			},
		},
		{
			name:       "Delete_no_body",
			wantMethod: http.MethodDelete,
			wantBody:   false,
			call: func(c *Client) (*Response, error) {
				return c.Delete(context.Background(), "/v1/items", nil, nil, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var rrec rec
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rrec.method = r.Method
				if r.Body != nil {
					b, _ := io.ReadAll(r.Body)
					rrec.bodySize = len(b)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			c := mustNewClient(t, srv.URL)
			resp, err := tc.call(c)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			assert.Equal(t, tc.wantMethod, rrec.method)
			if tc.wantBody {
				assert.NotZero(t, rrec.bodySize, "expected non-empty request body")
			} else {
				assert.Zero(t, rrec.bodySize, "expected empty/nil request body")
			}
		})
	}
}

/**************
 Helpers
**************/

func mustNewClient(t *testing.T, base string, opts ...Option) *Client {
	t.Helper()
	c, err := NewHTTPClient(base, opts...)
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

type roundTripperErr struct {
	err error
}

func (r roundTripperErr) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, r.err
}

type readErrBody struct {
	err error
}

func (b readErrBody) Read(_ []byte) (int, error) { return 0, b.err }
func (b readErrBody) Close() error               { return nil }

type roundTripperBodyErr struct {
	status int
	header http.Header
	err    error
}

func (r roundTripperBodyErr) RoundTrip(_ *http.Request) (*http.Response, error) {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return &http.Response{
		StatusCode: r.status,
		Header:     r.header,
		Body:       readErrBody{err: r.err},
	}, nil
}
