package middlewares

import (
	"os"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/gofiber/fiber/v2"
)

func DeSerializeCustomer(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("DeSerializeCustomer")
	token, err := helpers.GetTokenFromHeader(c)
	if err != nil {
		return c.Status(401).JSON(errRes("GetTokenFromHeader()", err, config.AUTH_REQUIRED))
	}

	sub, err := helpers.ValidateToken(token, os.Getenv(config.ACCT_PUBLIC_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("ValidateToken()", err, config.ACCT_EXPIRED))
	}

	// fmt.Println("token user claims['sub'] => ", sub)

	c.Locals(config.CURRENT_USER, sub)
	return c.Next()
}

// login etmedik hem bolsa su route-lara girip bolyar!
func DeSerializeOptionalCustomer(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("DeSerializeOptionalCustomer")
	token, err := helpers.GetTokenFromHeader(c)
	if err != nil { // login etmedik, currentUser=nil, sonda da su route-a girip bolyar
		return c.Next()
	}

	sub, err := helpers.ValidateToken(token, os.Getenv(config.ACCT_PUBLIC_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("ValidateToken()", err, config.ACCT_EXPIRED)) // "REFT_EXPIRED"
	}

	// fmt.Println("token user claims['sub'] => ", sub)

	c.Locals(config.CURRENT_USER, sub)
	return c.Next()
}
