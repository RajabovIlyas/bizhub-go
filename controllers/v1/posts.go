package v1

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetAllPosts(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetAllPosts")
	culture := helpers.GetCultureFromQuery(c)
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	posts := config.MI.DB.Collection(config.POSTS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var postsResult = []models.Post{}
	aggregationArray :=
		bson.A{
			bson.M{
				"$match": bson.M{
					"status": config.STATUS_PUBLISHED,
				},
			},
			bson.M{"$sort": bson.M{"created_at": -1}},
			bson.M{"$skip": pageIndex * limit},
			bson.M{"$limit": limit},
			bson.M{
				"$lookup": bson.M{
					"from":         "sellers",
					"localField":   "seller_id",
					"foreignField": "_id",
					"as":           "seller",
				},
			},
			bson.M{
				"$unwind": bson.M{
					"path": "$seller",
				},
			},

			bson.M{
				"$addFields": bson.M{
					"title": culture.Stringf("$title.%v"),
					"body":  culture.Stringf("$body.%v"),
				},
			},
			bson.M{
				"$project": bson.M{
					"seller.bio":      0,
					"seller.address":  0,
					"seller.owner_id": 0,
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from":         "cities",
					"localField":   "seller.city_id",
					"foreignField": "_id",
					"as":           "seller.city",
					"pipeline": bson.A{
						bson.M{
							"$addFields": bson.M{
								"name": culture.Stringf("$name.%v"),
							},
						},
					},
				},
			},
			bson.M{
				"$unwind": bson.M{
					"path":                       "$seller.city",
					"preserveNullAndEmptyArrays": true,
				},
			},

			bson.M{
				"$addFields": bson.M{
					"seller.city": bson.M{
						"$ifNull": bson.A{"$seller.city", nil},
					},
				},
			},
		}
	cursor, err := posts.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var post models.Post
		err = cursor.Decode(&post)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		postsResult = append(postsResult, post)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Post]{
		IsSuccess: true,
		Result:    postsResult,
	})
}

func DeleteAllPosts(c *fiber.Ctx) error { // TODO: su route oran howply! gerekmi?
	// errRes := helpers.ErrorResponse("Mobile.DeleteAllPosts")
	// posts := config.MI.DB.Collection(config.POSTS)
	// deleteResult, err := posts.DeleteMany(context.Background(), bson.M{})
	// if err != nil {
	// 	return c.JSON(errRes("DeleteMany()", err, config.DBQUERY_ERROR))
	// }
	// result := fiber.Map{
	// 	"deleted_count": deleteResult.DeletedCount,
	// 	"message":       config.DELETED,
	// }
	// return c.JSON(models.Response[fiber.Map]{
	// 	IsSuccess: true,
	// 	Result:    result,
	// })
	return nil
}
func GetPostDetails(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetPostDetails")
	posts := config.MI.DB.Collection(config.POSTS)
	postObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)
	var post models.PostDetail
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	updateResult, err := posts.UpdateOne(ctx, bson.M{"_id": postObjId}, bson.M{
		"$inc": bson.M{
			"viewed": 1,
		},
	})
	if err != nil {
		return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
	}
	if updateResult.MatchedCount == 0 {
		return c.JSON(errRes("ZeroMatchedCount", errors.New("Post not found."), config.NOT_FOUND))
	}
	aggregationArray := bson.A{

		bson.M{
			"$match": bson.M{
				"_id": postObjId,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "sellers",
				"localField":   "seller_id",
				"foreignField": "_id",
				"as":           "seller",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$seller",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"title": fmt.Sprintf("$title.%v", culture.Lang),
				"body":  fmt.Sprintf("$body.%v", culture.Lang),
			},
		},
		bson.M{
			"$project": bson.M{
				"seller.bio":      0,
				"seller.address":  0,
				"seller.owner_id": 0,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "cities",
				"localField":   "seller.city_id",
				"foreignField": "_id",
				"as":           "seller.city",
				"pipeline": bson.A{
					bson.M{
						"$addFields": bson.M{
							"name": fmt.Sprintf("$name.%v", culture.Lang),
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$seller.city",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"related_products": bson.M{
					"$map": bson.M{
						"input": "$related_products",
						"as":    "pro",
						"in": bson.M{
							"_id": "$$pro",
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "related_products._id",
				"foreignField": "_id",
				"as":           "related_products",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"_id": 1,
							"image": bson.M{
								"$first": "$images",
							},
						},
					},
				},
			},
		},
	}
	/*
		related_products = [ 62d8fa3e9346332069ff7088, 630c9e51daa4e8454b0f15e6]
		post id := 636f71e3727674239fbcf25a
		seller id := 62ce628c8cae982f654a3578
	*/
	cursor, err := posts.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	if cursor.Next(ctx) {
		err = cursor.Decode(&post)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	} else {
		return c.JSON(errRes("cursor.Next()", errors.New("Post not found."), config.DBQUERY_ERROR))
	}

	return c.JSON(models.Response[models.PostDetail]{
		IsSuccess: true,
		Result:    post,
	})
}
