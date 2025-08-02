package middleware

import (
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	fm "github.com/oapi-codegen/fiber-middleware"
	"net/http"
	"roflbeacon2/app/api"
)

func init() {
	openapi3filter.RegisterBodyDecoder("image/jpeg", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("image/jpg", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("image/png", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("image/bmp", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("image/gif", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("image/webp", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("image/webm", openapi3filter.FileBodyDecoder)
}

func NewOpenAPIValidator() fiber.Handler {
	spec, err := api.GetSwagger()
	if err != nil {
		log.Fatalf("Failed to get swagger spec: %v", err)
	}

	return fm.OapiRequestValidatorWithOptions(spec,
		&fm.Options{
			Options: openapi3filter.Options{},
			ErrorHandler: func(c *fiber.Ctx, message string, _ int) {
				c.Status(fiber.StatusForbidden).JSON(api.General{ //nolint:errcheck
					Error:      true,
					Msg:        message,
					StatusCode: http.StatusForbidden,
				})
			},
		})
}
