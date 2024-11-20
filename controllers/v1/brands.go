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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetBrandParents(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetBrandParents")
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	parentBrands := bson.A{
		bson.M{
			"$match": bson.M{
				"parent": nil,
			},
		},
		bson.M{"$sort": bson.M{"name": 1}},
		bson.M{
			"$skip": limit * pageIndex,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"name": 1,
				"logo": 1,
			},
		},
	}
	brandCategory := c.Query("category")
	if brandCategory != "" {
		category, err := primitive.ObjectIDFromHex(brandCategory)
		if err != nil {
			return c.JSON(errRes("Query(category)", err, config.CANT_DECODE))
		}
		parentBrands = bson.A{
			bson.M{
				"$match": bson.M{
					"parent":     bson.M{"$ne": nil},
					"categories": category,
				},
			},
			bson.M{
				"$group": bson.M{
					"_id": 0,
					"brands": bson.M{
						"$addToSet": "$parent",
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"brands": bson.M{
						"$slice": bson.A{"$brands", pageIndex * limit, limit},
					},
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from":         "brands",
					"localField":   "brands",
					"foreignField": "_id",
					"as":           "parent_brand",
					"pipeline": bson.A{
						bson.M{
							"$project": bson.M{
								"name": 1,
								"logo": 1,
							},
						},
					},
				},
			},
			bson.M{
				"$unwind": bson.M{
					"path": "$parent_brand",
				},
			},
			bson.M{
				"$replaceRoot": bson.M{
					"newRoot": "$parent_brand",
				},
			},
			bson.M{"$sort": bson.M{"name": 1}},
		}
	}
	brands := config.MI.DB.Collection(config.BRANDS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := brands.Aggregate(ctx, parentBrands)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	var brandsResult = []models.BrandParent{}

	for cursor.Next(ctx) {
		var brand models.BrandParent
		err := cursor.Decode(&brand)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		brandsResult = append(brandsResult, brand)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.BrandParent]{
		IsSuccess: true,
		Result:    brandsResult,
	})
}
func GetBrandChildren(c *fiber.Ctx) error {
	// culture := helpers.GetCultureFromQuery(c)
	errRes := helpers.ErrorResponse("Mobile.GetBrandChildren")
	parentBrandObjID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	brands := config.MI.DB.Collection(config.BRANDS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := brands.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"parent": parentBrandObjID,
			},
		},
		bson.M{"$sort": bson.M{"name": 1}},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
		bson.M{
			"$project": bson.M{
				"name": 1,
				"logo": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	var brandChildren = []models.BrandChild{}
	for cursor.Next(ctx) {
		var child models.BrandChild
		err = cursor.Decode(&child)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		brandChildren = append(brandChildren, child)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.BrandChild]{
		IsSuccess: true,
		Result:    brandChildren,
	})
}
