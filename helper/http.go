package helper

import (
	"github.com/gofiber/fiber/v2"
)

func GetPageAndLimitFromCtx(ctx *fiber.Ctx) (page uint64, limit uint64) {
	pageQuery := ctx.QueryInt("page")
	limitQuery := ctx.QueryInt("limit")

	if pageQuery >= 0 {
		page = uint64(pageQuery)
	}

	if limitQuery >= 0 {
		limit = uint64(limitQuery)
	}

	return page, limit
}
