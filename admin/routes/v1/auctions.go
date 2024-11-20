package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminAuctionRoutes(router fiber.Router) {
	auctions := router.Group("/auctions",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN}),
	)
	auctions.Post("/", controllers.CreateNewAuction)
	auctions.Get("/", controllers.GetAuctions)
	auctions.Get("/:id", controllers.GetAuctionDetail)
}
