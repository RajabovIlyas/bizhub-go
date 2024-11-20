package v1

import (
	mobileController "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/gofiber/fiber/v2"
)

func SetupV1CategoryRoutes(router fiber.Router) {
	categories := router.Group("categories")
	// categories.Get("/", v1.GetCategoryParents)
	categories.Get("/", mobileController.GetCategories)
	categories.Get("/:id", mobileController.GetCategoryChildren)
	categories.Get("/:id/attributes", mobileController.GetCategoryAttributes)
}
