package v1

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetSuggestions(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetSuggestions")
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aggregationArray := bson.A{
		bson.M{
			"$sort": bson.M{
				"created_at": -1, // in tazeler basda gorunsun
				"type":       -1,
			},
		},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
	}
	suggestionsColl := config.MI.DB.Collection("suggestions")
	cursor, err := suggestionsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var suggestions = []models.Suggestion{}
	for cursor.Next(ctx) {
		var sug models.Suggestion
		err = cursor.Decode(&sug)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		suggestions = append(suggestions, sug)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Suggestion]{
		IsSuccess: true,
		Result:    suggestions,
	})
}
func AddSuggestion(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.AddSuggestion")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	var sug models.Suggestion
	err = c.BodyParser(&sug)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	if !(sug.Type == models.BRAND || sug.Type == models.CATEGORY) {
		return c.JSON(errRes("InvalidType", errors.New("Suggestion type invalid."), config.TYPE_INVALID))
	}
	sug.SellerId = sellerObjId
	sug.CreatedAt = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sugColl := config.MI.DB.Collection(config.SUGGESTIONS)
	_, err = sugColl.InsertOne(ctx, sug)
	if err != nil {
		return c.JSON(errRes("InsertOne()", err, config.CANT_INSERT))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    "Thank you for your suggestion!",
	})
}
