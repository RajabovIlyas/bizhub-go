package v1

import (
	"context"
	"fmt"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func GetBanners(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetBanners")
	culture := helpers.GetCultureFromQuery(c)
	bannersColl := config.MI.DB.Collection(config.BANNERS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := bannersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$project": bson.M{
				"image": fmt.Sprintf("$image.%v", culture.Lang),
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	var bannersResult = []string{}

	for cursor.Next(ctx) {
		var ban models.Banner
		err := cursor.Decode(&ban)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		bannersResult = append(bannersResult, ban.Image)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]string]{
		IsSuccess: true,
		Result:    bannersResult,
	})
}
