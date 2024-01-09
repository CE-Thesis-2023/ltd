package publicapi

import (
	"github.com/CE-Thesis-2023/ltd/src/biz/service"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"
	"github.com/CE-Thesis-2023/ltd/src/models/ms"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func POSTMSAdmissionWebhook(ctx *fiber.Ctx) error {
	var req *ms.AdmissionWebhookRequest
	if err := ctx.BodyParser(req); err != nil {
		logger.SError("POSTMSAdmissionWebhook: parse request error",
			zap.Error(err))
		return err
	}

	resp, err := service.GetStreamManagementService().
		MediaService().
		AdmissionWebhook(ctx.Context(), req)
	if err != nil {
		return err
	}

	logger.SInfo("POSTMSAdmissionWebhook",
		logger.Json("response", resp))
	return ctx.JSON(resp)
}

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
