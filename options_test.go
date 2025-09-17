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

func TestNormalizeLogger(t *testing.T) {
	t.Parallel()

	t.Run("nil logger returns non-nil no-op", func(t *testing.T) {
		t.Parallel()

		got := normalizeLogger(nil)
		assert.NotNil(t, got, "normalizeLogger(nil) must not return nil")
		assert.IsType(t, zerolog.Logger{}, *got, "returned value must be a zerolog.Logger")
	})

	t.Run("non-nil logger is returned as-is", func(t *testing.T) {
		t.Parallel()

		l := zerolog.New(io.Discard)
		got := normalizeLogger(&l)
		assert.Same(t, &l, got, "custom logger pointer should be preserved")
	})
}

func TestNormalizeTimeout(t *testing.T) {
	t.Parallel()

	t.Run("zero timeout -> DefaultTimeout", func(t *testing.T) {
		t.Parallel()

		got := normalizeTimeout(0)
		assert.Equal(t, DefaultTimeout, got)
	})

	t.Run("non-zero timeout preserved", func(t *testing.T) {
		t.Parallel()

		want := 5 * time.Second
		got := normalizeTimeout(want)
		assert.Equal(t, want, got)
	})
}

func TestNormalizeHeaders(t *testing.T) {
	t.Parallel()

	t.Run("nil headers -> non-nil empty map", func(t *testing.T) {
		t.Parallel()

		got := normalizeHeaders(nil)
		require.NotNil(t, got)
		assert.Equal(t, 0, len(got))
	})

	t.Run("non-nil headers preserved (same reference)", func(t *testing.T) {
		t.Parallel()

		h := http.Header{
			"X-Test": []string{"a"},
		}
		got := normalizeHeaders(h)
		assert.Equal(t, h, got)
	})
}

func TestApplyOptions(t *testing.T) {
	t.Parallel()

	t.Run("Defaults", func(t *testing.T) {
		t.Parallel()

		o := applyOptions() // no options

		require.NotNil(t, o.lg)
		assert.Equal(t, DefaultTimeout, o.timeout)
		require.NotNil(t, o.headers)
		assert.Empty(t, o.headers)
	})

	t.Run("WithAll", func(t *testing.T) {
		t.Parallel()

		lg := zerolog.New(io.Discard)
		timeout := 1234 * time.Millisecond
		headers := http.Header{"A": []string{"1"}, "B": []string{"2"}}

		o := applyOptions(
			WithLogger(&lg),
			WithTimeout(timeout),
			WithHeaders(headers),
		)

		require.NotNil(t, o.lg)
		assert.Same(t, &lg, o.lg, "logger pointer should be retained")
		assert.Equal(t, timeout, o.timeout, "timeout should be taken as-is")
		require.NotNil(t, o.headers)
		assert.Equal(t, headers, o.headers, "headers should be the same map content")
	})

	t.Run("WithHeadersNilInput_Normalized", func(t *testing.T) {
		t.Parallel()

		o := applyOptions(WithHeaders(nil))
		require.NotNil(t, o.headers)
		assert.Empty(t, o.headers)
	})

	t.Run("WithTimeoutZero_UsesDefault", func(t *testing.T) {
		t.Parallel()

		o := applyOptions(WithTimeout(0))
		assert.Equal(t, DefaultTimeout, o.timeout)
	})
}
