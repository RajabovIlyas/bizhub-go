package v1

import (
	v1 "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/gofiber/fiber/v2"
)

func SetupV1CollectionRoutes(router fiber.Router) {
	collections := router.Group("collections")
	collections.Get("/", v1.GetCollectionsInfo)
	collections.Get("/new", v1.GetCollectionNewAll)
	collections.Get("/discounted", v1.GetCollectionDiscountedAll)
	collections.Get("/trending", v1.GetCollectionTrendingAll)
	quick := collections.Group("quick") // iki product gorkezyanler
	quick.Get("/new", v1.GetCollectionNewBrief)
	quick.Get("/discounted", v1.GetCollectionDiscountedBrief)
	quick.Get("/trending", v1.GetCollectionTrendingBrief)
}
