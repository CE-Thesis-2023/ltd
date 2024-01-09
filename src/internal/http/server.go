package custhttp

import (
	"context"
	"errors"
	"fmt"
	"github.com/CE-Thesis-2023/ltd/internal/configs"
	custerror "github.com/CE-Thesis-2023/ltd/internal/error"

	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type RegistrationFunc func(app *fiber.App)

type HttpServer struct {
	configs *HttpServerConfigs
	app     *fiber.App
}

func New(options ...Optioner) *HttpServer {
	configs := &HttpServerConfigs{}
	for _, option := range options {
		option(configs)
	}

	globalConfigs := configs.configs
	httpConfigs := fiber.Config{
		Network:               "tcp",
		AppName:               globalConfigs.Name,
		ErrorHandler:          configs.errorHandler,
		DisableStartupMessage: true,
	}

	if configs.templatePath != "" {
		engine := html.New(configs.templatePath, ".html")
		httpConfigs.Views = engine
	}

	app := fiber.New(httpConfigs)

	configs.registration(app)
	if len(configs.middlewares) > 0 {
		app.Use(configs.middlewares...)
	}

	return &HttpServer{
		app:     app,
		configs: configs,
	}
}

type HttpServerConfigs struct {
	configs      *configs.HttpConfigs
	registration RegistrationFunc
	errorHandler fiber.ErrorHandler
	middlewares  []interface{}
	templatePath string
}

type Optioner func(configs *HttpServerConfigs)

func WithErrorHandler(handler fiber.ErrorHandler) Optioner {
	return func(configs *HttpServerConfigs) {
		configs.errorHandler = handler
	}
}

func WithGlobalConfigs(conf *configs.HttpConfigs) Optioner {
	return func(configs *HttpServerConfigs) {
		configs.configs = conf
	}
}

func WithRegistration(handler RegistrationFunc) Optioner {
	return func(configs *HttpServerConfigs) {
		configs.registration = handler
	}
}

func WithMiddleware(middlewares ...interface{}) Optioner {
	return func(configs *HttpServerConfigs) {
		configs.middlewares = middlewares
	}
}

func WithTemplatePath(path string) Optioner {
	return func(configs *HttpServerConfigs) {
		configs.templatePath = path
	}
}

func (s *HttpServer) Start() error {
	globalConfigs := s.configs.configs
	tlsConfigs := globalConfigs.Tls
	port := fmt.Sprintf(":%d", globalConfigs.Port)

	if tlsConfigs.Cert == "" || tlsConfigs.Key == "" {
		if err := s.app.Listen(port); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return custerror.FormatInternalError("HttpServer.Start: err = %s", err)
		}
	} else {
		if err := s.app.ListenTLS(port, tlsConfigs.Cert, tlsConfigs.Key); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return custerror.FormatInternalError("HttpServer.Start: err = %s", err)
		}

	}
	return nil
}

func (s *HttpServer) Name() string {
	return s.configs.configs.Name
}

func (s *HttpServer) Stop(ctx context.Context) error {
	globalConfigs := s.configs.configs

	if err := s.app.ShutdownWithContext(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return custerror.FormatTimeout("HttpServer.Stop: server stopping deadline exceeded name = %s", globalConfigs.Name)
		}
		return custerror.FormatInternalError("HttpServer.Shutdown: err = %s", err)
	}

	return nil
}
