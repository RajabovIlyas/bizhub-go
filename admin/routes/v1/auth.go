package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminAuthRoutes(router fiber.Router) {
	auth := router.Group("/auth")
	auth.Post("/login", controllers.Login)
	auth.Post("/refresh", controllers.RefreshAccessToken)
	// auth.Get("/cookie", controllers.DummyCookie) // cookie iberip gorduk,
	auth.Post("/start_working", middlewares.DeSerializeEmployee,
		middlewares.AllowRoles(config.ALL_EMPLOYEES),
		controllers.StartWorking)
	auth.Post("/stop_working", middlewares.DeSerializeEmployee,
		middlewares.AllowRoles(config.ALL_EMPLOYEES),
		controllers.StopWorking)
	auth.Put("/change_password", middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN, config.EMPLOYEES_MANAGER,
			config.OWNER, config.ADMIN_CHECKER}),
		controllers.ChangePassword)
}
