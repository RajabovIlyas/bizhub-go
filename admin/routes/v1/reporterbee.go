package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/admin/controllers/v1"
	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupAdminReporterBeeRoutes(router fiber.Router) {
	reporterbee := router.Group("/reporterbee",
		middlewares.DeSerializeEmployee,
		middlewares.AllowRoles([]string{config.ADMIN}),
	)

	reporterbee.Get("/", controllers.GetReporterBee) // logo, name, objectID
	reporterbee.Put("/", controllers.EditReporterBee)
	reporterbee.Get("/posts", controllers.GetReporterBeePosts)
	reporterbee.Post("/posts", controllers.CreateNewPost)
	reporterbee.Get("/posts/:id", controllers.GetPostById)
}
