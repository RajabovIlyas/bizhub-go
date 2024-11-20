package v1

import (
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
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

// yeterlik seller bolanda tazeden gowy test etmeli!
// TODO: bulkwrite() yerine transaction etsek???
func BidAuction(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("BidAuction")
	auctionObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var bidSum struct {
		Sum float64 `json:"sum"`
	}
	err = c.BodyParser(&bidSum)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	var sellerObjId primitive.ObjectID
	err = helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	walletsColl := config.MI.DB.Collection(config.WALLETS)
	walletResult := walletsColl.FindOne(ctx, bson.M{"seller_id": sellerObjId})
	if err = walletResult.Err(); err != nil {
		return c.JSON(errRes("FindOne()", err, config.NOT_FOUND))
	}
	var sellerWallet models.SellerWallet
	err = walletResult.Decode(&sellerWallet)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	if sellerWallet.Balance < float64(bidSum.Sum) || sellerWallet.ClosedAt != nil || sellerWallet.Status != config.SELLER_STATUS_PUBLISHED {
		return c.JSON(errRes("BalanceNotEnough", errors.New("No enough funds."), config.NOT_ALLOWED))
	}
	for _, inAuction := range sellerWallet.InAuction {
		if inAuction.AuctionId == auctionObjId {
			return c.JSON(errRes("sellerWallet.InAuction", errors.New("Seller already in auction."), config.NOT_ALLOWED))
		}
	} // suna cenli problemsiz gelen bolsa, bid etsin
	auctionsColl := config.MI.DB.Collection(config.AUCTIONS)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": auctionObjId,
				"$expr": bson.M{
					"$and": bson.A{
						bson.M{
							"$gt": bson.A{"$$NOW", "$started_at"},
						},
						bson.M{
							"$lt": bson.A{"$$NOW", "$finished_at"},
						},
					},
				},
				"is_finished": false,
			},
		},
		bson.M{
			"$project": bson.M{
				"initial_min_bid": 1,
				"minimal_bid":     1,
				"winners":         1,
				"participants":    1,
				"heading":         1,
			},
		},
	}
	cursor, err := auctionsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	} // so the auction is valid!
	defer cursor.Close(ctx)
	var auction models.BidAuctionFind
	if cursor.Next(ctx) {
		err = cursor.Decode(&auction)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if auction.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Auction not found."), config.NOT_FOUND))
	}
	now := time.Now()
	if auction.MinimalBid < bidSum.Sum { // seller can bid!
		// fetch the last of winners in order to update some collections
		var lastWinner models.AuctionDetailWinner
		// update auctionsColl then update walletsColl
		newWinner := models.AuctionDetailNewWinner{
			SellerId:  sellerObjId,
			LastBid:   bidSum.Sum,
			CreatedAt: now,
		}
		bulkModels := []mongo.WriteModel{
			mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": auction.Id}).
				SetUpdate(
					bson.M{
						"$set": bson.M{
							"minimal_bid": bidSum.Sum,
						},
						"$push": bson.M{
							"winners": bson.M{
								"$each":     bson.A{newWinner},
								"$position": 0,
								// "$slice":    2,
							},
						},
					},
				),
		}
		if int64(len(auction.Winners)) == auction.Participants {
			// fmt.Printf("\nBidAuction: len(auction.Winners[]) == auction.Participants\n")

			lastWinner = auction.Winners[len(auction.Winners)-1]
			// fmt.Printf("\nBidAuction: lastWinner: %v\n", lastWinner)

			bulkModels = append(bulkModels,
				mongo.NewUpdateOneModel().
					SetFilter(bson.M{"_id": auction.Id}).
					SetUpdate(
						bson.M{
							"$pop": bson.M{"winners": 1},
						},
					))
		}
		// opts := options.BulkWrite().SetOrdered(false)
		updateRes, err := auctionsColl.BulkWrite(context.TODO(), bulkModels)
		if err != nil {
			if updateRes.ModifiedCount == 0 {
				return c.JSON(errRes("NewUpdateOneModel().PushWinner", err, config.CANT_UPDATE))
			}
			if updateRes.ModifiedCount == 1 {
				return c.JSON(errRes("NewUpdateOneModel().PopWinner", err, config.CANT_UPDATE))
			}
		}
		// fmt.Printf("\nBidAuction: auctionsColl.BulkWrite() successful.\n")

		// new winner's wallet must be updated!
		bulkModels = []mongo.WriteModel{
			mongo.NewUpdateOneModel().
				SetFilter(bson.M{"seller_id": newWinner.SellerId}).
				SetUpdate(
					bson.M{
						"$inc": bson.M{
							"balance": -newWinner.LastBid,
						},
						"$push": bson.M{
							"in_auction": bson.M{
								"$each": bson.A{fiber.Map{
									"auction_id": auction.Id,
									"amount":     newWinner.LastBid,
									"name":       auction.Heading,
								}},
								// "$position": 0,
							},
						},
					},
				),
		}
		// last winner's wallet must be updated if he has to be removed from winners list.
		if lastWinner.SellerId != primitive.NilObjectID {
			bulkModels = append(bulkModels,
				mongo.NewUpdateOneModel().
					SetFilter(bson.M{"seller_id": lastWinner.SellerId}).
					SetUpdate(
						bson.M{
							"$inc": bson.M{
								"balance": lastWinner.LastBid,
							},
							"$pull": bson.M{
								"in_auction": bson.M{
									"auction_id": auction.Id,
								},
							},
						},
					),
			)
		}
		var retryCount int = 0
	RetryPoint:
		// fmt.Printf("\nBidAuction: retryCount: %v\n", retryCount)

		updateRes, err = walletsColl.BulkWrite(context.TODO(), bulkModels)
		// fmt.Printf("\nWalletsColl BulkWrite Result: %v\n", updateRes)
		if err != nil {
			// fmt.Printf("\nBidAuction: walletsColl.BulkWrite() error!\n")

			if retryCount == 2 {
				return c.JSON(errRes("Retried walletsColl.BulkWrite() x 2", err, config.CANT_UPDATE))
			}
			if updateRes.ModifiedCount == 0 {
				// retry 2 times only!
				retryCount++
				goto RetryPoint
			}
			if updateRes.ModifiedCount == 1 && len(bulkModels) == 2 {
				// retry only second operation
				_, err = walletsColl.BulkWrite(context.TODO(), bulkModels[1:])
				if err != nil {
					return c.JSON(errRes("Retry: BulkWrite()[1:]", err, config.CANT_UPDATE))
				}
			}
		}
		// fmt.Printf("\nBidAuction: all updates successful!\n")

		return c.JSON(models.Response[fiber.Map]{
			IsSuccess: true,
			Result: fiber.Map{
				"balance": sellerWallet.Balance - float64(newWinner.LastBid),
				"in_auction": fiber.Map{
					"auction_id": auction.Id,
					"amount":     newWinner.LastBid,
					"name":       auction.Heading,
				},
			},
		})
	}
	return c.JSON(errRes("Auction.MinimalBid > BidSum", errors.New("Bid sum must be more than minimal bid amount."), config.NOT_ALLOWED))
}

func GetAuctionDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("GetAuctionDetail")
	culture := helpers.GetCultureFromQuery(c)
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
				// "$expr": bson.M{
				// 	"$and": bson.A{
				// 		bson.M{
				// 			"$gt": bson.A{"$$NOW", "$started_at"},
				// 		},
				// 		bson.M{
				// 			"$lt": bson.A{"$$NOW", "$finished_at"},
				// 		},
				// 	},
				// },
				// "is_finished": false,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"heading":     fmt.Sprintf("$heading.%v", culture.Lang),
				"description": fmt.Sprintf("$description.%v", culture.Lang),
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$winners",
				"preserveNullAndEmptyArrays": true,
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
							"type":    1,
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
										"name": fmt.Sprintf("$name.%v", culture.Lang),
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
				"path":                       "$winners.seller",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$set": bson.M{
				"winners": bson.M{
					"$cond": bson.A{
						bson.M{
							"$eq": bson.A{"$winners", bson.M{}},
						},
						"$$REMOVE",
						"$winners",
					},
				},
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": "$_id",
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
	},
	)

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

	var auctionsResult = []models.Auction{}

	auctionsCursor, err := auctionsColl.Aggregate(ctx, bson.A{
		// bson.M{
		// 	"$match": bson.M{
		// 		"$expr": bson.M{
		// 			"$and": bson.A{
		// 				bson.M{
		// 					"$gt": bson.A{"$$NOW", "$started_at"},
		// 				},
		// 				bson.M{
		// 					"$lt": bson.A{"$$NOW", "$finished_at"},
		// 				},
		// 			},
		// 		},
		// 		"is_finished": false,
		// 	},
		// },
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
				"heading":     culture.Stringf("$heading.%v"),
				"started_at":  1,
				"finished_at": 1,
				"is_finished": 1,
				"text_color":  1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.ACCT_EXPIRED))
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
	return c.JSON(models.Response[[]models.Auction]{
		IsSuccess: true,
		Result:    auctionsResult,
	})
}
