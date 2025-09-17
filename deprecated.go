package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog"
)

// New func
// DEPRECATED: use NewHTTPClient instead
func New(baseUrl string, timeout *int) *Client {
	tt := defaultTimeout

	if timeout != nil {
		tt = *timeout
	}

	nop := zerolog.Nop()

	return &Client{
		Headers: Headers{},
		baseUrl: baseUrl,
		httpClient: http.Client{
			Timeout: time.Second * time.Duration(tt),
		},
		lg: &nop,
	}
}

// SetHeader func
// DEPRECATED: use AddHeader instead
func (client *Client) SetHeader(key, val string) *Client {
	client.Headers[key] = val

	return client
}

// SendGet func
// DEPRECATED: use NewHTTPClient instead
func (client *Client) SendGet(path string, params Params, headers Headers) ([]byte, *int, error) {
	request, err := client.createRequest(http.MethodGet, path, params, nil)

	if err != nil {
		return nil, nil, err
	}

	client.fillRequestHeaders(request, headers)

	var response *http.Response

	response, err = client.getResponse(request)

	if err != nil {
		return nil, nil, err
	}

	return getResponseBody(response)
}

// SendPost func
// DEPRECATED: use NewHTTPClient instead
func (client *Client) SendPost(
	path string,
	jsonData []byte,
	queryParams Params,
	headers Headers,
) ([]byte, *int, error) {
	request, err := client.createRequest(http.MethodPost, path, queryParams, jsonData)

	if err != nil {
		return nil, nil, err
	}

	client.fillRequestHeaders(request, headers)

	var response *http.Response

	response, err = client.getResponse(request)

	if err != nil {
		return nil, nil, err
	}

	return getResponseBody(response)
}

// prepareUrlWithParams func
// DEPRECATED: use NewHTTPClient instead
func (client *Client) prepareUrlWithParams(path string, dirtyParams Params) (string, error) {
	params := url.Values{}

	for key, val := range dirtyParams {
		params.Add(key, val)
	}

	u, err := url.ParseRequestURI(client.baseUrl)

	if err != nil {
		return "", err
	}

	u.Path = path
	u.RawQuery = params.Encode()

	return fmt.Sprintf("%v", u), err
}

// createRequest func
// DEPRECATED: use NewHTTPClient instead
func (client *Client) createRequest(
	method string,
	path string,
	queryParams Params,
	jsonData []byte,
) (*http.Request, error) {
	var preparedUrl string
	var err error

	if len(queryParams) < 1 {
		preparedUrl = client.baseUrl + path
	} else {
		preparedUrl, err = client.prepareUrlWithParams(path, queryParams)

		if err != nil {
			return nil, err
		}
	}

	return http.NewRequest(method, preparedUrl, bytes.NewBuffer(jsonData))
}

// getResponse func
// DEPRECATED: use NewHTTPClient instead
func (client *Client) getResponse(request *http.Request) (*http.Response, error) {
	response, err := client.httpClient.Do(request)

	if err != nil {
		return nil, err
	}

	return response, nil
}

// closeResponseBody func
// DEPRECATED: use NewHTTPClient instead
func closeResponseBody(response *http.Response) {
	err := response.Body.Close()

	if err != nil {
		panic(err)
	}
}

// getResponseBody func
// DEPRECATED: use NewHTTPClient instead
func getResponseBody(response *http.Response) ([]byte, *int, error) {
	defer func() {
		closeResponseBody(response)
	}()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, &response.StatusCode, err
	}

	return body, &response.StatusCode, nil
}

// fillRequestHeaders func
// DEPRECATED: use NewHTTPClient instead
func (client *Client) fillRequestHeaders(r *http.Request, headers Headers) *Client {
	for key, val := range client.Headers {
		r.Header.Add(key, val)
	}

	for key, val := range headers {
		r.Header.Add(key, val)
	}

	return client
}
