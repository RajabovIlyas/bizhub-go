package v1

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(api fiber.Router) {
	v1 := api.Group("v1")
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("welcome to v1 api of bizhub")
	})
	SetupV1AuthRoutes(v1)
	SetupV1PostRoutes(v1)
	SetupV1SeederRoutes(v1)
	SetupV1CollectionRoutes(v1)
	SetupV1ProductsRoutes(v1)
	SetupV1SellerRoutes(v1)
	SetupV1CategoryRoutes(v1)
	SetupV1BrandRoutes(v1)
	SetupV1BannerRoutes(v1)
	SetupV1CityRoutes(v1)
	SetupV1FeedbackRoutes(v1)
	SetupV1SuggestionRoutes(v1)
	SetupV1PackageRoutes(v1)
	SetupV1AuctionRoutes(v1)
	SetupV1WalletRoutes(v1)
	SetupV1ChatRoutes(v1)
}
