package v1

import (
	controller "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/gofiber/fiber/v2"
)

func SetupV1BannerRoutes(router fiber.Router) {
	banners := router.Group("/banners")
	banners.Get("/", controller.GetBanners)
	// banners.Get("/:id", controller.GetBrandChildren)
}
