package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"

	"github.com/gofiber/fiber/v2"
)

func SetupAdminBrandsRoutes(router fiber.Router) {
	brands := router.Group("/brands",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN_CHECKER}),
	)
	brands.Get("/", controllers.GetBrands)
	brands.Post("/", controllers.AddNewBrand)
	brands.Get("/:id", controllers.GetBrand)
	brands.Put("/:id", controllers.EditBrand)
	brands.Get("/parents", controllers.GetParentBrands)
}
