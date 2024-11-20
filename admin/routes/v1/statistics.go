package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminStatisticRoutes(router fiber.Router) {
	statistics := router.Group("/statistics",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN, config.OWNER}))
	statistics.Get("/users", controllers.UsersActivity)
	statistics.Get("/employees", controllers.EmployeesActivity)
	statistics.Get("/users/detail", controllers.UsersActivityDetail)
	statistics.Get("/sellers", controllers.SellersActivity)
	statistics.Get("/sellers/detail", controllers.SellersActivityDetail)
	statistics.Get("/published_items", controllers.PublishedProductsAndPosts)
	statistics.Get("/published_items/detail", controllers.PublishedProductsAndPostsDetail)
	statistics.Get("/money", controllers.MoneyActivity)
	statistics.Get("/money/detail", controllers.MoneyActivityDetail)
	statistics.Get("/expenses", controllers.ExpensesActivity)
	statistics.Post("/run", controllers.RunStatistic)
}
