package client

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultTimeout = 10
)

type Client struct {
	Headers    Headers
	baseUrl    string
	httpClient http.Client
	logger     *zerolog.Logger
}

func New(baseUrl string, timeout *int, log *zerolog.Logger, nolog bool) (*Client, error) {
	if log == nil && !nolog {
		return nil, errors.New("no logger provided")
	}

	if nolog {
		tmp := zerolog.Nop()
		log = &tmp
	}

	tt := defaultTimeout

	if timeout != nil {
		tt = *timeout
	}

	return &Client{
		Headers: Headers{},
		baseUrl: baseUrl,
		httpClient: http.Client{
			Timeout: time.Second * time.Duration(tt),
		},
		logger: log,
	}, nil
}

func (client *Client) SetHeader(key, val string) *Client {
	client.Headers[key] = val

	return client
}

func (client *Client) fillRequestHeaders(r *http.Request, headers Headers) *Client {
	for key, val := range client.Headers {
		r.Header.Add(key, val)
	}

	for key, val := range headers {
		r.Header.Add(key, val)
	}

	return client
}

func (client *Client) SendGet(path string, params Params, headers Headers) ([]byte, *int, error) {
	request, err := client.createRequest(http.MethodGet, path, params, nil)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", http.MethodGet).
			Str("url", client.baseUrl+path).
			Msg("failed to build HTTP request")
		return nil, nil, err
	}

	client.fillRequestHeaders(request, headers)

	var response *http.Response

	response, err = client.getResponse(request)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", request.Method).
			Str("url", request.URL.String()).
			Msg("failed to send HTTP request")
		return nil, nil, err
	}

	client.logger.Info().
		Str("method", request.Method).
		Str("url", request.URL.String()).
		Int("status", response.StatusCode).
		Msg("http request succeeded")

	return getResponseBody(response, client.logger)
}

func (client *Client) SendPost(
	path string,
	jsonData []byte,
	queryParams Params,
	headers Headers,
) ([]byte, *int, error) {

	request, err := client.createRequest(http.MethodPost, path, queryParams, jsonData)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", http.MethodPost).
			Str("url", client.baseUrl+path).
			Msg("failed to build HTTP request")
		return nil, nil, err
	}

	client.fillRequestHeaders(request, headers)

	var response *http.Response

	response, err = client.getResponse(request)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", request.Method).
			Str("url", request.URL.String()).
			Msg("failed to send HTTP request")
		return nil, nil, err
	}
	client.logger.Info().
		Str("method", request.Method).
		Str("url", request.URL.String()).
		Int("status", response.StatusCode).
		Msg("http request succeeded")

	return getResponseBody(response, client.logger)
}

func (client *Client) SendPut(
	path string,
	jsonData []byte,
	queryParams Params,
	headers Headers,
) ([]byte, *int, error) {
	request, err := client.createRequest(http.MethodPut, path, queryParams, jsonData)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", http.MethodPut).
			Str("url", client.baseUrl+path).
			Msg("failed to build HTTP request")
		return nil, nil, err
	}

	client.fillRequestHeaders(request, headers)

	response, err := client.getResponse(request)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", request.Method).
			Str("url", request.URL.String()).
			Msg("failed to send HTTP request")
		return nil, nil, err
	}

	client.logger.Info().
		Str("method", request.Method).
		Str("url", request.URL.String()).
		Int("status", response.StatusCode).
		Msg("http request succeeded")

	return getResponseBody(response, client.logger)
}

func (client *Client) SendPatch(
	path string,
	jsonData []byte,
	queryParams Params,
	headers Headers,
) ([]byte, *int, error) {
	request, err := client.createRequest(http.MethodPatch, path, queryParams, jsonData)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", http.MethodPatch).
			Str("url", client.baseUrl+path).
			Msg("failed to build HTTP request")
		return nil, nil, err
	}

	client.fillRequestHeaders(request, headers)

	response, err := client.getResponse(request)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", request.Method).
			Str("url", request.URL.String()).
			Msg("failed to send HTTP request")
		return nil, nil, err
	}

	client.logger.Info().
		Str("method", request.Method).
		Str("url", request.URL.String()).
		Int("status", response.StatusCode).
		Msg("http request succeeded")

	return getResponseBody(response, client.logger)
}

func (client *Client) SendDelete(path string, params Params, headers Headers) ([]byte, *int, error) {
	request, err := client.createRequest(http.MethodDelete, path, params, nil)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", http.MethodDelete).
			Str("url", client.baseUrl+path).
			Msg("failed to build HTTP request")
		return nil, nil, err
	}

	client.fillRequestHeaders(request, headers)

	response, err := client.getResponse(request)
	if err != nil {
		client.logger.Error().
			Err(err).
			Str("method", request.Method).
			Str("url", request.URL.String()).
			Msg("failed to send HTTP request")
		return nil, nil, err
	}

	client.logger.Info().
		Str("method", request.Method).
		Str("url", request.URL.String()).
		Int("status", response.StatusCode).
		Msg("http request succeeded")

	return getResponseBody(response, client.logger)
}

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

func (client *Client) getResponse(request *http.Request) (*http.Response, error) {
	response, err := client.httpClient.Do(request)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func closeResponseBody(response *http.Response) error {
	err := response.Body.Close()

	if err != nil {
		return errors.New("failed to close response body")
	}
	return nil
}

func getResponseBody(response *http.Response, logger *zerolog.Logger) ([]byte, *int, error) {
	defer func() {
		if err := closeResponseBody(response); err != nil {
			logger.Warn().
				Err(err).
				Msg("failed to close response body")
		}
	}()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, &response.StatusCode, err
	}

	return body, &response.StatusCode, nil
}
