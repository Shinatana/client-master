// Package client provides a small, focused HTTP client wrapper built on top of
// net/http. It standardizes base-URL handling, default headers, timeouts,
// logging, and exposes convenience methods for common HTTP verbs.
//
// Overview
//
//   - Base URL + path: A Client is constructed with a base URL. Each request
//     can supply an extra path that is safely joined to the base path, and
//     query parameters that are encoded into the URL.
//   - Headers: The client maintains default headers that are merged into every
//     request. Per-request headers are merged using Add semantics (values are
//     appended; existing keys are not replaced).
//   - Timeouts: A default timeout is applied to the underlying http.Client,
//     configurable via options.
//   - Context: All request methods accept a context to control cancellation and
//     deadlines.
//   - Logging: An optional zerolog.Logger can be supplied via options. By
//     default, a no-op logger is used. Normal (successful) requests are logged
//     at a debug level.
//
// # Construction
//
// Use NewHTTPClient to create a client. Functional options configure timeout,
// logging, and initial default headers. Notes about options:
//   - Order is not guaranteed. Do not rely on the relative order in which
//     options are applied.
//   - Passing multiple WithXxx options does not guarantee order or precedence.
//     Avoid specifying the same setting more than once; the effective value is
//     undefined if duplicates are provided.
//
// Example
//
//	cli, err := client.NewHTTPClient(
//		"https://api.example.com",
//		client.WithTimeout(5*time.Second),
//		client.WithHeaders(http.Header{
//			"User-Agent": []string{"my-app/1.0"},
//			"Accept":     []string{"application/json"},
//		}),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Add more default header values (appended, not replaced).
//	cli.AddHeader("X-Trace", []string{"abc123"})
//
//	// Send a GET request to /v1/items?id=42
//	ctx := context.Background()
//	resp, err := cli.Get(ctx, "/v1/items", url.Values{"id": []string{"42"}}, nil)
//	if err != nil {
//		// err may wrap ErrStatusCodeNotSuccess for non-2xx responses
//		log.Println("request failed:", err)
//	}
//	_ = resp
//
// # Request methods
//
// The client exposes SendRequest for full control and thin wrappers for common
// verbs: Get, Head, Options, Post, Put, Patch, Delete. Each method accepts:
//   - ctx: context for cancellation/deadline.
//   - path: extra path joined to the client's base URL path.
//   - params: query parameters (url.Values) encoded into the URL.
//   - headers: request-specific headers merged with the client's defaults.
//   - body: optional request body for methods that allow one.
//
// # Response and errors
//
// Response contains StatusCode, Body, and Headers for the HTTP response.
// Read failures wrap ErrFailedToReadResponseBody. Non-2xx responses result in
// an error that wraps ErrStatusCodeNotSuccess, returned alongside the Response
// so callers can inspect status, headers, or body for diagnostics.
// If reading the response body fails, a non-nil Response is still returned
// together with the error; it contains the status code and headers (the Body
// may be empty or nil depending on the failure).
//
// # Concurrency and mutability
//
// Sending requests concurrently is safe. However, mutating the client's default
// headers via AddHeader, AddHeaders, or ReplaceHeaders is not safe while other
// goroutines are issuing requests. Perform such mutations during setup or
// ensure external synchronization.
//
// # Defaults
//
// DefaultTimeout defines the package's default http.Client timeout when a custom
// timeout is not provided. A no-op logger is used unless WithLogger is supplied.
//
// # Deprecations
//
// The package includes legacy, deprecated types and helpers for backward
// compatibility. Prefer the new API (NewHTTPClient, SendRequest, and the thin
// wrappers) for all new code.
package client
