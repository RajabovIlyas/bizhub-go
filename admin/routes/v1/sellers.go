package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminSellerRoutes(router fiber.Router) {
	sellers := router.Group("/sellers",
		middlewares.DeSerializeEmployee,
	)
	sellers.Get("/find",
		middlewares.AllowRoles([]string{config.CASHIER}),
		controllers.FindSeller)
	sellers.Get("/",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetAllSellers)
	sellers.Get("/:id",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetSellerProfile)
	sellers.Get("/:id/transfers",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetSellerTransfers)
	sellers.Get("/:id/package_history",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetSellerPackageHistory)
	sellers.Post("/:id/extend_package",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.ExtendSellerPackage)
	sellers.Post("/:id/reduce_package",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.ReduceSellerPackage)
	sellers.Post("/:id/promote",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.PromoteSellerType)
	sellers.Post("/:id/demote",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.DemoteSellerType)
	sellers.Post("/:id/block",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.BlockSeller)
	sellers.Post("/:id/unblock",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.UnBlockSeller)
	sellers.Get("/:id/products",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetSellerProducts)
	sellers.Get("/:id/products/:productId",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetSellerProductDetails)
	sellers.Get("/:id/posts",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetSellerPosts)
	sellers.Get("/:id/posts/:postId",
		middlewares.AllowRoles([]string{config.ADMIN}),
		controllers.GetSellerPostDetails)
	// sellers.Post("/:id/reasons", middlewares.DeSerializeEmployee,
	// 	middlewares.AllowRoles([]string{config.EMPLOYEES_MANAGER}),
	// 	controllers.GivePermission)
}
