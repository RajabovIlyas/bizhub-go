package v1

import (
	"github.com/devzatruk/bizhubBackend/seeders"
	"github.com/gofiber/fiber/v2"
)

func SetupV1SeederRoutes(router fiber.Router) {
	seeder := router.Group("seeder")
	seeder.Get("/posts", seeders.SeedPosts)
}
