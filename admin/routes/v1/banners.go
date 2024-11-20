package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminBannerRoutes(router fiber.Router) {
	banners := router.Group("/banners",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN}))
	banners.Get("/", controllers.GetBanners)
	banners.Put("/:id", controllers.EditBanner)
}
