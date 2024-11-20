package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupV1FeedbackRoutes(router fiber.Router) {
	feeds := router.Group("/feedbacks")
	feeds.Post("/", middlewares.DeSerializeCustomer, controllers.CreateFeedback)
}
