//GetCollectionsInfo
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
)

/*
	Collections Names in Collections Page
*/
func GetCollectionsInfo(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCollectionInfo")
	culture := helpers.GetCultureFromQuery(c)
	collections := config.MI.DB.Collection(config.COLLECTIONS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collections.Aggregate(ctx, bson.A{
		bson.M{
			"$addFields": bson.M{
				"name": culture.Stringf("$name.%v"),
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var collectionsResult = []models.Collection{}

	for cursor.Next(ctx) {
		var collection models.Collection
		err := cursor.Decode(&collection)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		collectionsResult = append(collectionsResult, collection)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Collection]{
		IsSuccess: true,
		Result:    collectionsResult,
	})
}

/*
	Collections page-de New Section-daky product card info
*/
func GetCollectionNewBrief(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCollectionNewBrief")
	products := config.MI.DB.Collection(config.PRODUCTS)
	culture := helpers.GetCultureFromQuery(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	today := time.Now()
	y, m, d := today.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"created_at": bson.M{
					"$gt": lastWeek,
				},
				"status": config.STATUS_PUBLISHED,
			},
		},
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$limit": 2,
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
				"heading": culture.Stringf("$heading.%v"),
				"image": bson.M{
					"$first": "$images",
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"images": 0,
			},
		},
	}

	cursor, err := products.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var productsResult = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err = cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		productsResult = append(productsResult, product)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    productsResult,
	})
}

/*
	Get all new products when see-all clicked
*/
func GetCollectionNewAll(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCollectionNewAll")
	products := config.MI.DB.Collection(config.PRODUCTS)
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	today := time.Now()
	y, m, d := today.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)

	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"created_at": bson.M{
					"$gt": lastWeek,
				},
				"status": config.STATUS_PUBLISHED,
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
				"heading": culture.Stringf("$heading.%v"),
				"image": bson.M{
					"$first": "$images",
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"images": 0,
			},
		},
	}

	cursor, err := products.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var productsResult = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err = cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		productsResult = append(productsResult, product)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    productsResult,
	})
}
func GetCollectionDiscountedBrief(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCollectionDiscountedBrief")
	products := config.MI.DB.Collection(config.PRODUCTS)
	culture := helpers.GetCultureFromQuery(c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	today := time.Now()
	y, m, d := today.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"discount": bson.M{
					"$gt": 0,
				},
				"status": config.STATUS_PUBLISHED,
			},
		},
		bson.M{
			"$sort": bson.M{
				"discount": -1,
			},
		},
		bson.M{
			"$limit": 2,
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
					"$gt": bson.A{
						"$created_at", lastWeek,
					},
				},
				"heading": culture.Stringf("$heading.%v"),
				"image": bson.M{
					"$first": "$images",
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"images": 0,
			},
		},
	}

	cursor, err := products.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var productsResult = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err = cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		productsResult = append(productsResult, product)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    productsResult,
	})
}
func GetCollectionDiscountedAll(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCollectionDiscountedAll")
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	today := time.Now()
	y, m, d := today.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)

	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"discount": bson.M{
					"$gt": 0,
				},
				"status": config.STATUS_PUBLISHED,
			},
		},
		bson.M{
			"$sort": bson.M{
				"discount": -1,
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
					"$gt": bson.A{
						"$created_at", lastWeek,
					},
				},
				"heading": culture.Stringf("$heading.%v"),
				"image": bson.M{
					"$first": "$images",
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"images": 0,
			},
		},
	}
	products := config.MI.DB.Collection(config.PRODUCTS)
	cursor, err := products.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var productsResult = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err = cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		productsResult = append(productsResult, product)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    productsResult,
	})
}
func GetCollectionTrendingBrief(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCollectionTrendingBrief")
	culture := helpers.GetCultureFromQuery(c)
	today := time.Now()
	y, m, d := today.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)
	last_month := yesterday.AddDate(0, -1, 0)

	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"created_at": bson.M{
					"$gt": last_month,
				},
				"status": config.STATUS_PUBLISHED,
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": "$product_id",
				"likes": bson.M{
					"$sum": 1,
				},
			},
		},
		bson.M{
			"$sort": bson.M{
				"likes": -1,
			},
		},
		bson.M{
			"$limit": 2,
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "products",
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "product",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$product",
			},
		},
		bson.M{
			"$replaceRoot": bson.M{
				"newRoot": "$product",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"is_new": bson.M{
					"$gt": bson.A{
						"$created_at", lastWeek,
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"heading":     culture.Stringf("$heading.%v"),
				"image":       bson.M{"$first": "$images"},
				"price":       1,
				"discount":    1,
				"is_favorite": 1,
				"is_new":      1,
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	favorite_products := config.MI.DB.Collection(config.FAV_PRODS)
	cursor, err := favorite_products.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	var products = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err = cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		products = append(products, product)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    products,
	})
}
func GetCollectionTrendingAll(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCollectionTrendingAll")
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))

	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	today := time.Now()
	y, m, d := today.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)
	lastMonth := yesterday.AddDate(0, -1, 0)

	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"created_at": bson.M{
					"$gt": lastMonth,
				},
				"status": config.STATUS_PUBLISHED,
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": "$product_id",
				"likes": bson.M{
					"$sum": 1,
				},
			},
		},
		bson.M{
			"$sort": bson.M{
				"likes": -1,
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
				"from":         "products",
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "product",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$product",
			},
		},
		bson.M{
			"$replaceRoot": bson.M{
				"newRoot": "$product",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"is_new": bson.M{
					"$gt": bson.A{"$created_at", lastWeek},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"heading":     culture.Stringf("$heading.%v"),
				"image":       bson.M{"$first": "$images"},
				"price":       1,
				"discount":    1,
				"is_favorite": 1,
				"is_new":      1,
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	favorite_products := config.MI.DB.Collection(config.FAV_PRODS)
	cursor, err := favorite_products.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))

	}
	defer cursor.Close(ctx)
	var products = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err = cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		products = append(products, product)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    products,
	})
}
