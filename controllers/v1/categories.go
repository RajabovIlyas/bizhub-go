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

func GetCategoryParents(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCategoryParents")
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	categories := config.MI.DB.Collection(config.CATEGORIES)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := categories.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"parent": nil,
			},
		},
		bson.M{"$sort": bson.M{culture.Stringf("name.%v"): 1}},
		bson.M{"$skip": limit * pageIndex},
		bson.M{"$limit": limit},
		bson.M{
			"$project": bson.M{
				"name":  culture.Stringf("$name.%v"),
				"image": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	var categoriesResult = []models.CategoryParent{}

	for cursor.Next(ctx) {
		var category models.CategoryParent
		err := cursor.Decode(&category)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		categoriesResult = append(categoriesResult, category)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.CategoryParent]{
		IsSuccess: true,
		Result:    categoriesResult,
	})
}
func GetCategoryAttributes(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCategoryAttributes")
	categoryObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	categoriesColl := config.MI.DB.Collection(config.CATEGORIES)
	aggregate := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": categoryObjId,
				"parent": bson.M{
					"$ne": nil,
				},
				// "image": bson.M{
				// 	"$eq": nil,
				// },
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "attributes",
				"localField":   "attributes",
				"foreignField": "_id",
				"as":           "attributes",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":        culture.Stringf("$name.%v"),
							"is_number":   1,
							"units_array": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"attributes": 1,
			},
		},
	}
	cursor, err := categoriesColl.Aggregate(ctx, aggregate)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)

	var category models.CategoryAttributes
	if cursor.Next(ctx) {
		err = cursor.Decode(&category)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if category.Id != categoryObjId {
		return c.JSON(errRes("CategoryNotFound", errors.New("Category not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[models.CategoryAttributes]{
		IsSuccess: true,
		Result:    category,
	})

}

// returns parent categories
func GetCategories(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCategories")
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	categoriesColl := config.MI.DB.Collection(config.CATEGORIES)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"parent": nil,
			},
		},
		bson.M{"$sort": bson.M{culture.Stringf("name.%v"): 1}},
		bson.M{"$skip": limit * pageIndex},
		bson.M{"$limit": limit},

		// bson.M{
		// 	"$lookup": bson.M{
		// 		"from":         "categories",
		// 		"localField":   "_id",
		// 		"foreignField": "parent",
		// 		"as":           "sub_categories",
		// 		"pipeline": bson.A{
		// 			bson.M{
		// 				"$sort": bson.M{
		// 					"order": 1,
		// 				},
		// 			},
		// 			bson.M{
		// 				"$project": bson.M{
		// 					"name": "$name.en",
		// 				},
		// 			},
		// 		},
		// 	},
		// },
		bson.M{
			"$project": bson.M{
				"name":  culture.Stringf("$name.%v"),
				"image": 1,
				// "sub_categories": 1,
			},
		},
	}
	var categories = []models.CategoryParent{}
	cursor, err := categoriesColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var cat models.CategoryParent
		err = cursor.Decode(&cat)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		categories = append(categories, cat)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.CategoryParent]{
		IsSuccess: true,
		Result:    categories,
	})
}

func GetCategoryChildren(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCategoryChildren")
	culture := helpers.GetCultureFromQuery(c)
	parentCategoryObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
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
	categories := config.MI.DB.Collection(config.CATEGORIES)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := categories.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"parent": parentCategoryObjId,
			},
		},
		bson.M{"$sort": bson.M{culture.Stringf("name.%v"): 1}},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
		bson.M{
			"$project": bson.M{
				"name":  culture.Stringf("$name.%v"),
				"image": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	var categoryChildren = []models.CategoryChild{}
	for cursor.Next(ctx) {
		var cat models.CategoryChild
		err = cursor.Decode(&cat)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		categoryChildren = append(categoryChildren, cat)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.CategoryChild]{
		IsSuccess: true,
		Result:    categoryChildren,
	})
}
