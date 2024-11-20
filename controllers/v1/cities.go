package v1

import (
	"context"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func GetCities(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCities")
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	cities := config.MI.DB.Collection(config.CITIES)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := cities.Aggregate(ctx, bson.A{
		bson.M{"$sort": bson.M{culture.Stringf("name.%v"): 1}},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
		bson.M{
			"$project": bson.M{
				"name": culture.Stringf("$name.%v"),
			},
		},
	})
	defer cursor.Close(ctx) // her func-da suny hem yatla!
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	var citiesResult = []models.City{}

	for cursor.Next(ctx) {
		var city models.City
		err := cursor.Decode(&city)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		citiesResult = append(citiesResult, city)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.City]{
		IsSuccess: true,
		Result:    citiesResult,
	})
}
