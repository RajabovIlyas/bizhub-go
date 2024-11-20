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
	transactionmanager "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func EditReporterBee(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.EditReporterBee")
	// admin mi ya basga birimi?
	var payload struct {
		Logo string `json:"logo" bson:"logo"`
		Name string `json:"name" bson:"name"`
	}
	payload.Name = c.FormValue("name")
	payload.Logo, _ = helpers.SaveImageFile(c, "logo", config.FOLDER_SELLERS)
	dbData := bson.M{}
	if len(payload.Name) > 0 {
		dbData["name"] = payload.Name
	}
	if len(payload.Logo) > 0 {
		dbData["logo"] = payload.Logo
	}
	if len(dbData) > 0 {
		forDb := bson.M{
			"$set": dbData,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		sellersColl := config.MI.DB.Collection(config.SELLERS)
		updateResult, err := sellersColl.UpdateOne(ctx, bson.M{"type": config.SELLER_TYPE_REPORTERBEE}, forDb)
		if err != nil {
			if payload.Logo != "" {
				helpers.DeleteImageFile(payload.Logo)
			}
			return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
		}
		if updateResult.ModifiedCount == 0 {
			if payload.Logo != "" {
				helpers.DeleteImageFile(payload.Logo)
			}
			return c.JSON(errRes("UpdateOne()", errors.New("Reporterbee not found."), config.NOT_FOUND))
		}
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.UPDATED,
		})
	} else {
		return c.JSON(errRes("HasEmptyFields", errors.New("No data provided to update."), config.NOT_ALLOWED))
	}
}
func GetReporterBee(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetReporterBee")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	findResult := config.MI.DB.Collection(config.SELLERS).FindOne(ctx, bson.M{"type": "reporterbee"})
	if err := findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(reporterbee)", err, config.NOT_FOUND))
	}
	type ReporterBee struct {
		Id   primitive.ObjectID `json:"_id" bson:"_id"`
		Name string             `json:"name" bson:"name"`
		Logo string             `json:"logo" bson:"logo"`
	}
	var reporterBee ReporterBee
	err := findResult.Decode(&reporterBee)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	return c.JSON(models.Response[ReporterBee]{
		IsSuccess: true,
		Result:    reporterBee,
	})
}
func GetReporterBeePosts(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetReporterBeePosts")

	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sellersColl := config.MI.DB.Collection(config.SELLERS)
	postsColl := config.MI.DB.Collection(config.POSTS)

	var reporterBee models.ReporterBee

	findResult := sellersColl.FindOne(ctx, bson.M{
		"type": config.SELLER_TYPE_REPORTERBEE,
	})
	if err := findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(reporterbee)", err, config.NOT_FOUND))
	}
	err = findResult.Decode(&reporterBee)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}

	var postsResult []models.ReporterBeePost

	postsCursor, err := postsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": reporterBee.Id,
			},
		},
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$addFields": bson.M{
				"title": "$title.en",
				"body":  "$body.en",
			},
		},
	})

	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer postsCursor.Close(ctx)
	for postsCursor.Next(ctx) {
		var post models.ReporterBeePost
		err = postsCursor.Decode(&post)
		if err != nil {
			return c.JSON(errRes("postsCursor.Decode()", err, config.CANT_DECODE))
		}
		post.ReporterBee = reporterBee
		postsResult = append(postsResult, post)
	}
	if err = postsCursor.Err(); err != nil {
		return c.JSON(errRes("postsCursor.Err()", err, config.DBQUERY_ERROR))
	}
	if postsResult == nil {
		postsResult = make([]models.ReporterBeePost, 0)
	}

	return c.JSON(models.Response[[]models.ReporterBeePost]{
		IsSuccess: true,
		Result:    postsResult,
	})
}
func CreateNewPost(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.CreateNewPost")
	var payload struct {
		Image       string
		Title       string
		Description string
	}
	payload.Title = c.FormValue("title")
	payload.Description = c.FormValue("description")
	image, err := helpers.SaveImageFile(c, "image", config.FOLDER_POSTS)
	if err != nil {
		return c.JSON(errRes("SaveImageFile()", err, config.BODY_NOT_PROVIDED))
	}
	if payload.Title == "" || payload.Description == "" {
		return c.JSON(errRes("HasEmptyFields", errors.New("Some data not provided."), config.BODY_NOT_PROVIDED))
	}
	payload.Image = image
	var reporterBee struct {
		Id primitive.ObjectID `bson:"_id"`
	} // sellersCollection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellersColl := config.MI.DB.Collection(config.SELLERS)
	findResult := sellersColl.FindOne(ctx, bson.M{"type": config.SELLER_TYPE_REPORTERBEE})
	if err := findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(reporterbee)", err, config.NOT_FOUND))
	}
	err = findResult.Decode(&reporterBee)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	t := payload.Title
	d := payload.Description
	newPost := models.PostUpsert{
		Title:           models.Translation{En: t, Tm: "", Ru: "", Tr: ""},
		Body:            models.Translation{En: d, Tm: "", Ru: "", Tr: ""},
		Image:           payload.Image,
		RelatedProducts: make([]primitive.ObjectID, 0),
		SellerId:        reporterBee.Id,
		Viewed:          0,
		Likes:           0,
		Status:          config.STATUS_CHECKING,
		Auto:            false,
		CreatedAt:       time.Now(),
	}
	tr_manager := transactionmanager.NewTransaction(&ctx, config.MI.DB, 3)
	tr_postsColl := tr_manager.Collection(config.POSTS)
	insert_model := transactionmanager.NewModel().SetDocument(newPost)
	insert_result, err := tr_postsColl.InsertOne(insert_model)
	if err != nil {
		trErr := tr_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("InsertOne(auctions)", err, config.CANT_INSERT))
	}

	config.CheckerTaskService.Writer.Post(insert_result.InsertedID.(primitive.ObjectID), true, newPost.Title.En, newPost.SellerId)

	// Insert into NEW-TASKS to be checked by admin-checkers!!
	// new_task := taskmanager.NewTask(config.MI.DB, 3)
	// new_task.Post(newPost.Id, newPost.Title.En, true, newPost.SellerId)
	// err = new_task.Commit()
	// if err != nil {
	// 	// transaction-y rollback etsek gowy bolar
	// 	trErr := tr_manager.Rollback()
	// 	if trErr != nil {
	// 		err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
	// 	}
	// 	return c.JSON(errRes("InsertOne(tasks)", err, config.CANT_INSERT))
	// }

	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.CREATED,
	})
}
func GetPostById(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetPostById")
	postObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellersColl := config.MI.DB.Collection(config.SELLERS)
	postsColl := config.MI.DB.Collection(config.POSTS)

	var reporterBee models.ReporterBee

	sellerSingleResult := sellersColl.FindOne(ctx, bson.M{
		"type": config.SELLER_TYPE_REPORTERBEE,
	})
	if err := sellerSingleResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(reporterbee)", err, config.NOT_FOUND))
	}
	err = sellerSingleResult.Decode(&reporterBee)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	postsCursor, err := postsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id":       postObjId,
				"seller_id": reporterBee.Id,
			},
		},
		bson.M{
			"$limit": 1,
		},
		bson.M{
			"$addFields": bson.M{
				"title": "$title.en",
				"body":  "$body.en",
			},
		},
	})

	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer postsCursor.Close(ctx)
	var post models.ReporterBeePostDetail
	if postsCursor.Next(ctx) {
		err = postsCursor.Decode(&post)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if post.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Post not found."), config.NOT_FOUND))
	}

	return c.JSON(models.Response[models.ReporterBeePostDetail]{
		IsSuccess: true,
		Result:    post,
	})
}
