package routes

import (
	"fmt"

	"github.com/devzatruk/bizhubBackend/ojologger"
	v1 "github.com/devzatruk/bizhubBackend/routes/v1"
	"github.com/gofiber/fiber/v2"
)

var (
	AuthV1Logger = ojologger.LoggerService.Logger("Auth v1")
)

func isGenuine(c *fiber.Ctx) error {
	// if ctx.GetReqHeaders()
	ojoMobile := c.GetReqHeaders()["Ojo-Mobile"]
	fmt.Printf("\nojoMobile: %v; \n %v\n", c.GetReqHeaders(), ojoMobile)
	// if err := bcrypt.CompareHashAndPassword([]byte(ojoMobile), []byte("omar ikinji ilon mask")); err != nil  {
	if ojoMobile != "by OJO dev." {
		return c.Status(fiber.StatusNotFound).SendString(fmt.Sprintf("Cannot %v %v", c.Method(), c.Path()))
	}

	return c.Next()
}

func SetupApiRoutes(app *fiber.App) {
	api := app.Group("api") // , isGenuine)
	v1.SetupRoutes(api)
}
