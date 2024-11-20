package v1

import (
	controllers "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/devzatruk/bizhubBackend/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupV1SuggestionRoutes(router fiber.Router) {
	suggestions := router.Group("/suggestions")
	// suggestions.Get("/", middlewares.DeSerializeCustomer, controllers.GetFeedbacks)
	suggestions.Get("/", controllers.GetSuggestions)
	suggestions.Post("/", middlewares.DeSerializeCustomer, controllers.AddSuggestion)
}
