package helpers

import (
	"fmt"

	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
)

type ResponseFunc func(fn string, err error, code string) models.Response[any]

func ErrorResponse(from string) ResponseFunc {
	return func(fn string, err error, code string) models.Response[any] {
		_code := code
		if len(code) == 0 {
			_code = "message"
		}
		mes := fmt.Sprintf("[%v]:{ %v }:%v", from, fn, err.Error())

		return models.Response[any]{
			IsSuccess: false,
			Result:    nil,
			Error: fiber.Map{
				"message": mes,
				"code":    _code,
			},
		}
	}
}
