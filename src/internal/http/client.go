package custhttp

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/CE-Thesis-2023/ltd/src/internal/logger"

	"encoding/json"

	"go.uber.org/zap"
)

type Options struct {
	timeout time.Duration
}

type ClientOptioner func(o *Options)

func WithTimeout(dur time.Duration) ClientOptioner {
	return func(o *Options) {
		o.timeout = dur
	}
}

func NewHttpClient(ctx context.Context, opts ...ClientOptioner) *http.Client {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	client := &http.Client{
		Timeout: options.timeout,
	}
	return client
}

type HttpRequestOptions struct {
	headers  map[string]string
	body     []byte
	username string
	password string
}

func (o *HttpRequestOptions) hasBasicAuth() bool {
	return o.username != "" && o.password != ""
}

type HttpRequestOptioner func(o *HttpRequestOptions)

func WithHeader(key string, value string) HttpRequestOptioner {
	return func(o *HttpRequestOptions) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers[key] = value
	}
}

func WithXMLBody(body interface{}) HttpRequestOptioner {
	return func(o *HttpRequestOptions) {
		bodyBytes, err := xml.Marshal(body)
		if err != nil {
			logger.SDebug("failed to marshal XML body",
				zap.Error(err))
			return
		}
		o.body = bodyBytes
	}
}

func WithJSONBody(body interface{}) HttpRequestOptioner {
	return func(o *HttpRequestOptions) {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			logger.SDebug("failed to marshal JSON body",
				zap.Error(err))
			return
		}
		o.body = bodyBytes
	}
}

func WithBasicAuth(username, password string) HttpRequestOptioner {
	return func(o *HttpRequestOptions) {
		o.username = username
		o.password = password
	}
}

func WithContentType(contentType string) HttpRequestOptioner {
	return func(o *HttpRequestOptions) {
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers["Content-Type"] = contentType
	}
}

func NewHttpRequest(ctx context.Context, url *url.URL, method string, options ...HttpRequestOptioner) (*http.Request, error) {
	reqOptions := &HttpRequestOptions{}
	for _, o := range options {
		o(reqOptions)
	}

	var bodyReader io.Reader
	if reqOptions.body != nil {
		bodyReader = bytes.NewReader(reqOptions.body)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		url.String(),
		bodyReader)
	if err != nil {
		logger.SDebug("unable to create new http request",
			zap.Error(err))
		return nil, err
	}

	if reqOptions.hasBasicAuth() {
		req.SetBasicAuth(reqOptions.username, reqOptions.password)
	}
	for key, value := range reqOptions.headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func JSONResponse(resp *http.Response, dest interface{}) error {
	body := resp.Body
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		logger.SDebug("failed to read HTTP response body",
			zap.Error(err))
		return err
	}

	if err := json.Unmarshal(bodyBytes, dest); err != nil {
		logger.SDebug("failed to unmarshal JSON response",
			zap.Error(err))
		return err
	}

	return nil
}

func XMLResponse(resp *http.Response, dest interface{}) error {
	body := resp.Body
	defer body.Close()
	
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		logger.SDebug("failed to read HTTP response body", zap.Error(err))
		return err
	}
	fmt.Println(string(bodyBytes))

	if err := xml.Unmarshal(bodyBytes, dest); err != nil {
		logger.SDebug("failed to unmarshal XML response", zap.Error(err))
		return err
	}

	return nil
}

