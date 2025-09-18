package client

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	log := zerolog.Nop()
	c, err := New(baseURL, nil, &log, false, "orders-service/1.0")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	c.SetHeader("X-Global", "G")
	return c
}

func TestSendGet_SetsMethodPathHeadersAndUA(t *testing.T) {
	var gotMethod, gotPath, hGlobal, hLocal, hUA string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		hGlobal = r.Header.Get("X-Global")
		hLocal = r.Header.Get("X-Local")
		hUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)

	body, status, err := c.SendGet("/items", Params{"q": "1"}, Headers{"X-Local": "L"})
	if err != nil {
		t.Fatalf("SendGet error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("method=%s", gotMethod)
	}
	if gotPath != "/items?q=1" {
		t.Fatalf("path=%s", gotPath)
	}
	if hGlobal != "G" || hLocal != "L" {
		t.Fatalf("headers: X-Global=%s X-Local=%s", hGlobal, hLocal)
	}
	if hUA != "orders-service/1.0" {
		t.Fatalf("user-agent=%s", hUA)
	}
	if status == nil || *status != 200 || string(body) != "ok" {
		t.Fatalf("resp status/body: %v %s", status, string(body))
	}
}

func TestSendPost_SendsBodyAndHeaders(t *testing.T) {
	var gotMethod, gotPath, hGlobal, hLocal, hUA, gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		hGlobal = r.Header.Get("X-Global")
		hLocal = r.Header.Get("X-Local")
		hUA = r.Header.Get("User-Agent")
		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok")
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)

	body, status, err := c.SendPost("/create", []byte(`{"x":1}`), Params{"p": "2"}, Headers{"X-Local": "L"})
	if err != nil {
		t.Fatalf("SendPost error: %v", err)
	}
	if gotMethod != http.MethodPost || !strings.Contains(gotPath, "/create?p=2") {
		t.Fatalf("method/path: %s %s", gotMethod, gotPath)
	}
	if hGlobal != "G" || hLocal != "L" || hUA != "orders-service/1.0" {
		t.Fatalf("headers: %s %s %s", hGlobal, hLocal, hUA)
	}
	if !strings.Contains(gotBody, `"x":1`) {
		t.Fatalf("body=%s", gotBody)
	}
	if status == nil || *status != 200 || string(body) != "ok" {
		t.Fatalf("resp: %v %s", status, string(body))
	}
}

func TestSendPut_SendsBody(t *testing.T) {
	var gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		gotBody = string(b)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, status, err := c.SendPut("/upd", []byte(`{"a":2}`), nil, nil)
	if err != nil {
		t.Fatalf("SendPut error: %v", err)
	}
	if gotMethod != http.MethodPut || status == nil || *status != 200 {
		t.Fatalf("method/status: %s %v", gotMethod, status)
	}
	if gotBody != `{"a":2}` {
		t.Fatalf("body=%s", gotBody)
	}
}

func TestSendPatch_SendsBody(t *testing.T) {
	var gotMethod, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		gotBody = string(b)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, status, err := c.SendPatch("/p", []byte(`{"a":3}`), nil, nil)
	if err != nil {
		t.Fatalf("SendPatch error: %v", err)
	}
	if gotMethod != http.MethodPatch || status == nil || *status != 200 {
		t.Fatalf("method/status: %s %v", gotMethod, status)
	}
	if gotBody != `{"a":3}` {
		t.Fatalf("body=%s", gotBody)
	}
}

func TestSendDelete_SetsMethodAndParams(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, status, err := c.SendDelete("/d", Params{"x": "1"}, nil)
	if err != nil {
		t.Fatalf("SendDelete error: %v", err)
	}
	if gotMethod != http.MethodDelete || !strings.Contains(gotPath, "/d?x=1") {
		t.Fatalf("method/path: %s %s", gotMethod, gotPath)
	}
	if status == nil || *status != 200 {
		t.Fatalf("status=%v", status)
	}
}

func TestNon2xx_ReturnsErrorAndNoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadRequest) // 400
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	body, status, err := c.SendGet("/x", nil, nil)
	if err == nil {
		t.Fatal("expected error on non-2xx")
	}
	if status == nil || *status != http.StatusBadRequest {
		t.Fatalf("status=%v", status)
	}
	if body != nil {
		t.Fatalf("body must be nil on error")
	}
}
func TestCreateRequest_InvalidBaseURL(t *testing.T) {
	log := zerolog.Nop()
	c, err := New("http://[::1]:namedport", nil, &log, false, "ua")
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	_, _, err = c.SendGet("/x", nil, nil)
	if err == nil {
		t.Fatal("expected error on invalid baseURL, got nil")
	}
}

func TestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer srv.Close()

	one := 1
	log := zerolog.Nop()
	c, err := New(srv.URL, &one, &log, false, "ua")
	if err != nil {
		t.Fatalf("New error: %v", err)
	}

	start := time.Now()
	_, _, err = c.SendGet("/slow", nil, nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	elapsed := time.Since(start)
	if elapsed > 1500*time.Millisecond {
		t.Fatalf("timeout too long: %v", elapsed)
	}
}

func TestSendGet_TransportError(t *testing.T) {
	log := zerolog.Nop()
	c, err := New("http://127.0.0.1:1", nil, &log, false, "ua")
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	body, status, err := c.SendGet("/x", nil, nil)
	if err == nil {
		t.Fatal("expected transport error, got nil")
	}
	if body != nil || status != nil {
		t.Fatalf("want body=nil,status=nil on Do error, got body=%v status=%v", body, status)
	}
}

type errCloseBody struct{}

func (errCloseBody) Read(p []byte) (int, error) { return 0, io.EOF }
func (errCloseBody) Close() error               { return fmt.Errorf("close boom") }

func TestGetResponseBody_CloseErrorIsLoggedNotReturned(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       errCloseBody{},
	}
	log := zerolog.Nop()
	body, status, err := getResponseBody(resp, &log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status == nil || *status != http.StatusOK {
		t.Fatalf("status: %v", status)
	}
	if len(body) != 0 {
		t.Fatalf("body should be empty, got: %q", string(body))
	}
}
