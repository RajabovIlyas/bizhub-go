package helpers

import (
	"errors"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetCurrentCustomer(c *fiber.Ctx, customerObjId *primitive.ObjectID) error {
	if user, ok := (c.Locals(config.CURRENT_USER).(map[string]any)); ok {
		objId, err := primitive.ObjectIDFromHex(user["_id"].(string))
		if err != nil {
			return err
		}
		// fmt.Printf("\nGetCurrentCustomer > customer Object ID: %v\n", objId)
		if objId != primitive.NilObjectID {
			*customerObjId = objId
			return nil
		} else {
			return errors.New("User ID is not valid.")
		}
	}
	return errors.New("Not a user.")
}
func GetCurrentSeller(c *fiber.Ctx, sellerID *primitive.ObjectID) error {
	if user, ok := (c.Locals(config.CURRENT_USER).(map[string]any)); ok {
		if user["seller_id"] == nil {
			return errors.New("Not a seller.")
		}
		sellerObjId, err := primitive.ObjectIDFromHex(user["seller_id"].(string))
		if err != nil {
			return err
		}
		if sellerObjId != primitive.NilObjectID {
			*sellerID = sellerObjId
			return nil
		} else {
			return errors.New("Seller Id is not valid.")
		}
	}
	return errors.New("Not a user.")
}
func GetCurrentEmployee(c *fiber.Ctx, empID *primitive.ObjectID) error {

	if user, ok := (c.Locals(config.CURRENT_EMPLOYEE).(map[string]any)); ok {
		employeeObjId, err := primitive.ObjectIDFromHex(user["_id"].(string))
		if err != nil {
			return err
		}
		if employeeObjId != primitive.NilObjectID {
			*empID = employeeObjId
			return nil
		} else {
			return errors.New("Employee ID is not valid.")
		}
	}
	return errors.New("Not an employee.")
}

func IsSeller(c *fiber.Ctx) bool {
	if user, ok := (c.Locals(config.CURRENT_USER).(map[string]any)); ok {
		if user["seller_id"] == nil {
			return false
		}
		sellerObjId, err := primitive.ObjectIDFromHex(user["seller_id"].(string))
		if err != nil {
			return false
		}
		if sellerObjId != primitive.NilObjectID {
			return true
		}
	}
	return false
}
