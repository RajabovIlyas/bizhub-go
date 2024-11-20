package v1

import "github.com/gofiber/fiber/v2"

func SetupRoutes(api fiber.Router) {
	v1 := api.Group("v1")
	// v1.Get("/", func(c *fiber.Ctx) error {
	// 	return c.SendString("welcome to v1 api of bizhub admin panel!")
	// })
	SetupAdminAuthRoutes(v1)
	SetupAdminEmployeeRoutes(v1)
	SetupAdminStatisticRoutes(v1)
	SetupAdminCashierRoutes(v1)
	SetupAdminSellerRoutes(v1)
	SetupAdminNotificationRoutes(v1)
	SetupAdminFeedbackRoutes(v1)
	SetupAdminBannerRoutes(v1)
	SetupAdminReporterBeeRoutes(v1)
	SetupAdminAuctionRoutes(v1)
	SetupAdminTaskRoutes(v1)
	SetupAdminCategoriesRoutes(v1)
	SetupAdminAttributesRoutes(v1)
	SetupAdminBrandsRoutes(v1)
}
