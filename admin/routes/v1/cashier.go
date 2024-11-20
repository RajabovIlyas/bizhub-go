package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminCashierRoutes(router fiber.Router) {
	cashier := router.Group("/cashier",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.CASHIER}),
	)
	cashier.Post("/withdraw", controllers.Withdraw)
	cashier.Post("/code", controllers.Code)
	cashier.Post("/deposit", controllers.Deposit)
	cashier.Get("/find", controllers.FindSeller)
	cashier.Get("/completed_tasks", controllers.CompletedTasks)
}
