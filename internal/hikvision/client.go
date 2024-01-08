package hikvision

import (
	"fmt"
	"labs/local-transcoder/internal/logger"
	"net/http"
	"net/url"
	"time"

	fastshot "github.com/opus-domini/fast-shot"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

type Client interface {
	PtzCtrl(credentials *Credentials) PtzApiClientInterface
	Smart(credentials *Credentials) SmartApiInterface
	Event(credentials *Credentials) EventApiInterface
	System(credentials *Credentials) SystemApiInterface
	Streams(credentials *Credentials) StreamsApiInterface
}

type client struct {
	options *hikvisionOptions
	pool    *ants.Pool
}

func NewClient(options ...HikvisionClientOptioner) (Client, error) {
	opts := hikvisionOptions{}
	for _, o := range options {
		o(&opts)
	}

	var pool *ants.Pool
	if opts.Poolsize != 0 {
		pool, _ = ants.NewPool(opts.Poolsize,
			ants.WithLogger(logger.NewZapToAntsLogger(logger.Logger())))
	} else {
		pool, _ = ants.NewPool(20,
			ants.WithLogger(logger.NewZapToAntsLogger(logger.Logger())))
	}

	return &client{
		options: &opts,
		pool:    pool,
	}, nil
}

type Credentials struct {
	WebSessionId     string `json:"webSessionId"`
	WebSessionCookie string `json:"WebSessionCookie"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Ip               string `json:"ip"`
}

func (c *client) getRestClient(opts *Credentials) fastshot.ClientHttpMethods {
	ip := opts.Ip

	if opts.Username != "" && opts.Password != "" {
		u, err := url.Parse(ip)
		if err != nil {
			logger.SError("hikvision.NewClient: parse IP error", zap.Error(err))
			return nil
		}
		u.Scheme = "http"
		u.User = url.UserPassword(opts.Username, opts.Password)
		ip = u.String()
	}

	restClientBuilder := fastshot.NewClient(fmt.Sprintf("%s/ISAPI", ip))

	if opts.WebSessionCookie != "" && opts.WebSessionId != "" {
		restClientBuilder.
			Cookie().Add(&http.Cookie{
			Name:  fmt.Sprintf("WebSession_%s", opts.WebSessionId),
			Value: fmt.Sprintf("WebSession_%s", opts.WebSessionCookie),
		})
	}

	restClient := restClientBuilder.
		Config().SetFollowRedirects(true).
		Config().SetTimeout(2 * time.Second).
		Header().AddContentType("application/xml").
		Build()

	return restClient
}

func (c *client) PtzCtrl(credentials *Credentials) PtzApiClientInterface {
	return &ptzApiClient{
		restClient: c.getRestClient(credentials),
		pool:       c.pool,
	}
}

func (c *client) Smart(credentials *Credentials) SmartApiInterface {
	return &smartApiClient{
		restClient: c.getRestClient(credentials),
		pool:       c.pool,
	}
}

func (c *client) Event(credentials *Credentials) EventApiInterface {
	return &eventApiClient{
		restClient: c.getRestClient(credentials),
		pool:       c.pool,
	}
}

func (c *client) System(credentials *Credentials) SystemApiInterface {
	return &systemApiClient{
		restClient: c.getRestClient(credentials),
		pool:       c.pool,
	}
}

func (c *client) Streams(credentials *Credentials) StreamsApiInterface {
	return &streamApiClient{
		restClient: c.getRestClient(credentials),
		pool:       c.pool,
	}
}
