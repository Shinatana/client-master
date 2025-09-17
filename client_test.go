package client

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	t.Run("invalid base URL -> error", func(t *testing.T) {
		t.Parallel()

		c, err := NewHTTPClient("://bad url")
		require.Error(t, err)
		assert.Nil(t, c)
	})

	t.Run("defaults applied when no options", func(t *testing.T) {
		t.Parallel()

		c, err := NewHTTPClient("https://example.com")
		require.NoError(t, err)
		require.NotNil(t, c)

		// Default timeout
		assert.Equal(t, DefaultTimeout, c.httpClient.Timeout)

		// Default headers initialized to non-nil empty
		require.NotNil(t, c.headers)
		assert.Empty(t, c.headers)

		// Logger non-nil (no-op by default)
		require.NotNil(t, c.lg)
	})

	t.Run("options propagated (timeout, headers, logger)", func(t *testing.T) {
		t.Parallel()

		lg := zerolog.New(io.Discard)
		h := http.Header{
			"User-Agent": {"my-app/1.0"},
			"Accept":     {"application/json"},
		}
		c, err := NewHTTPClient(
			"https://example.com/base",
			WithTimeout(1234*time.Millisecond),
			WithHeaders(h),
			WithLogger(&lg),
		)
		require.NoError(t, err)
		require.NotNil(t, c)

		assert.Equal(t, 1234*time.Millisecond, c.httpClient.Timeout)
		assert.Equal(t, h, c.headers)
		assert.Same(t, &lg, c.lg)
	})
}

func TestClient_AddHeader(t *testing.T) {
	t.Parallel()

	t.Run("initializes map if nil and adds values", func(t *testing.T) {
		t.Parallel()

		c := &Client{}
		c.AddHeader("X-Test", []string{"a", "b"})

		require.NotNil(t, c.headers)
		assert.Equal(t, []string{"a", "b"}, c.headers["X-Test"])
	})

	t.Run("appends values (Add semantics) without replacing", func(t *testing.T) {
		t.Parallel()

		c := &Client{headers: http.Header{"X": {"1"}}}
		c.AddHeader("X", []string{"2", "3"})

		assert.Equal(t, []string{"1", "2", "3"}, c.headers["X"])
	})
}

func TestClient_AddHeaders(t *testing.T) {
	t.Parallel()

	t.Run("nil input -> no-op and no panic", func(t *testing.T) {
		t.Parallel()

		c := &Client{}
		require.NotPanics(t, func() {
			c.AddHeaders(nil)
		})

		assert.Nil(t, c.headers, "headers map should remain nil on nil input")
	})

	t.Run("initializes map if nil and copies values", func(t *testing.T) {
		t.Parallel()

		c := &Client{}
		in := http.Header{"A": {"1"}, "B": {"x", "y"}}
		c.AddHeaders(in)

		require.NotNil(t, c.headers)
		assert.ElementsMatch(t, []string{"1"}, c.headers["A"])
		assert.ElementsMatch(t, []string{"x", "y"}, c.headers["B"])

		// Modifying input should not affect stored values (since Add copies by value)
		in.Add("A", "2")
		assert.ElementsMatch(t, []string{"1"}, c.headers["A"])
	})

	t.Run("merges with existing using Add semantics", func(t *testing.T) {
		t.Parallel()

		c := &Client{headers: http.Header{"K": {"b1"}, "A": {"a1"}}}
		in := http.Header{"K": {"e1", "e2"}, "Z": {"z"}}
		c.AddHeaders(in)

		assert.Equal(t, []string{"b1", "e1", "e2"}, c.headers["K"])
		assert.Equal(t, []string{"a1"}, c.headers["A"])
		assert.Equal(t, []string{"z"}, c.headers["Z"])
	})
}

func TestClient_ReplaceHeaders(t *testing.T) {
	t.Parallel()

	t.Run("nil input -> reset to empty non-nil map", func(t *testing.T) {
		t.Parallel()

		initial := http.Header{"A": {"1"}}
		c := &Client{headers: initial}
		c.ReplaceHeaders(nil)

		require.NotNil(t, c.headers)
		assert.Equal(t, initial, c.headers)
	})

	t.Run("nil input + nil value -> nil", func(t *testing.T) {
		t.Parallel()

		c := &Client{headers: nil}
		c.ReplaceHeaders(nil)

		require.Nil(t, c.headers)
	})

	t.Run("clones input map and slices (no aliasing)", func(t *testing.T) {
		t.Parallel()

		src := http.Header{
			"K": {"v1", "v2"},
			"A": {"a1"},
		}
		c := &Client{}
		c.ReplaceHeaders(src)

		// Same values at time of replace
		require.NotNil(t, c.headers)
		assert.Equal(t, []string{"v1", "v2"}, c.headers["K"])
		assert.Equal(t, []string{"a1"}, c.headers["A"])

		// Mutate src: ensure client headers not affected (cloned)
		src.Add("K", "v3")
		src["A"][0] = "aX"
		assert.Equal(t, []string{"v1", "v2"}, c.headers["K"])
		assert.Equal(t, []string{"a1"}, c.headers["A"])

		// Mutate client: ensure src not affected
		c.headers.Add("K", "client")
		assert.Equal(t, []string{"v1", "v2", "v3"}, src["K"])
	})
}

/*
*************

	Helpers

*************
*/
func mustNewClient(t *testing.T, base string, opts ...Option) *Client {
	t.Helper()
	c, err := NewHTTPClient(base, opts...)
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}
