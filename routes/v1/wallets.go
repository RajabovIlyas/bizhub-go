package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupV1WalletRoutes(router fiber.Router) {
	wallet := router.Group("/wallet")
	wallet.Get("/",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controllers.GetMyWallet)
	wallet.Post("/withdraw",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controllers.Withdraw)
	wallet.Post("/pay",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controllers.PayForPackage)
	wallet.Post("/withdraw/:id/cancel",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controllers.CancelWithdraw)
	wallet.Get("/history",
		middlewares.DeSerializeCustomer,
		middlewares.AllowSeller(),
		controllers.GetWalletHistory)
}
