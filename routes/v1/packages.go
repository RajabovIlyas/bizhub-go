package v1

import (
	controller "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupV1PackageRoutes(router fiber.Router) {
	packages := router.Group("packages")
	packages.Get("/:packageType",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controller.GetPackage)
	packages.Get("/",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controller.GetPackages)
}
