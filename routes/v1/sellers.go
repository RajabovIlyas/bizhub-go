package v1

import (
	v1 "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/gofiber/fiber/v2"
)

func SetupV1SellerRoutes(router fiber.Router) {
	sellers := router.Group("sellers")
	sellers.Get("/", v1.GetAllSellers)
	sellers.Get("/top", v1.GetTopSellers)
	sellers.Get("/search", v1.SearchSellers)
	sellers.Get("/filter", v1.FilterSellers)
	sellers.Get("/filter/aggregations", v1.GetSellersFilterAggregations)
	sellers.Get("/:id", v1.GetProfileOfAnySeller)
	sellers.Get("/:id/products", v1.GetProductsBySellerId)
	sellers.Get("/:id/posts", v1.GetPostBySellerId)
	sellers.Get("/:id/categories", v1.GetCategoriesBySellerId)
	// sellers.Get("/deleteAll", v1.DeleteAllsellers)
	// sellers.Get("/:id", v1.GetSellerDetails)
}
