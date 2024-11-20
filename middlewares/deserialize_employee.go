package middlewares

import (
	"errors"
	"os"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/gofiber/fiber/v2"
)

func DeSerializeEmployee(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("DeSerializeEmployee")
	token, err := helpers.GetTokenFromHeader(c)
	// fmt.Println("\nGelen token =>", token)
	// fmt.Printf("\n-- Token: %v\n-- Path: %v\n-- ReqHeaders: %v\n", token, c.Path(), c.GetReqHeaders())
	if err != nil {
		return c.Status(401).JSON(errRes("GetTokenFromHeader()", err, config.AUTH_REQUIRED))
	}
	sub, err := helpers.ValidateToken(token, os.Getenv(config.ACCT_PUBLIC_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("ValidateToken()", err, config.ACCT_EXPIRED))
	}

	// fmt.Println("\nemployee claims['sub'] => ", sub)

	c.Locals(config.CURRENT_EMPLOYEE, sub)
	return c.Next()
}
func DeSerializeEmployeeFromQuery(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("DeSerializeEmployeeFromQuery")
	token := c.Query("token")
	if token == "" {
		return c.Status(401).JSON(errRes("Query(token)", errors.New("Employee not found."), config.AUTH_REQUIRED))
	}
	sub, err := helpers.ValidateToken(token, os.Getenv(config.ACCT_PUBLIC_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("ValidateToken()", err, config.ACCT_EXPIRED))
	}

	// fmt.Println("employee claims['sub'] => ", sub)

	c.Locals(config.CURRENT_EMPLOYEE, sub)
	return c.Next()
}

// login etmedik hem bolsa su route-lara girip bolyar!
func DeSerializeOptionalEmployee(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("DeSerializeOptionalEmployee")
	token, err := helpers.GetTokenFromHeader(c)
	if err != nil { // login etmedik, currentUser=nil, sonda da su route-a girip bolyar
		return c.Next()
	}

	sub, err := helpers.ValidateToken(token, os.Getenv(config.ACCT_PUBLIC_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("ValidateToken()", err, config.ACCT_EXPIRED)) // "REFT_EXPIRED"
	}

	// fmt.Println("employee claims['sub'] => ", sub)

	c.Locals(config.CURRENT_EMPLOYEE, sub)
	return c.Next()
}
