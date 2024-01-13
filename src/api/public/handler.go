package publicapi

import (
	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"

	"github.com/gofiber/fiber/v2"
)

func GETDebugListStreams(ctx *fiber.Ctx) error {
	resp, err := service.
		GetCommandService().
		DebugListStreams(ctx.Context())
	if err != nil {
		return err
	}
	logger.SDebug("GETDebugListStreams", logger.Json("response", resp))
	return ctx.JSON(resp)
}

func GETHealthcheck(ctx *fiber.Ctx) error {
	return ctx.SendStatus(200)
}
