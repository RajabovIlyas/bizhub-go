package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminAttributesRoutes(router fiber.Router) {
	attributes := router.Group("/attributes",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN_CHECKER}),
	)
	attributes.Get("/", controllers.GetAttributes)
	attributes.Post("/", controllers.AddNewAttribute)
	attributes.Get("/:id", controllers.GetAttribute)
	attributes.Put("/:id", controllers.EditAttribute)
}
