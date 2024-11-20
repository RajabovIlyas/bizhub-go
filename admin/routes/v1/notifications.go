package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminNotificationRoutes(router fiber.Router) {
	notifications := router.Group("/notifications")
	// notifications.Get("/", middlewares.DeSerializeEmployee,
	// 	middlewares.AllowRoles([]string{config.ADMIN}),
	// 	controllers.GetNotifications)
	notifications.Post("/",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.NewNotification)
	// notifications.Get("/:id",
	// // middlewares.DeSerializeEmployee,
	// middlewares.AllowRoles([]string{config.ADMIN}),
	// controllers.GetNotification)
	// notifications.Put("/:id",
	// 	middlewares.DeSerializeEmployee,
	// 	middlewares.AllowRoles([]string{config.ADMIN}),
	// 	controllers.CheckNotification)
	// notifications.Post("/:id/reasons", middlewares.DeSerializeEmployee,
	// 	middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER}),
	// 	controllers.GivePermission)
}
