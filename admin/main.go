package admin

import (
	v1 "github.com/devzatruk/bizhubBackend/admin/routes/v1"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminRoutes(app *fiber.App) {
	be := app.Group("be")
	v1.SetupRoutes(be)
}
