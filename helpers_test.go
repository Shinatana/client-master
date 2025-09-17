package client

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr string
	}{
		{name: "GET uppercase", in: "GET", want: http.MethodGet},
		{name: "get lowercase", in: "get", want: http.MethodGet},
		{name: "POST mixed", in: "pOSt", want: http.MethodPost},
		{name: "PUT", in: "put", want: http.MethodPut},
		{name: "PATCH", in: "PATCH", want: http.MethodPatch},
		{name: "DELETE", in: "delete", want: http.MethodDelete},
		{name: "HEAD", in: "HeAd", want: http.MethodHead},
		{name: "OPTIONS", in: "options", want: http.MethodOptions},
		{name: "TRACE", in: "TrAcE", want: http.MethodTrace},
		{name: "CONNECT", in: "CONNECT", want: http.MethodConnect},
		{name: "empty -> error", in: "", wantErr: "http method is empty"},
		{name: "unsupported -> error", in: "FOO", wantErr: "unsupported http method"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := validateMethod(tc.in)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBuildURL(t *testing.T) {
	t.Parallel()

	t.Run("nil base -> error", func(t *testing.T) {
		t.Parallel()

		got, err := buildURL(nil, "", nil)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("no extra path or params -> copy of base", func(t *testing.T) {
		t.Parallel()

		base := mustParseURL(t, "https://example.com/api?x=1")
		got, err := buildURL(base, "", nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, base.Scheme, got.Scheme)
		assert.Equal(t, base.Host, got.Host)
		assert.Equal(t, base.Path, got.Path)
		assert.Equal(t, base.RawQuery, got.RawQuery)
		// Ensure it's a copy (modifying got should not modify base)
		got.Host = "changed.example.com"
		assert.Equal(t, "example.com", base.Host)
	})

	t.Run("extra path absolute joins with base path", func(t *testing.T) {
		t.Parallel()

		base := mustParseURL(t, "https://example.com/api")
		got, err := buildURL(base, "/v12//items", nil)
		require.NoError(t, err)
		assert.Equal(t, "/api/v12/items", got.Path)
	})

	t.Run("extra path relative joins with base path", func(t *testing.T) {
		t.Parallel()

		base := mustParseURL(t, "https://example.com/api")
		got, err := buildURL(base, "v2/items", nil)
		require.NoError(t, err)
		assert.Equal(t, "/api/v2/items", got.Path)
	})

	t.Run("params appended to existing query using Add semantics", func(t *testing.T) {
		t.Parallel()

		base := mustParseURL(t, "https://example.com/api?foo=1")
		params := url.Values{
			"foo": []string{"2"},
			"bar": []string{"a", "b"},
		}
		got, err := buildURL(base, "", params)
		require.NoError(t, err)
		require.NotNil(t, got)

		assertQueryIncludes(t, got, url.Values{
			"foo": []string{"1", "2"},
			"bar": []string{"a", "b"},
		})
	})
}

func TestMergeHeaders(t *testing.T) {
	t.Parallel()

	t.Run("both nil -> new empty header", func(t *testing.T) {
		t.Parallel()

		got := mergeHeaders(nil, nil)
		require.NotNil(t, got)
		assert.Len(t, got, 0)
	})

	t.Run("base only -> copy values", func(t *testing.T) {
		t.Parallel()

		base := http.Header{
			"X-A": {"1", "2"},
			"X-B": {"b"},
		}
		got := mergeHeaders(base, nil)
		require.NotNil(t, got)
		assert.ElementsMatch(t, []string{"1", "2"}, got.Values("X-A"))
		assert.ElementsMatch(t, []string{"b"}, got.Values("X-B"))

		// Modifying result should not affect base
		got.Add("X-A", "3")
		assert.ElementsMatch(t, []string{"1", "2"}, base["X-A"])
	})

	t.Run("extra only -> copy values", func(t *testing.T) {
		t.Parallel()

		extra := http.Header{
			"Y": {"a"},
		}
		got := mergeHeaders(nil, extra)
		require.NotNil(t, got)
		assert.ElementsMatch(t, []string{"a"}, got.Values("Y"))
	})

	t.Run("both -> values appended (Add semantics)", func(t *testing.T) {
		t.Parallel()

		base := http.Header{
			"K": {"b1", "b2"},
			"A": {"a1"},
		}
		extra := http.Header{
			"K": {"e1"},
			"Z": {"z1", "z2"},
		}
		got := mergeHeaders(base, extra)

		// Verify presence and full set of values (order within a key: base then extra)
		assert.Equal(t, []string{"b1", "b2", "e1"}, got["K"])
		assert.Equal(t, []string{"a1"}, got["A"])
		assert.ElementsMatch(t, []string{"z1", "z2"}, got["Z"])
	})
}

func TestClient_newRequestWithParams(t *testing.T) {
	t.Parallel()

	t.Run("success builds request with merged headers, query and method", func(t *testing.T) {
		t.Parallel()

		c := &Client{
			base:    mustParseURL(t, "https://api.example.com/base"),
			headers: http.Header{"X-Def": {"A"}, "X-Both": {"Base"}},
		}
		ctx := context.Background()
		params := url.Values{"p": {"1", "2"}}
		headers := http.Header{"X-Req": {"B"}, "X-Both": {"Req"}}

		req, err := c.newRequestWithParams(ctx, "get", "v1/items", params, headers, io.NopCloser(strings.NewReader("")).(io.Reader))
		require.NoError(t, err)
		require.NotNil(t, req)

		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "https", req.URL.Scheme)
		assert.Equal(t, "api.example.com", req.URL.Host)
		assert.Equal(t, "/base/v1/items", req.URL.Path)
		assertQueryIncludes(t, req.URL, url.Values{"p": {"1", "2"}})

		// Headers merged with Add semantics
		assert.ElementsMatch(t, []string{"A"}, req.Header.Values("X-Def"))
		assert.ElementsMatch(t, []string{"B"}, req.Header.Values("X-Req"))
		assert.Equal(t, []string{"Base", "Req"}, req.Header["X-Both"])

		// Context attached
		assert.NotNil(t, req.Context())
	})

	t.Run("invalid method -> error", func(t *testing.T) {
		t.Parallel()

		c := &Client{
			base: mustParseURL(t, "https://api.example.com"),
		}
		req, err := c.newRequestWithParams(context.Background(), "INVALID", "", nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, req)
	})

	t.Run("nil base -> error", func(t *testing.T) {
		t.Parallel()

		c := &Client{
			base: nil,
		}
		req, err := c.newRequestWithParams(context.Background(), "GET", "", nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, req)
	})

	t.Run("invalid URL for http.NewRequestWithContext -> error", func(t *testing.T) {
		t.Parallel()

		c := &Client{
			base: &url.URL{Scheme: "http", Host: "bad host"},
		}
		req, err := c.newRequestWithParams(context.Background(), "GET", "", nil, nil, nil)
		assert.Error(t, err, "expected error from http.NewRequestWithContext due to invalid URL")
		assert.Nil(t, req)
	})
}

/**************
 Helpers
**************/

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return u
}

func assertQueryIncludes(t *testing.T, u *url.URL, want url.Values) {
	t.Helper()
	got := u.Query()
	for k, wantVals := range want {
		gotVals := got[k]
		require.NotNil(t, gotVals, "missing query key %q", k)
		assert.ElementsMatch(t, wantVals, gotVals, "query values mismatch for key %q", k)
	}
}
