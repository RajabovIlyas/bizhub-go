package v1

import (
	v1 "github.com/devzatruk/bizhubBackend/controllers/v1"
	"github.com/gofiber/fiber/v2"
)

func SetupV1PostRoutes(router fiber.Router) {
	posts := router.Group("posts")
	posts.Get("/", v1.GetAllPosts)
	// posts.Get("/deleteAll", v1.DeleteAllPosts) // TODO: gerek mi su route??? very dangerous!
	posts.Get("/:id", v1.GetPostDetails)
}
