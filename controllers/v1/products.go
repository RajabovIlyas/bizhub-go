package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	ojocronlisteners "github.com/devzatruk/bizhubBackend/config/ojocron_listeners"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SetProductDiscount(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.SetProductDiscount")
	var discountData models.DiscountData
	err := c.BodyParser(&discountData)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	err = helpers.ValidateDiscountData(discountData)
	if err != nil {
		return c.JSON(errRes("ValidateDiscountData()", err, config.NOT_ALLOWED))
	}
	productObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var sellerObjId primitive.ObjectID
	err = helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.NO_PERMISSION))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	findResult := productsColl.FindOne(ctx, bson.M{"_id": productObjId})
	if err = findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(product)", err, config.NOT_FOUND))
	}
	var productDetail struct {
		Id       primitive.ObjectID `bson:"_id"`
		Price    float64            `bson:"price"`
		SellerId primitive.ObjectID `bson:"seller_id"`
	}
	err = findResult.Decode(&productDetail)
	if err != nil {
		return c.JSON(errRes("Decode(product)", err, config.CANT_DECODE))
	}
	if sellerObjId != productDetail.SellerId {
		return c.JSON(errRes("NotOwner", errors.New("Seller not the owner."), config.NOT_ALLOWED))
	}
	var discount float64
	if discountData.Type == config.DISCOUNT_PERCENT {
		// if discountData.Percent > 0 {
		discount = discountData.Percent
		// discountData.Price = productDetail.Price * (100.0 - discount) / 100.0 // satmaly bahasy
		discountData.Price = productDetail.Price * discount / 100.0 // indirim etmeli mukdary, bu satmaly bahasy dal!
	} else if discountData.Type == config.DISCOUNT_PRICE {
		discount = discountData.Price * 100.0 / productDetail.Price
		discountData.Percent = discount
	}

	_, err = productsColl.UpdateOne(ctx, bson.M{"_id": productObjId},
		bson.M{
			"$set": bson.M{
				"discount":      discountData.Percent,
				"discount_data": discountData,
			},
		})
	if err != nil {
		return c.JSON(errRes("UpdateOne(product)", err, config.CANT_UPDATE))
	}
	// calculate when to send auto posts and schedule
	err = ojocronlisteners.ScheduleAutoPostAddDiscountTimes(discountData.Duration, discountData.DurationType, productObjId)
	if err != nil {
		c.JSON(errRes("ScheduleAutoPostAddDiscountTimes()", err, config.CANT_INSERT))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    "Discount applied successfully.",
	})
}
func RemoveProductDiscount(c *fiber.Ctx) error {
	// TODO: transaction-a owursemmi?
	errRes := helpers.ErrorResponse("Mobile.RemoveProductDiscount")
	productObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var sellerObjId primitive.ObjectID
	err = helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.NO_PERMISSION))
	}
	log := config.ProductsV1Logger.Group(fmt.Sprintf("RemoveProductDiscount(productId: %v)", productObjId))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	// indi products collection-da product-y tapyp, discount = 0 etmeli, discountDetails =nil etmeli
	updateResult := productsColl.FindOneAndUpdate(ctx, bson.M{"_id": productObjId, "seller_id": sellerObjId},
		bson.M{
			"$set": bson.M{
				"discount":      0,
				"discount_data": nil,
			},
		})
	if err = updateResult.Err(); err != nil {
		log.Errorf("Update product discount failed.")
		return c.JSON(errRes("FindOneAndUpdate()", err, config.NOT_FOUND))
	}
	log.Logf("Product discount removed.")

	pResult, err := productsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": productObjId,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "sellers",
				"foreignField": "_id",
				"localField":   "seller_id",
				"as":           "seller",
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
				"path": "$seller",
			},
		},
		bson.M{
			"$project": bson.M{
				"image": bson.M{
					"$first": "$images",
				},
				"seller_id":   1,
				"heading":     1,
				"seller_name": "$seller.name",
				"discount":    1,
			},
		},
	})

	if err != nil {
		log.Errorf("Aggregate error: %v", err)
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}

	var product struct {
		SellerId   primitive.ObjectID `bson:"seller_id"`
		Image      string             `bson:"image"`
		Heading    models.Translation `bson:"heading"`
		SellerName string             `bson:"seller_name"`
		Discount   float64            `bson:"discount"`
	}
	if pResult.Next(ctx) {
		err = pResult.Decode(&product)
		if err != nil {
			log.Errorf("Decode error: %v", err)
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}

	if err = pResult.Err(); err != nil {
		log.Errorf("Error: %v", err)
		return c.JSON(errRes("Err()", err, config.DBQUERY_ERROR))
	}

	heading := models.Translation{
		En: fmt.Sprintf(os.Getenv("AnnounceEnRem"), product.SellerName,
			product.Discount, product.Heading.En),
		Tm: fmt.Sprintf(os.Getenv("AnnounceTmRem"), product.SellerName,
			product.Heading.Tm, product.Discount),
		Ru: fmt.Sprintf(os.Getenv("AnnounceRuRem"), product.SellerName,
			product.Discount, product.Heading.Ru),
		Tr: fmt.Sprintf(os.Getenv("AnnounceTrRem"), product.SellerName,
			product.Heading.Tr, product.Discount),
	}

	post := models.PostUpsert{
		Image:           product.Image,
		SellerId:        product.SellerId,
		Title:           heading,
		Body:            models.Translation{Tm: "", Ru: "", En: "", Tr: ""},
		RelatedProducts: []primitive.ObjectID{productObjId},
		Viewed:          0,
		Likes:           0,
		Auto:            true,
	}
	postsColl := config.MI.DB.Collection(config.POSTS)
	result, err := postsColl.InsertOne(ctx, post)
	if err != nil {
		log.Errorf("Create autopost error: %v", err)
		return c.JSON(errRes("InsertOne(auto_post)", err, config.CANT_INSERT))
	}
	// gercek posts collection-a gosmaly
	log.Logf("AutoPost - post published -> %v", result.InsertedID)

	_ = config.OjoCronService.RemoveJobsByGroup(productObjId)
	// if err != nil {
	// 	log.Errorf("Couldn't delete from cron jobs...")
	// 	return c.JSON(errRes("RemoveJobsByGroup(cron_job)", err, config.CANT_DELETE))
	// }
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    "Discount removed successfully.",
	})
}
func GetProductDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetProductDetail")
	objId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}

	culture := helpers.GetCultureFromQuery(c)
	products := config.MI.DB.Collection(config.PRODUCTS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = products.UpdateOne(ctx, bson.M{"_id": objId}, bson.M{
		"$inc": bson.M{
			"viewed": 1,
		},
	})
	if err != nil {
		return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
	}

	// aggregationArray := bson.A{

	// 	bson.M{
	// 		"$match": bson.M{
	// 			"_id": objId,
	// 		},
	// 	},
	// 	bson.M{
	// 		"$lookup": bson.M{
	// 			"from":         "sellers",
	// 			"localField":   "seller_id",
	// 			"foreignField": "_id",
	// 			"as":           "seller",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$unwind": bson.M{
	// 			"path": "$seller",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$lookup": bson.M{
	// 			"from":         "cities",
	// 			"localField":   "seller.city_id",
	// 			"foreignField": "_id",
	// 			"as":           "seller.city",
	// 		},
	// 	},
	// 	bson.M{
	// 		"$unwind": bson.M{
	// 			"path": "$seller.city",
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
	// 			"heading":              fmt.Sprintf("$heading.%v", culture.Lang),
	// 			"more_details":         fmt.Sprintf("$more_details.%v", culture.Lang),
	// 			"seller.city.name":     fmt.Sprintf("$seller.city.name.%v", culture.Lang),
	// 			"category.name":        fmt.Sprintf("$category.name.%v", culture.Lang),
	// 			"category.parent.name": fmt.Sprintf("$category.parent.name.%v", culture.Lang),
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
	// 			"seller.address":             0,
	// 			"seller.bio":                 0,
	// 			"seller.owner_id":            0,
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
	// 					"$addFields": bson.M{
	// 						"attrs.attr.name": fmt.Sprintf("$attrs.attr.name.%v", culture.Lang),
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
				"_id": objId,
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
										"name": culture.Stringf("$name.%v"),
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
							"name":        culture.Stringf("$name.%v"),
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
										"name": culture.Stringf("$name.%v"),
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
							"name":   culture.Stringf("$name.%v"),
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
				"heading":       culture.Stringf("$heading.%v"),
				"more_details":  culture.Stringf("$more_details.%v"),
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
	var productDetail models.ProductDetail
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
	return c.JSON(models.Response[models.ProductDetail]{
		IsSuccess: true,
		Result:    productDetail,
	})

}

func SearchProduct(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.SearchProduct")
	culture := helpers.GetCultureFromQuery(c)
	searchQuery := c.Query("q")
	if len(searchQuery) == 0 {
		return c.JSON(errRes("Query(q)", errors.New("Search query parameter not provided."), config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}

	productsCollection := config.MI.DB.Collection(config.PRODUCTS)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	currentDate := time.Now()
	y, m, d := currentDate.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)
	aggregationArray := bson.A{
		bson.M{
			"$facet": bson.M{
				"byHeading": bson.A{
					bson.M{
						"$match": bson.M{
							fmt.Sprintf("heading.%v", culture.Lang): primitive.Regex{Pattern: fmt.Sprintf("(%v)", searchQuery), Options: "gi"},
						},
					},
				},
				"byMoreDetails": bson.A{
					bson.M{
						"$match": bson.M{
							fmt.Sprintf("more_details.%v", culture.Lang): primitive.Regex{Pattern: fmt.Sprintf("(%v)", searchQuery), Options: "gi"},
						},
					},
				},
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
				"from": "categories",
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							fmt.Sprintf("name.%v", culture.Lang): primitive.Regex{Pattern: fmt.Sprintf("(%v)", searchQuery), Options: "gi"},
						},
					},
					bson.M{
						"$project": bson.M{
							"_id": 1,
						},
					},
					bson.M{
						"$lookup": bson.M{
							"from":         "products",
							"localField":   "_id",
							"foreignField": "category_id",
							"as":           "products",
						},
					},
				},
				"as": "category_results",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"category_results": bson.M{
					"$reduce": bson.M{
						"input": bson.M{
							"$map": bson.M{
								"input": "$category_results",
								"as":    "cat",
								"in":    "$$cat.products",
							},
						},
						"initialValue": bson.A{},
						"in": bson.M{
							"$concatArrays": bson.A{"$$this", "$$value"},
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"result": bson.M{
					"$concatArrays": bson.A{"$byHeading", "$byMoreDetails", "$category_results"},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$result",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id":    "$result._id",
				"result": bson.M{"$first": "$result"},
			},
		},
		bson.M{
			"$replaceRoot": bson.M{
				"newRoot": "$result",
			},
		},
		bson.M{
			"$sort": bson.M{
				fmt.Sprintf("heading.%v", culture.Lang): -1,
			},
		},
		// bson.M{
		// 	"$skip": pageIndex * limit,
		// },
		// bson.M{
		// 	"$limit": limit,
		// }, // TODO: su $skip we $limit yokary gecirdim, dogrumy?
		bson.M{
			"$project": bson.M{
				"image": bson.M{
					"$first": "$images",
				},
				"heading":  fmt.Sprintf("$heading.%v", culture.Lang),
				"price":    1,
				"discount": 1,
				"viewed":   1,
				"is_new": bson.M{
					"$gt": bson.A{"$created_at", lastWeek},
				},
			},
		},
	}

	cursor, err := productsCollection.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var products = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err := cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		products = append(products, product)
	}
	if err := cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    products,
	})
}

func FilterProduct(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.FilterProduct")
	culture := helpers.GetCultureFromQuery(c)
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}

	categoriesQuery := c.Query("categories", "all")
	brandsQuery := c.Query("brands", "all")
	sellersQuery := c.Query("sellers", "all") // ["_id"] sekilde
	citiesQuery := c.Query("cities", "all")
	priceQuery := c.Query("price", "none")
	sortQuery := c.Query("sort", "none")

	match := bson.M{}
	aggregationArray := bson.A{}

	if categoriesQuery != "all" {
		var bsonCategories = bson.A{}
		var categories = []string{}
		err := json.Unmarshal([]byte(categoriesQuery), &categories)
		if err != nil {
			return c.JSON(errRes("json.Unmarshal(categories)", err, config.CANT_DECODE))
		}

		for i := 0; i < len(categories); i++ {
			catId, err := primitive.ObjectIDFromHex(categories[i])
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(categories)", err, config.CANT_DECODE))
			}
			bsonCategories = append(bsonCategories, catId)
		}
		if len(categories) != 0 {
			match["category_id"] = bson.M{
				"$in": bsonCategories,
			}
		}
	}

	if brandsQuery != "all" {
		var bsonBrands = bson.A{}
		var brands = []string{}
		err := json.Unmarshal([]byte(brandsQuery), &brands)
		if err != nil {
			return c.JSON(errRes("json.Unmarshal(brands)", err, config.CANT_DECODE))
		}

		for i := 0; i < len(brands); i++ {
			catId, err := primitive.ObjectIDFromHex(brands[i])
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(brands)", err, config.CANT_DECODE))
			}
			bsonBrands = append(bsonBrands, catId)
		}
		if len(brands) != 0 {
			match["brand_id"] = bson.M{
				"$in": bsonBrands,
			}
		}
	}

	if sellersQuery != "all" {
		var bsonSellers = bson.A{}
		var sellers = []string{}
		err := json.Unmarshal([]byte(sellersQuery), &sellers)
		if err != nil {
			return c.JSON(errRes("json.Unmarshal(sellers)", err, config.CANT_DECODE))
		}

		for i := 0; i < len(sellers); i++ {
			catId, err := primitive.ObjectIDFromHex(sellers[i])
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(sellers)", err, config.CANT_DECODE))
			}
			bsonSellers = append(bsonSellers, catId)
		}
		if len(sellers) != 0 {
			match["seller_id"] = bson.M{
				"$in": bsonSellers,
			}
		}
	}

	if citiesQuery != "all" {
		var bsonCities = bson.A{}
		var cities = []string{}
		err := json.Unmarshal([]byte(citiesQuery), &cities)
		if err != nil {
			return c.JSON(errRes("json.Unmarshal(cities)", err, config.CANT_DECODE))
		}

		for i := 0; i < len(cities); i++ {
			catId, err := primitive.ObjectIDFromHex(cities[i])
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(cities)", err, config.CANT_DECODE))
			}
			bsonCities = append(bsonCities, catId)
		}
		if len(cities) != 0 {
			match["seller.city_id"] = bson.M{
				"$in": bsonCities,
			}
			aggregationArray = append(aggregationArray, bson.M{
				"$lookup": bson.M{
					"from":         "sellers",
					"localField":   "seller_id",
					"foreignField": "_id",
					"as":           "seller",
				},
			}, bson.M{
				"$unwind": bson.M{
					"path": "$seller",
				},
			})
		}
	}

	if priceQuery != "none" {
		var price = []any{}
		err := json.Unmarshal([]byte(priceQuery), &price)
		if err != nil {
			return c.JSON(errRes("json.Unmarshal(price)", err, config.CANT_DECODE))
		}
		// fmt.Println(price)
		if len(price) == 2 {
			_p := bson.M{}
			if price[0] != "none" {
				_p["$gte"] = price[0]
			}
			if price[1] != "none" {
				_p["$lte"] = price[1]
			}
			if len(_p) != 0 {
				match["price"] = _p
			}
		}
	}
	if len(match) != 0 {
		aggregationArray = append(aggregationArray, bson.M{"$match": match})
	}
	productsCollection := config.MI.DB.Collection(config.PRODUCTS)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	currentDate := time.Now()
	y, m, d := currentDate.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)

	aggregationArray = append(aggregationArray, bson.A{
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"images":     1,
				"heading":    1,
				"price":      1,
				"discount":   1,
				"created_at": 1,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"is_new": bson.M{
					"$gt": bson.A{"$created_at", lastWeek},
				},
				"heading": fmt.Sprintf("$heading.%v", culture.Lang),
				"image": bson.M{
					"$first": "$images",
				},
			},
		},
	}...)

	if sortQuery != "none" {
		index, err := strconv.Atoi(sortQuery)
		if err != nil {
			return c.JSON(errRes("Atoi(sortQuery)", err, config.CANT_DECODE))
		}
		if index == 0 { // Price: Low to High
			aggregationArray = append(aggregationArray, bson.M{
				"$sort": bson.M{
					"price": 1,
				},
			})
		} else if index == 1 { // Price: High to Low
			aggregationArray = append(aggregationArray, bson.M{
				"$sort": bson.M{
					"price": -1,
				},
			})
		} else if index == 2 { // New Products
			aggregationArray = append(aggregationArray, bson.M{
				"$sort": bson.M{
					"is_new": -1,
				},
			})
		} else if index == 3 { // Trending
			// TODO: su yerini yazmandyrys!!!
		} else if index == 4 { // Discount
			aggregationArray = append(aggregationArray, bson.M{
				"$sort": bson.M{
					"discount": -1,
				},
			})
		}
	}

	cursor, err := productsCollection.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var products = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err := cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		products = append(products, product)
	}
	if err := cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    products,
	})
}
