package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupV1AuctionRoutes(router fiber.Router) {
	auctions := router.Group("/auctions")
	auctions.Get("/",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controllers.GetAuctions)
	auctions.Get("/:id",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controllers.GetAuctionDetail)
	auctions.Post("/:id/bid",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controllers.BidAuction)
}
