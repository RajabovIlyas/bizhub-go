package helpers

import (
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
)

func GetCultureFromQuery(c *fiber.Ctx) models.Culture {
	_c := c.Query("culture", "tm")
	switch _c {
	case "tm":
		break
	case "tr":
		break
	case "en":
		break
	case "ru":
		break
	default:
		_c = "tm"
		break
	}
	return models.Culture{
		Lang: _c,
	}
}
