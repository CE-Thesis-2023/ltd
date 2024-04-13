package hikvision

import (
	"context"
	"net/http"
	"net/url"
	"time"

	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
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
}

func NewClient(options ...HikvisionClientOptioner) (Client, error) {
	opts := hikvisionOptions{}
	for _, o := range options {
		o(&opts)
	}
	return &client{
		options: &opts,
	}, nil
}

type Credentials struct {
	WebSessionId     string `json:"webSessionId"`
	WebSessionCookie string `json:"WebSessionCookie"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Ip               string `json:"ip"`
}

func (c *client) getRestClient(opts *Credentials) (*http.Client, string) {
	ip := opts.Ip

	if opts.Username != "" && opts.Password != "" {
		u, err := url.Parse(ip)
		if err != nil {
			logger.SError("failed to parse camera IP", zap.Error(err))
			return nil, ""
		}
		u = u.JoinPath("/ISAPI")
		u.Scheme = "http"
		u.User = url.UserPassword(opts.Username, opts.Password)
		ip = u.String()
	}

	httpClient := custhttp.NewHttpClient(
		context.Background(),
		custhttp.WithTimeout(2*time.Second))

	return httpClient, ip
}

func (c *client) PtzCtrl(credentials *Credentials) PtzApiClientInterface {
	client, ip := c.getRestClient(credentials)
	return &ptzApiClient{
		httpClient: client,
		ip:         ip,
		username:   credentials.Username,
		password:   credentials.Password,
	}
}

func (c *client) Smart(credentials *Credentials) SmartApiInterface {
	client, ip := c.getRestClient(credentials)
	return &smartApiClient{
		httpClient: client,
		ip:         ip,
		username:   credentials.Username,
		password:   credentials.Password,
	}
}

func (c *client) Event(credentials *Credentials) EventApiInterface {
	client, ip := c.getRestClient(credentials)
	return &eventApiClient{
		httpClient: client,
		ip:         ip,
		username:   credentials.Username,
		password:   credentials.Password,
	}
}

func (c *client) System(credentials *Credentials) SystemApiInterface {
	client, ip := c.getRestClient(credentials)
	return &systemApiClient{
		httpClient: client,
		ip:         ip,
		username:   credentials.Username,
		password:   credentials.Password,
	}
}

func (c *client) Streams(credentials *Credentials) StreamsApiInterface {
	client, ip := c.getRestClient(credentials)
	return &streamApiClient{
		ip:         ip,
		httpClient: client,
		username:   credentials.Username,
		password:   credentials.Password,
	}
}
