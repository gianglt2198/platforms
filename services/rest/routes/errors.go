package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type APIError struct {
	Status  int
	Message interface{}
}

func FromError(ctx *fiber.Ctx, err error) error {
	var apiError APIError
	var svcError *fiber.Error
	if errors.As(err, &svcError) {
		apiError.Message = ErrorResponse(svcError)
		apiError.Status = svcError.Code
	}

	return ctx.Status(apiError.Status).JSON(apiError.Message)
}
