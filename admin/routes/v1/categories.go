package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminCategoriesRoutes(router fiber.Router) {
	categories := router.Group("/categories",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN_CHECKER}),
	)
	categories.Get("/", controllers.GetCategories)
	categories.Post("/", controllers.AddNewCategory)
	categories.Get("/:id", controllers.GetCategory)
	categories.Put("/:id", controllers.EditCategory)
}
