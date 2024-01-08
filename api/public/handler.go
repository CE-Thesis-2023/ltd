package publicapi

import (
	"labs/local-transcoder/biz/service"
	"labs/local-transcoder/internal/logger"
	"labs/local-transcoder/models/ms"

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
