package v1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

func GetPackage(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetPackage")
	packType := c.Params("packageType")
	if !helpers.SliceContains(config.PACKAGE_TYPES, packType) {
		return c.JSON(errRes("Params(packageType)", errors.New("Package Type not valid."), config.PARAM_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	packagesColl := config.MI.DB.Collection(config.PACKAGES)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"type": packType,
			},
		},
		bson.M{
			"$project": bson.M{
				"name":         fmt.Sprintf("$name.%v", culture.Lang),
				"type":         1,
				"price":        1,
				"max_products": 1,
			},
		},
	}
	cursor, err := packagesColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var pkg models.PackageForPayment
	if cursor.Next(ctx) {
		err = cursor.Decode(&pkg)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}

	return c.JSON(models.Response[models.PackageForPayment]{
		IsSuccess: true,
		Result:    pkg,
	})

}
func GetPackages(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetPackages")
	culture := helpers.GetCultureFromQuery(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	packagesColl := config.MI.DB.Collection(config.PACKAGES)
	aggregationArray := bson.A{
		bson.M{
			"$project": bson.M{
				"name":         fmt.Sprintf("$name.%v", culture.Lang),
				"type":         1,
				"price":        1,
				"max_products": 1,
				"color":        1,
				"text_color":   1,
			},
		},
		bson.M{"$sort": bson.M{"price": 1}},
	}
	cursor, err := packagesColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var packages = []models.Package{}
	for cursor.Next(ctx) {
		var p models.Package
		err = cursor.Decode(&p)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		packages = append(packages, p)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Package]{
		IsSuccess: true,
		Result:    packages,
	})
}
