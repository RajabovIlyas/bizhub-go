package v1

import (
	adminController "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	mobileController "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/gofiber/fiber/v2"
)

func SetupV1BrandRoutes(router fiber.Router) {
	brands := router.Group("/brands")
	brands.Get("/", adminController.GetBrands) // TODO: suny mobileController-e gecirmeli!
	brands.Get("/parents", mobileController.GetBrandParents)
	brands.Get("/:id", mobileController.GetBrandChildren)
}
