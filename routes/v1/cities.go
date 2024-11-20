package v1

import (
	v1 "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/gofiber/fiber/v2"
)

func SetupV1CityRoutes(router fiber.Router) {
	cities := router.Group("cities")
	cities.Get("/", v1.GetCities)
	// cities.Get("/:id", v1.GetBrandChildren)
}
