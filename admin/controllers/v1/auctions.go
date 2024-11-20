package v1

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"

	// taskmanager "github.com/devzatruk/bizhubBackend/task_manager"
	transactionmanager "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
)

func GetAuctions(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetAuctions")

	culture := helpers.GetCultureFromQuery(c)
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "3")) // default limit=3!!!
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}

	auctionsColl := config.MI.DB.Collection(config.AUCTIONS)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var auctionsResult []models.Auction

	auctionsCursor, err := auctionsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$sort": bson.M{
				"finished_at": -1,
			},
		},
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"image":       1,
				"heading":     fmt.Sprintf("$heading.%v", culture.Lang),
				"started_at":  1,
				"finished_at": 1,
				"is_finished": 1,
				"text_color":  1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer auctionsCursor.Close(ctx)
	for auctionsCursor.Next(ctx) {
		var auction models.Auction
		err := auctionsCursor.Decode(&auction)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		auctionsResult = append(auctionsResult, auction)
	}

	if auctionsResult == nil {
		auctionsResult = make([]models.Auction, 0)
	}

	return c.JSON(models.Response[[]models.Auction]{
		IsSuccess: true,
		Result:    auctionsResult,
	})
}
func CreateNewAuction(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.CreateNewAuction")
	payload := models.NewAuction{}
	payload.Description = models.Translation{En: c.FormValue("description"), Tm: "", Tr: "", Ru: ""}
	payload.Heading = models.Translation{En: c.FormValue("heading"), Tm: "", Tr: "", Ru: ""}
	payload.TextColor = c.FormValue("text_color")
	participants, err := strconv.Atoi(c.FormValue("participants"))
	if err != nil {
		return c.JSON(errRes("NoParticipants", err, config.BODY_NOT_PROVIDED))
	}
	payload.Participants = participants
	payload.InitialMinimalBid, err = strconv.Atoi(c.FormValue("initial_minimal_bid"))
	if err != nil {
		return c.JSON(errRes("NoInitialMinimalBid", err, config.BODY_NOT_PROVIDED))
	}
	startedAt, err := helpers.StringToDate(c.FormValue("started_at"))
	if err != nil {
		return c.JSON(errRes("StringToDate(started_at)", err, config.BODY_NOT_PROVIDED))
	}
	payload.StartedAt = startedAt
	finishedAt, err := helpers.StringToDate(c.FormValue("finished_at"))
	if err != nil {
		return c.JSON(errRes("StringToDate(finished_at)", err, config.BODY_NOT_PROVIDED))
	}
	payload.Winners = make([]models.AuctionDetailNewWinner, 0)
	payload.FinishedAt = finishedAt
	payload.CreatedAt = time.Now()
	payload.IsFinished = false
	payload.MinimalBid = 0
	image, err := helpers.SaveImageFile(c, "image", config.FOLDER_AUCTIONS)
	if err != nil {
		return c.JSON(errRes("SaveImageFile()", err, config.BODY_NOT_PROVIDED))
	}
	payload.Image = image // "images/auctions/1.webp"
	if payload.HasEmptyFields() {
		return c.JSON(errRes("HasEmptyFields()", errors.New("Some data not provided."), config.NOT_ALLOWED))
	}
	payload.Status = config.STATUS_CHECKING
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tr_manager := transactionmanager.NewTransaction(&ctx, config.MI.DB, 3)
	tr_auctionsColl := tr_manager.Collection(config.AUCTIONS)
	insert_model := transactionmanager.NewModel().SetDocument(payload)
	insertResult, err := tr_auctionsColl.InsertOne(insert_model)
	if err != nil {
		trErr := tr_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		if payload.Image != "" {
			helpers.DeleteImageFile(payload.Image)
		}
		return c.JSON(errRes("InsertOne(auctions)", err, config.CANT_INSERT))
	}
	config.CheckerTaskService.Writer.Auction(insertResult.InsertedID.(primitive.ObjectID),
		payload.Heading.En)

	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.CREATED,
	})
}
func GetAuctionDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetAuctionDetail")
	auctionId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	auctionsColl := config.MI.DB.Collection(config.AUCTIONS)
	var auctionResult models.AuctionDetail
	cursor, err := auctionsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": auctionId,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"heading":     "$heading.en",
				"description": "$description.en",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$winners",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "sellers",
				"localField":   "winners.seller_id",
				"foreignField": "_id",
				"as":           "winners.seller",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"_id":     1,
							"name":    1,
							"logo":    1,
							"city_id": 1,
						},
					},
					bson.M{
						"$lookup": bson.M{
							"from":         "cities",
							"localField":   "city_id",
							"foreignField": "_id",
							"as":           "city",
							"pipeline": bson.A{
								bson.M{
									"$addFields": bson.M{
										"name": "$name.en",
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path": "$city",
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$winners.seller",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": "_id",
				"root": bson.M{
					"$first": "$$ROOT",
				},
				"winners": bson.M{
					"$push": "$winners",
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"root.winners": "$winners",
			},
		},
		bson.M{
			"$replaceRoot": bson.M{
				"newRoot": "$root",
			},
		},
	})

	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}

	if cursor.Next(ctx) {
		err = cursor.Decode(&auctionResult)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}

	defer cursor.Close(ctx)

	if auctionResult.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Auction not found."), config.NOT_FOUND))
	}

	return c.JSON(models.Response[models.AuctionDetail]{
		IsSuccess: true,
		Result:    auctionResult,
	})
}
