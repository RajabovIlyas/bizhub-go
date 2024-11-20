package middlewares

import (
	"errors"
	"fmt"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler func(c *fiber.Ctx) error

func AllowSeller() Handler {
	errRes := helpers.ErrorResponse("AllowSeller")

	return func(c *fiber.Ctx) error {
		if user, ok := (c.Locals(config.CURRENT_USER).(map[string]any)); ok {
			if user["seller_id"] == nil {
				return c.JSON(errRes("NilObjectID", errors.New("Only Sellers are allowed."), config.NO_PERMISSION))
			}
			sellerObjId, err := primitive.ObjectIDFromHex(user["seller_id"].(string))
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex()", err, config.NO_PERMISSION))
			}
			fmt.Printf("\nGetCurrentSellerInfo > Seller Object Id: %v\n", sellerObjId)
			if sellerObjId != primitive.NilObjectID {
				return c.Next() // seller is valid and has permission!
			} else {
				fmt.Printf("\nseller object id is primitive.NilObjectID: %v\n", sellerObjId)
				return c.JSON(errRes("NilObjectID", errors.New("Only Sellers are allowed."), config.NO_PERMISSION))
			}
		}
		return c.JSON(errRes("c.Locals(currentuser)", errors.New("\nAuthentication required."), config.AUTH_REQUIRED))
	}
}

// Allow Employees by ROLES
func AllowRoles(roles []string) Handler {
	errRes := helpers.ErrorResponse("AllowRoles")

	return func(c *fiber.Ctx) error {
		if employee, ok := (c.Locals(config.CURRENT_EMPLOYEE).(map[string]any)); ok {
			fmt.Printf("\nemployee.job >>> %v\n", employee["job"].(map[string]any)["name"])

			job, job_ok := employee["job"].(map[string]any)
			if !job_ok {
				return c.JSON(errRes("Employee[job]", errors.New("\nEmployee.job type assertion failed"), config.NO_PERMISSION))
			}
			job_name := job["name"].(string)
			if helpers.SliceContains(roles, job_name) {
				c.Locals(config.EMPLOYEE_JOB, job_name)
				return c.Next() // has permission
			} else {
				return c.JSON(errRes("Employee[job]", errors.New("\nNo permission."), config.NO_PERMISSION))
			}
		}
		return c.JSON(errRes("c.Locals(currentEmployee)", errors.New("\nAuthentication required."), config.AUTH_REQUIRED))
	}
}
