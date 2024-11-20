package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminFeedbackRoutes(router fiber.Router) {
	feedbacks := router.Group("/feedbacks",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN}))
	feedbacks.Get("/", controllers.GetFeedbacks)
	// feedbacks.Post("/:id/reasons", middlewares.DeSerializeEmployee,
	// 	middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER}),
	// 	controllers.GivePermission)
}
