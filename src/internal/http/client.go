package custhttp

import (
	"context"
	"encoding/xml"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"io"
	"time"

	"encoding/json"
	"github.com/opus-domini/fast-shot"
	"go.uber.org/zap"
)

type Options struct {
	baseUrl string
	timeout time.Duration
	header  map[string]string
}

type ClientOptioner func(o *Options)

func WithBaseUrl(url string) ClientOptioner {
	return func(o *Options) {
		o.baseUrl = url
	}
}

func WithTimeout(dur time.Duration) ClientOptioner {
	return func(o *Options) {
		o.timeout = dur
	}
}

func WithHeader(key string, value string) ClientOptioner {
	return func(o *Options) {
		if o.header == nil {
			o.header = map[string]string{}
		}
		o.header[key] = value
	}
}

func NewHttpClient(ctx context.Context, opts ...ClientOptioner) fastshot.ClientHttpMethods {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	fs := fastshot.NewClient(options.baseUrl).
		Config().SetFollowRedirects(true).
		Config().SetTimeout(options.timeout)

	fs.Header().AddAll(options.header)

	return fs.Build()
}

func JSONResponse(resp *fastshot.Response, dest interface{}) error {
	body := resp.RawResponse.Body
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		logger.SDebug("ParseResponseBody: io.ReadAll",
			zap.Error(err))
		return err
	}

	if err := json.Unmarshal(bodyBytes, dest); err != nil {
		logger.SDebug("ParseResponseBody: json.Unmarshal",
			zap.Error(err))
		return err
	}

	return nil
}

func XMLResponse(resp *fastshot.Response, dest interface{}) error {
	body := resp.RawResponse.Body
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		logger.SDebug("ParseResponseBody: io.ReadAll", zap.Error(err))
		return err
	}

	logger.SDebug("ParseResponseBody: xml response", zap.String("data", string(bodyBytes)))

	if err := xml.Unmarshal(bodyBytes, dest); err != nil {
		logger.SDebug("ParseResponseBody: xml.Unmarshal", zap.Error(err))
		return err
	}

	return nil
}
