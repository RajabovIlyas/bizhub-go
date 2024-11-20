package v1

import (
	v1 "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupV1ProductsRoutes(router fiber.Router) {
	products := router.Group("products")
	products.Get("/search", v1.SearchProduct)
	products.Get("/filter", v1.FilterProduct)
	products.Get("/:id", v1.GetProductDetail)
	products.Post("/:id/discount",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		v1.SetProductDiscount)
	products.Delete("/:id/discount",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		v1.RemoveProductDiscount)
}
