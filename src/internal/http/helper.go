package custhttp

import (
	"database/sql"
	"labs/local-transcoder/internal/configs"
	custerror "labs/local-transcoder/internal/error"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/panjf2000/ants/v2"
)

func CommonPublicMiddlewares(configs *configs.HttpConfigs) []interface{} {
	return []interface{}{
		limiter.New(limiter.Config{
			Max:        10,
			Expiration: 1 * time.Second,
			LimitReached: func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusTooManyRequests)
			},
		}),
		cors.New(cors.Config{
			AllowOriginsFunc: func(origin string) bool {
				switch origin {
				case "http://*":
				case "https://*":
				default:
					return false
				}
				return true
			},
			AllowMethods: "GET,HEAD",
			AllowHeaders: "Origin, Content-Type, Accept-Encoding, Host",
		}),
		etag.New(),
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),
		recover.New(recover.ConfigDefault),
		logger.New(logger.Config{
			DisableColors: true,
			Format:        "PUBLIC ${time} [${ip}:${port}] ${latency} ${method} ${status} ${path}\n",
			TimeFormat:    time.RFC3339,
		}),
		cache.New(cache.Config{
			Expiration:   time.Minute * 1,
			CacheControl: false,
			CacheHeader:  "X-Cache",
			Methods: []string{
				fiber.MethodGet,
				fiber.MethodHead,
			},
		}),
		helmet.New(helmet.ConfigDefault),
	}
}

func CommonPrivateMiddlewares(configs *configs.HttpConfigs) []interface{} {
	username := configs.Auth.Username
	token := configs.Auth.Token

	return []interface{}{
		limiter.New(limiter.Config{
			Max:        5,
			Expiration: 1 * time.Second,
			LimitReached: func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusTooManyRequests)
			},
		}),
		cors.New(cors.Config{
			AllowOriginsFunc: func(origin string) bool {
				switch origin {
				case "http://*":
				case "https://*":
				default:
					return false
				}
				return true
			},
			AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
			AllowHeaders: "Origin, Content-Type, Accept-Encoding, Host, Authorization",
		}),
		etag.New(),
		compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}),
		basicauth.New(basicauth.Config{
			Users: map[string]string{
				username: token,
			},
			ContextUsername: "username",
			ContextPassword: "token",
			Unauthorized: func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusUnauthorized)
			},
		}),
		recover.New(recover.ConfigDefault),
		logger.New(logger.Config{
			DisableColors: true,
			Format:        "ADMIN ${time} [${ip}:${port}] (${latency}) [${locals:username}:${locals:token}] ${method} ${status}  ${path}\n",
			TimeFormat:    time.RFC3339,
		}),
	}
}

func GlobalErrorHandler() func(c *fiber.Ctx, err error) error {
	return func(c *fiber.Ctx, err error) error {
		customError, yes := err.(*custerror.CustomError)
		if !yes {
			switch err {
			case sql.ErrNoRows:
				customError = custerror.ErrorNotFound
			case sql.ErrTxDone, sql.ErrConnDone:
				customError = custerror.ErrorInternal
			case ants.ErrInvalidLoadBalancingStrategy,
				ants.ErrInvalidPoolExpiry,
				ants.ErrInvalidPoolIndex,
				ants.ErrInvalidPreAllocSize,
				ants.ErrLackPoolFunc,
				ants.ErrPoolClosed,
				ants.ErrPoolOverload:
				customError = custerror.ErrorInternal
			case mqtt.ErrConnectionClosed,
				mqtt.ErrListenerIDExists,
				mqtt.ErrMinimumKeepalive,
				mqtt.ErrConnectionClosed,
				mqtt.ErrInlineClientNotEnabled,
				mqtt.ErrInvalidConfigType,
				mqtt.ErrConnectionClosed:
				customError = custerror.ErrorInternal
			default:
				customError = custerror.ErrorInternal
			}
		}
		return customError.
			Parent().
			Fiber(c)
	}
}
