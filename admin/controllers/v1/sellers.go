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
	ojoTr "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetAllSellers(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetAllSellers")
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
	seller_type := c.Query("type", "all")
	status := c.Query("status", "all")

	sellersColl := config.MI.DB.Collection(config.SELLERS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aggregateArray := bson.A{}
	if seller_type == "all" {
		aggregateArray = append(aggregateArray, bson.M{
			"$match": bson.M{
				"type": bson.M{
					"$not": bson.M{
						"$eq": "reporterbee",
					},
				},
			},
		},
		)
	} else {
		aggregateArray = append(aggregateArray, bson.M{
			"$match": bson.M{
				"type": seller_type,
			},
		},
		)
	}
	if status != "all" {
		aggregateArray = append(aggregateArray, bson.M{
			"$match": bson.M{
				"status": status,
			},
		},
		)
	}
	aggregateArray = append(aggregateArray, bson.A{
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "cities",
				"localField":   "city_id",
				"foreignField": "_id",
				"as":           "city",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$city",
			},
		},
		bson.M{
			"$project": bson.M{
				"logo": 1,
				"name": 1,
				"city": bson.M{
					"_id":  "$city._id",
					"name": "$city.name.tm",
				},
				"type":   1,
				"status": 1,
			},
		},
	}...)
	cursor, err := sellersColl.Aggregate(ctx, aggregateArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var sellers []models.SellerWithStatus
	for cursor.Next(ctx) {
		var seller models.SellerWithStatus
		// var seller interface{}
		err = cursor.Decode(&seller)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		sellers = append(sellers, seller)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if sellers == nil {
		sellers = make([]models.SellerWithStatus, 0)
	}
	sellersCount, _ := sellersColl.EstimatedDocumentCount(ctx)
	var sellersWithCount models.SellersWithCount
	sellersWithCount.Sellers = sellers
	sellersWithCount.SellersCount = sellersCount - 1
	return c.JSON(models.Response[models.SellersWithCount]{
		IsSuccess: true,
		Result:    sellersWithCount,
	})
}

func GetSellerProfile(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetSellerProfile")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var sellerProfileResult models.SellerProfile
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sellersColl := config.MI.DB.Collection(config.SELLERS)

	cursor, err := sellersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": sellerObjId,
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
		bson.M{
			"$lookup": bson.M{
				"from":         "categories",
				"localField":   "categories",
				"foreignField": "_id",
				"as":           "categories",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name": "$name.tm",
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "packages",
				"localField":   "package.type",
				"foreignField": "type",
				"as":           "package.detail",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name": "$name.en",
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$package.detail",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"package.name": "$package.detail.name",
				"address":      "$address.en",
				"bio":          "$bio.en",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "wallet_history",
				"localField":   "transfers",
				"foreignField": "_id",
				"as":           "transfers",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "employees",
							"localField":   "employee_id",
							"foreignField": "_id",
							"as":           "by",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"full_name": bson.M{
											"$concat": bson.A{"$name", " ", "$surname"},
										},
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path":                       "$by",
							"preserveNullAndEmptyArrays": true,
						},
					},
					bson.M{
						"$project": bson.M{
							"created_at": 1,
							"intent":     1,
							"code":       1,
							"amount":     1,
							"by": bson.M{
								"$ifNull": bson.A{"$by", nil},
							},
							"note": "$note.en",
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"categories": bson.M{
					"$map": bson.M{
						"input": "$categories",
						"as":    "category",
						"in":    "$$category.name",
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "customers",
				"localField":   "owner_id",
				"foreignField": "_id",
				"as":           "owner",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"phone": 1,
							"name":  1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$owner",
			},
		},
		bson.M{
			"$project": bson.M{
				"package.detail": 0,
				"owner_id":       0,
				"likes":          0,
				"posts_count":    0,
				"products_count": 0,
				"city_id":        0,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}

	if cursor.Next(ctx) {
		err = cursor.Decode(&sellerProfileResult)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}

	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}

	if helpers.IsNilObjectID(sellerProfileResult.Id) {
		return c.JSON(errRes("IsNilObjectID()", errors.New("Seller profile not found."), config.NOT_FOUND))
	}

	return c.JSON(models.Response[models.SellerProfile]{
		IsSuccess: true,
		Result:    sellerProfileResult,
	})
}

// TODO: indi transfers collection yerine wallet_history ulanmaly!!!
func GetSellerTransfers(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetSellerTransfers")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, ""))
	}

	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, ""))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transfersColl := config.MI.DB.Collection(config.WALLETHISTORY)
	cursor, err := transfersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
				"status":    "completed",
			},
		},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
		bson.M{
			"$lookup": bson.M{
				"from":         "employees",
				"localField":   "employee_id",
				"foreignField": "_id",
				"as":           "by",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"full_name": bson.M{
								"$concat": bson.A{"$name", " ", "$surname"},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$by",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$project": bson.M{
				"created_at": 1,
				"intent":     1,
				"code":       1,
				"amount":     1,
				"by": bson.M{
					"$ifNull": bson.A{"$by", nil},
				},
				"note": "$note.en",
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var transfers []models.SellerProfileTransfer
	for cursor.Next(ctx) {
		var trans models.SellerProfileTransfer
		err = cursor.Decode(&trans)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		transfers = append(transfers, trans)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if transfers == nil {
		transfers = make([]models.SellerProfileTransfer, 0)
	}
	return c.JSON(models.Response[[]models.SellerProfileTransfer]{
		IsSuccess: true,
		Result:    transfers,
	})
}
func GetSellerPackageHistory(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetSellerPackageHistory")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}

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
	packageHistoryColl := config.MI.DB.Collection(config.PACKAGEHISTORY)
	cursor, err := packageHistoryColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
			},
		},
		bson.M{
			"$sort": bson.M{"created_at": -1},
		},
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"created_at": 1,
				"text":       1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var history []models.SellerPackageHistory
	for cursor.Next(ctx) {
		var event models.SellerPackageHistory
		err = cursor.Decode(&event)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		history = append(history, event)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if history == nil {
		history = make([]models.SellerPackageHistory, 0)
	}
	return c.JSON(models.Response[[]models.SellerPackageHistory]{
		IsSuccess: true,
		Result:    history,
	})
}

func ExtendSellerPackage(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.ExtendSellerPackage")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var extendBy struct {
		Days int64 `json:"days"`
	}
	err = c.BodyParser(&extendBy)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	packageHistoryColl := config.MI.DB.Collection(config.PACKAGEHISTORY)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := packageHistoryColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
			},
		},
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$limit": 1,
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var sellerPack models.SellerPackageHistoryFull
	if cursor.Next(ctx) {
		err = cursor.Decode(&sellerPack)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if sellerPack.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Seller not found."), config.NOT_FOUND))
	}
	sellerPack.Id = primitive.NewObjectID()
	now := time.Now()
	to := sellerPack.To
	if now.Before(to) {
		sellerPack.To = to.AddDate(0, 0, int(extendBy.Days))
		sellerPack.From = to
	} else if now.After(to) || now.Equal(to) {
		sellerPack.From = now
		sellerPack.To = now.AddDate(0, 0, int(extendBy.Days))
	}
	sellerPack.Action = "extended"
	sellerPack.Text = fmt.Sprintf("Extended for %v days.", extendBy.Days)
	sellerPack.CreatedAt = time.Now()

	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_packageHistoryColl := transaction_manager.Collection(config.PACKAGEHISTORY)
	insert_model := ojoTr.NewModel().SetDocument(sellerPack)
	insertResult, err := tr_packageHistoryColl.InsertOne(insert_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("InsertOne(package_history)", err, config.CANT_INSERT))
	}

	tr_sellersColl := transaction_manager.Collection(config.SELLERS)
	update_model := ojoTr.NewModel().SetFilter(bson.M{"_id": sellerObjId}).
		SetUpdate(bson.M{
			"$set": bson.M{
				"package.to":         sellerPack.To,
				"package_history_id": insertResult.InsertedID.(primitive.ObjectID),
			},
		}).
		SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
			oldData := i.(bson.M)
			p := oldData["package"].(bson.M)
			return bson.M{
				"$set": bson.M{
					"package.to":                 p["to"],
					"package.package_history_id": p["package_history_id"],
				},
			}
		})
	_, err = tr_sellersColl.FindOneAndUpdate(update_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(sellers)", err, config.CANT_UPDATE))
	}

	return c.JSON(models.Response[models.SellerPackageHistoryFull]{
		IsSuccess: true,
		Result:    sellerPack,
	})
}

// TODO: galan yerim
/*
	bir selleri block edende name bolyar? products, posts, profile gorunenokmy?
	dine account data gorunenok, yagny profile bolyar oydyan.
*/
func BlockSeller(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.BlockSeller")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellersColl := config.MI.DB.Collection(config.SELLERS)
	updateResult, err := sellersColl.UpdateOne(ctx, bson.M{"_id": sellerObjId}, bson.M{
		"$set": bson.M{
			"status": config.SELLER_STATUS_BLOCKED,
		},
	})
	if err != nil {
		return c.JSON(errRes("UpdateOne(seller)", err, config.CANT_UPDATE))
	}
	if updateResult.MatchedCount == 0 {
		return c.JSON(errRes("ZeroMatchCount", errors.New("No documents matched."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.STATUS_COMPLETED,
	})
}
func UnBlockSeller(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.UnBlockSeller")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellersColl := config.MI.DB.Collection(config.SELLERS)
	updateResult, err := sellersColl.UpdateOne(ctx, bson.M{"_id": sellerObjId}, bson.M{
		"$set": bson.M{
			"status": config.SELLER_STATUS_PUBLISHED,
		},
	})
	if err != nil {
		return c.JSON(errRes("UpdateOne(seller)", err, config.CANT_UPDATE))
	}
	if updateResult.MatchedCount == 0 {
		return c.JSON(errRes("ZeroMatchCount", errors.New("No documents matched."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.STATUS_COMPLETED,
	})
}
func PromoteSellerType(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.PromoteSellerType")
	var sellerObjId primitive.ObjectID
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellersColl := config.MI.DB.Collection(config.SELLERS)
	updateResult, err := sellersColl.UpdateOne(ctx, bson.M{"_id": sellerObjId}, bson.M{
		"$set": bson.M{
			"type": config.SELLER_TYPE_MANUFACTURER,
		}})
	if err != nil {
		return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
	}
	if updateResult.MatchedCount == 0 {
		return c.JSON(errRes("ZeroMatchCount", errors.New("No documents matched."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    "Seller promoted succesfully.",
	})
}
func DemoteSellerType(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.DemoteSellerType")
	var sellerObjId primitive.ObjectID
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellersColl := config.MI.DB.Collection(config.SELLERS)
	updateResult, err := sellersColl.UpdateOne(ctx, bson.M{"_id": sellerObjId}, bson.M{
		"$set": bson.M{
			"type": config.SELLER_TYPE_REGULAR,
		}})
	if err != nil {
		return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
	}
	if updateResult.MatchedCount == 0 {
		return c.JSON(errRes("ZeroMatchCount", errors.New("No documents matched."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    "Seller demoted succesfully.",
	})
}
func GetSellerProducts(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetSellerProducts")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	now := time.Now()
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	last_week := today.AddDate(0, 0, -6)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
			},
		},
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
			"$project": bson.M{
				"image": bson.M{
					"$first": "$images",
				},
				"heading":  "$heading.en",
				"price":    1,
				"discount": 1,
				"is_new": bson.M{
					"$gt": bson.A{"$created_at", last_week},
				},
			},
		},
	}
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := productsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var products []models.Product

	for cursor.Next(ctx) {
		var prod models.Product
		err = cursor.Decode(&prod)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		products = append(products, prod)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if products == nil {
		products = make([]models.Product, 0)
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    products,
	})
}
func GetSellerProductDetails(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetSellerProductDetails")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	prodObjId, err := primitive.ObjectIDFromHex(c.Params("productId"))
	if err != nil {
		return c.JSON(errRes("Params(productId)", err, config.PARAM_NOT_PROVIDED))
	}
	products := config.MI.DB.Collection(config.PRODUCTS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// aggregationArray := bson.A{

	// 	bson.M{
	// 		"$match": bson.M{
	// 			"_id":       prodObjId,
	// 			"seller_id": sellerObjId,
	// 		},
	// 	},
	// 	bson.M{
	// 		"$lookup": bson.M{
	// 			"from":         "categories",
	// 			"localField":   "category_id",
	// 			"foreignField": "_id",
	// 			"as":           "category",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$unwind": bson.M{
	// 			"path": "$category",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$lookup": bson.M{
	// 			"from":         "categories",
	// 			"localField":   "category.parent",
	// 			"foreignField": "_id",
	// 			"as":           "category.parent",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$unwind": bson.M{
	// 			"path": "$category.parent",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$addFields": bson.M{
	// 			"heading":              "$heading.en",
	// 			"more_details":         "$more_details.en",
	// 			"category.name":        "$category.name.en",
	// 			"category.parent.name": "$category.parent.name.en",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$lookup": bson.M{
	// 			"from":         "brands",
	// 			"localField":   "brand_id",
	// 			"foreignField": "_id",
	// 			"as":           "brand",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$unwind": bson.M{
	// 			"path": "$brand",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$lookup": bson.M{
	// 			"from":         "brands",
	// 			"localField":   "brand.parent",
	// 			"foreignField": "_id",
	// 			"as":           "brand.parent",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$unwind": bson.M{
	// 			"path": "$brand.parent",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$project": bson.M{
	// 			"category.image":             0,
	// 			"category.order":             0,
	// 			"category.attributes":        0,
	// 			"category.parent.image":      0,
	// 			"category.parent.parent":     0,
	// 			"category.parent.order":      0,
	// 			"category.parent.attributes": 0,
	// 			"brand.logo":                 0,
	// 			"brand.parent.logo":          0,
	// 			"brand.parent.parent":        0,
	// 		},
	// 	},
	// 	bson.M{
	// 		"$facet": bson.M{
	// 			"root": bson.A{},
	// 			"attrs": bson.A{
	// 				bson.M{
	// 					"$unwind": bson.M{
	// 						"path": "$attrs",
	// 					},
	// 				},
	// 				bson.M{
	// 					"$lookup": bson.M{
	// 						"from":         "attributes",
	// 						"localField":   "attrs.attr_id",
	// 						"foreignField": "_id",
	// 						"as":           "attrs.attr",
	// 					},
	// 				},
	// 				bson.M{
	// 					"$unwind": bson.M{
	// 						"path": "$attrs.attr",
	// 					},
	// 				},
	// 				bson.M{
	// 					"$lookup": bson.M{
	// 						"from":         "units",
	// 						"localField":   "attrs.attr.unit",
	// 						"foreignField": "_id",
	// 						"as":           "attrs.attr.unit",
	// 					},
	// 				},
	// 				bson.M{
	// 					"$unwind": bson.M{
	// 						"path": "$attrs.attr.unit",
	// 					},
	// 				},
	// 				bson.M{
	// 					"$addFields": bson.M{
	// 						"attrs.attr.name":        "$attrs.attr.name.en",
	// 						"attrs.attr.placeholder": "$attrs.attr.placeholder.en",
	// 					},
	// 				},
	// 				bson.M{
	// 					"$group": bson.M{
	// 						"_id": "$_id",
	// 						"attrs": bson.M{
	// 							"$push": "$attrs",
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// 	bson.M{
	// 		"$addFields": bson.M{
	// 			"attrs": bson.M{
	// 				"$arrayElemAt": bson.A{"$attrs.attrs", 0},
	// 			},
	// 		},
	// 	},
	// 	bson.M{
	// 		"$addFields": bson.M{
	// 			"root.attrs": bson.M{
	// 				"$cond": bson.A{
	// 					bson.M{
	// 						"$isArray": bson.A{"$attrs"},
	// 					},
	// 					"$attrs",
	// 					bson.A{},
	// 				},
	// 			},
	// 		},
	// 	},
	// 	bson.M{
	// 		"$replaceRoot": bson.M{
	// 			"newRoot": bson.M{
	// 				"$arrayElemAt": bson.A{"$root", 0},
	// 			},
	// 		},
	// 	},
	// }

	aggregationArray := bson.A{

		bson.M{
			"$match": bson.M{
				"_id":       prodObjId,
				"seller_id": sellerObjId,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "sellers",
				"localField":   "seller_id",
				"foreignField": "_id",
				"as":           "seller",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "cities",
							"localField":   "city_id",
							"foreignField": "_id",
							"as":           "city",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
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
					bson.M{
						"$project": bson.M{
							"name": 1,
							"type": 1,
							"city": 1,
							"logo": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$seller",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "attributes",
				"localField":   "attrs.attr_id",
				"foreignField": "_id",
				"as":           "attrs_detail",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":        "$name.en",
							"units_array": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"attrs": bson.M{
					"$map": bson.M{
						"input": "$attrs",
						"in": bson.M{
							"$mergeObjects": bson.A{
								"$$this",
								bson.M{
									"attr": bson.M{
										"$arrayElemAt": bson.A{
											"$attrs_detail",
											bson.M{
												"$indexOfArray": bson.A{"$attrs", "$$this"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"attrs": bson.M{
					"$map": bson.M{
						"input": "$attrs",
						"in": bson.M{
							"$mergeObjects": bson.A{
								"$$this",
								bson.M{
									"unit": bson.M{
										"$arrayElemAt": bson.A{
											"$$this.attr.units_array",
											"$$this.unit_index",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "categories",
				"localField":   "category_id",
				"foreignField": "_id",
				"as":           "category",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "categories",
							"localField":   "parent",
							"foreignField": "_id",
							"as":           "parent",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"name": "$name.en",
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path": "$parent",
						},
					},
					bson.M{
						"$project": bson.M{
							"name":   "$name.en",
							"parent": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "brands",
				"localField":   "brand_id",
				"foreignField": "_id",
				"as":           "brand",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "brands",
							"localField":   "parent",
							"foreignField": "_id",
							"as":           "parent",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"name": 1,
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path": "$parent",
						},
					},
					bson.M{
						"$project": bson.M{
							"name":   1,
							"parent": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$category",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$brand",
			},
		},
		bson.M{
			"$project": bson.M{
				"seller":        1,
				"category":      1,
				"brand":         1,
				"attrs":         1,
				"heading":       "$heading.en",
				"more_details":  "$more_details.en",
				"likes":         1,
				"viewed":        1,
				"status":        1,
				"discount":      1,
				"price":         1,
				"discount_data": 1,
				"brand_id":      1,
				"category_id":   1,
				"images":        1,
				"seller_id":     1,
			},
		},
	}
	cursor, err := products.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var productDetail models.ProductDetailWithoutSeller
	if cursor.Next(ctx) {
		err = cursor.Decode(&productDetail)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	} else {
		return c.JSON(errRes("cursor.Next()", errors.New("Product not found."), config.NOT_FOUND))
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[models.ProductDetailWithoutSeller]{
		IsSuccess: true,
		Result:    productDetail,
	})
}

func FindSeller(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.FindSeller")

	var sellerPhone string
	sellerPhone = c.Query("phone")
	if len(sellerPhone) == 0 {
		return c.JSON(errRes("Query(phone)", errors.New("Phone number not provided"), config.QUERY_NOT_PROVIDED))
	}
	// err := c.BodyParser(&sellerPhone) // POST REQUEST etsek suny ulan!
	// if err != nil {
	// 	return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	// }
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	customersColl := config.MI.DB.Collection(config.CUSTOMERS)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"phone": sellerPhone,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "sellers",
				"localField":   "seller_id",
				"foreignField": "_id",
				"as":           "seller",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "cities",
							"localField":   "city_id",
							"foreignField": "_id",
							"as":           "city",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
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
					bson.M{
						"$project": bson.M{
							"name": 1,
							"logo": 1,
							"type": 1,
							"city": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$seller",
			},
		},
		bson.M{
			"$replaceRoot": bson.M{
				"newRoot": "$seller",
			},
		},
	}
	cursor, err := customersColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var seller models.Seller
	if cursor.Next(ctx) {
		err = cursor.Decode(&seller)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if seller.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Seller not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[models.Seller]{
		IsSuccess: true,
		Result:    seller,
	})
}
func GetSellerPostDetails(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetSellerPostDetails")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	postObjId, err := primitive.ObjectIDFromHex(c.Params("postId"))
	if err != nil {
		return c.JSON(errRes("Params(postId)", err, config.PARAM_NOT_PROVIDED))
	}
	postsColl := config.MI.DB.Collection(config.POSTS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id":       postObjId,
				"seller_id": sellerObjId,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"title": "$title.en",
				"body":  "$body.en",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "related_products",
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
	cursor, err := postsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var postDetail models.PostDetailWithoutSeller
	if cursor.Next(ctx) {
		err = cursor.Decode(&postDetail)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	} else {
		return c.JSON(errRes("cursor.Next()", errors.New("Post not found."), config.NOT_FOUND))
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[models.PostDetailWithoutSeller]{
		IsSuccess: true,
		Result:    postDetail,
	})
}

func GetSellerPosts(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetSellerPosts")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
				"auto":      false,
				"status":    config.STATUS_PUBLISHED,
			},
		},
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
			"$project": bson.M{
				"seller_id": 1,
				"image":     1,
				"title":     "$title.en",
				"viewed":    1,
				"likes":     1,
			},
		},
	}
	postsColl := config.MI.DB.Collection(config.POSTS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := postsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var posts []models.PostWithoutSeller

	for cursor.Next(ctx) {
		var prod models.PostWithoutSeller
		err = cursor.Decode(&prod)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		posts = append(posts, prod)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if posts == nil {
		posts = make([]models.PostWithoutSeller, 0)
	}
	return c.JSON(models.Response[[]models.PostWithoutSeller]{
		IsSuccess: true,
		Result:    posts,
	})
}

func ReduceSellerPackage(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.ReduceSellerPackage")
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var extendBy struct {
		Days int64 `json:"days"`
	}
	err = c.BodyParser(&extendBy)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	packageHistoryColl := config.MI.DB.Collection(config.PACKAGEHISTORY)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := packageHistoryColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
			},
		},
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$limit": 1,
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var sellerPack models.SellerPackageHistoryFull
	if cursor.Next(ctx) {
		err = cursor.Decode(&sellerPack)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if sellerPack.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Package history not found."), config.NOT_FOUND))
	}
	sellerPack.Id = primitive.NewObjectID() // onki maglumatlary kopyalap, kabirini uytgedip insert edyas!
	now := time.Now()
	to := sellerPack.To
	if now.Before(to) {
		max_days := to.Sub(now).Truncate(24*time.Hour).Hours() / 24
		if extendBy.Days > int64(max_days) {
			return c.JSON(errRes("extendBy.Days > max_days", fmt.Errorf("Number of days cannot exceed %v days.", max_days), config.NOT_ALLOWED))
		}
		sellerPack.To = to.AddDate(0, 0, -int(extendBy.Days))
		sellerPack.Action = "reduced"
		sellerPack.Text = fmt.Sprintf("Reduced by %v days.", extendBy.Days)
		sellerPack.CreatedAt = now

		transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
		tr_packageHistColl := transaction_manager.Collection(config.PACKAGEHISTORY)
		insert_model := ojoTr.NewModel().SetDocument(sellerPack)
		tr_insertResult, err := tr_packageHistColl.InsertOne(insert_model)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("InsertOne(package_history)", err, config.CANT_INSERT))
		}

		tr_sellersColl := transaction_manager.Collection(config.SELLERS)
		update_model := ojoTr.NewModel().
			SetFilter(bson.M{"_id": sellerObjId}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"package.to":                 sellerPack.To,
					"package.package_history_id": tr_insertResult.InsertedID.(primitive.ObjectID),
				},
			}).
			SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
				oldData := i.(bson.M)
				p := oldData["package"].(bson.M)

				return bson.M{
					"$set": bson.M{
						"package.to":                 p["to"],
						"package.package_history_id": p["package_history_id"],
					},
				}
			})
		_, err = tr_sellersColl.FindOneAndUpdate(update_model)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(seller)", err, config.CANT_UPDATE))
		}
		if err = transaction_manager.Err(); err != nil { // && err != mongo.ErrNoDocuments { bu kayarym bolup biler
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
		}

		return c.JSON(models.Response[models.SellerPackageHistoryFull]{
			IsSuccess: true,
			Result:    sellerPack,
		})
	}
	return c.JSON(errRes("now.After(to)", errors.New("Package already expired."), config.PACKAGE_EXPIRED))
}
