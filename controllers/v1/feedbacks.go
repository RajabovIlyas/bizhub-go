package v1

import (
	"context"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateFeedback(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.CreateFeedback")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	var feedback struct {
		Text string `json:"text" bson:"text"`
	}
	err = c.BodyParser(&feedback)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	feeds := config.MI.DB.Collection(config.FEEDBACKS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = feeds.InsertOne(ctx, bson.M{
		"text":       feedback.Text,
		"sent_by":    sellerObjId,
		"read_by":    nil,
		"is_read":    false,
		"created_at": time.Now(),
	})
	if err != nil {
		return c.JSON(errRes("InsertOne()", err, config.CANT_INSERT))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.CREATED,
	})
}
