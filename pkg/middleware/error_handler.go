// nolint: wrapcheck
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/samber/oops"
	"log/slog"
	"net/http"
	"roflbeacon2/app/api"
)

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	statusCode := http.StatusInternalServerError

	if oopsErr, ok := oops.AsOops(err); ok {
		statusCodeOpt := oopsErr.Context()["statusCode"]
		if statusCodeOpt != nil {
			statusCode, _ = statusCodeOpt.(int)
		}
	}

	general := api.General{
		Error:      true,
		Msg:        err.Error(),
		StatusCode: statusCode,
	}

	switch statusCode {
	case http.StatusInternalServerError:
		slog.LogAttrs(ctx.UserContext(), slog.LevelError, "Internal Server Error", slog.Any("error", err))
	case http.StatusBadRequest:
		slog.LogAttrs(ctx.UserContext(), slog.LevelError, "Bad Request", slog.Any("error", err))
	case http.StatusForbidden:
		slog.LogAttrs(ctx.UserContext(), slog.LevelError, "Forbidden", slog.Any("error", err))
	}

	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(general.StatusCode)
	return ctx.JSON(general)
}
