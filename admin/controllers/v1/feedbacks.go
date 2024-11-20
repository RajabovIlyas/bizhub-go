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

func GetFeedbacks(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetFeedbacks")
	var managerObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &managerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	filter := c.Query("filter", "all") // all, today, this_week, this_month
	sort := c.Query("sort", "all")     // all, unread, read

	aggregationArray := bson.A{}
	if sort == "unread" {
		aggregationArray = append(aggregationArray, bson.M{
			"$match": bson.M{
				"is_read": false,
			},
		})
	} else if sort == "read" {
		aggregationArray = append(aggregationArray, bson.M{
			"$match": bson.M{
				"is_read": true,
			},
		})
	}
	now := time.Now()
	y, m, d := now.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	this_week := yesterday.AddDate(0, 0, -6) // duyne gora 6 gun suysmeli!
	this_month_first := time.Date(y, m, 1, 0, 0, 0, 0, time.Local)
	this_month := this_month_first.AddDate(0, 0, -1)
	if filter == "today" {
		aggregationArray = append(aggregationArray, bson.M{
			"$match": bson.M{
				"created_at": bson.M{
					"$gt": yesterday,
				},
			},
		},
		)
	} else if filter == "this_week" {
		aggregationArray = append(aggregationArray, bson.M{
			"$match": bson.M{
				"created_at": bson.M{
					"$gt": this_week,
				},
			},
		},
		)
	} else if filter == "this_month" {
		aggregationArray = append(aggregationArray, bson.M{
			"$match": bson.M{
				"created_at": bson.M{
					"$gt": this_month,
				},
			},
		},
		)
	}
	aggregationArray = append(aggregationArray, bson.A{
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "sellers",
				"localField":   "sent_by",
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
			"$lookup": bson.M{
				"from":         "cities",
				"localField":   "seller.city_id",
				"foreignField": "_id",
				"as":           "seller.city",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$seller.city",
			},
		},
		bson.M{
			"$project": bson.M{
				"seller": bson.M{
					"_id":       1, // "$seller._id",
					"name":      1, // "$seller.name",
					"type":      1, // "$seller.type",
					"logo":      1, // "$seller.logo",
					"city.name": "$seller.city.name.en",
					"city._id":  "$seller.city._id",
				},
				"text":       1,
				"is_read":    1,
				"created_at": 1,
			},
		},
	}...)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	feedsColl := config.MI.DB.Collection(config.FEEDBACKS)
	cursor, err := feedsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var feeds []models.Feedback
	for cursor.Next(ctx) {
		var feed models.Feedback
		err = cursor.Decode(&feed)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		feeds = append(feeds, feed)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if feeds == nil {
		feeds = make([]models.Feedback, 0)
	}
	return c.JSON(models.Response[[]models.Feedback]{
		IsSuccess: true,
		Result:    feeds,
	})
}
