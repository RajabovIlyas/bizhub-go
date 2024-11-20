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

func GetBrands(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("GetBrands")
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.CANT_DECODE))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.CANT_DECODE))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	brandsColl := config.MI.DB.Collection(config.BRANDS)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"parent": nil,
			},
		},
		bson.M{"$sort": bson.M{"name": 1}},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
		bson.M{
			"$lookup": bson.M{
				"from":         "brands",
				"localField":   "_id",
				"foreignField": "parent",
				"as":           "sub_brands",
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
			"$project": bson.M{
				"name":       1,
				"logo":       1,
				"sub_brands": 1,
			},
		},
	}
	var brands = []models.BrandWithSubBrands{}
	cursor, err := brandsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var cat models.BrandWithSubBrands
		err = cursor.Decode(&cat)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		brands = append(brands, cat)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.BrandWithSubBrands]{
		IsSuccess: true,
		Result:    brands,
	})
}

func GetParentBrands(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("GetParentBrands")
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.CANT_DECODE))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.CANT_DECODE))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	brandsColl := config.MI.DB.Collection(config.BRANDS)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"parent": nil,
			},
		},
		bson.M{"$sort": bson.M{"name": 1}},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
		bson.M{
			"$project": bson.M{
				"name": 1,
				"logo": 1, // onki version da "parent": 0 -dy!
			},
		},
	}
	var brands []models.BrandParent
	cursor, err := brandsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var cat models.BrandParent
		err = cursor.Decode(&cat)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		brands = append(brands, cat)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if brands == nil {
		brands = make([]models.BrandParent, 0)
	}
	return c.JSON(models.Response[[]models.BrandParent]{
		IsSuccess: true,
		Result:    brands,
	})
}

func AddNewBrand(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.AddNewBrand")
	var newBrand models.NewBrand
	// parent=null ise ozi parent, diymek image hem bolmaly
	// parent!=null ise categories[].notEmpty bolmaly
	newBrand.Name = c.FormValue("name")
	newBrand.Categories = []primitive.ObjectID{}
	parent, err := primitive.ObjectIDFromHex(c.FormValue("parent"))
	if err == nil {
		newBrand.Parent = &parent
	}

	if newBrand.Parent == nil {
		logo, err := helpers.SaveImageFile(c, "logo", config.FOLDER_BRANDS)
		if err != nil {
			return c.JSON(errRes("SaveImageFile()", err, config.BODY_NOT_PROVIDED))
		}
		newBrand.Logo = &logo
	} else {
		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(errRes("MultipartForm()", err, config.CANT_DECODE))
		}
		categories := form.Value["categories"]
		for _, cat := range categories {
			catObjId, err := primitive.ObjectIDFromHex(cat)
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(category)", err, config.CANT_DECODE))
			}
			newBrand.Categories = append(newBrand.Categories, catObjId)
		}
		if len(newBrand.Categories) == 0 {
			return c.JSON(errRes("NoCategories", errors.New("Categories not provided."), config.BODY_NOT_PROVIDED))
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	brandsColl := config.MI.DB.Collection(config.BRANDS)
	insertResult, err := brandsColl.InsertOne(ctx, newBrand)
	if err != nil {
		if newBrand.Logo != nil {
			helpers.DeleteImageFile(*newBrand.Logo)
		}
		return c.JSON(errRes("InsertOne(brand)", err, config.CANT_INSERT))
	}
	newBrand.Id = insertResult.InsertedID.(primitive.ObjectID)
	return c.JSON(models.Response[models.NewBrand]{
		IsSuccess: true,
		Result:    newBrand,
	})
}
func EditBrand(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.EditBrand")
	brandObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var brand, oldBrand models.NewBrand
	brand.Id = brandObjId
	brand.Name = c.FormValue("name")
	if len(brand.Name) == 0 {
		return c.JSON(errRes("NameNotProvided", errors.New("Name not provided."), config.BODY_NOT_PROVIDED))
	}
	updateModel := bson.M{"name": brand.Name}
	parent, err := primitive.ObjectIDFromHex(c.FormValue("parent"))
	if err == nil {
		brand.Parent = &parent
		updateModel["parent"] = brand.Parent
	}
	imageChanged, err := strconv.ParseBool(c.FormValue("is_image_changed"))
	if err != nil {
		return c.JSON(errRes("ParseBool(is_image_changed)", err, config.BODY_NOT_PROVIDED))
	}
	if brand.Parent == nil {
		if imageChanged {
			logo, err := helpers.SaveImageFile(c, "logo", config.FOLDER_BRANDS)
			if err != nil {
				return c.JSON(errRes("SaveImageFile()", err, config.BODY_NOT_PROVIDED))
			}
			brand.Logo = &logo
			updateModel["logo"] = brand.Logo
		}
	} else {
		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(errRes("MultipartForm()", err, config.CANT_DECODE))
		}
		categories := form.Value["categories"]
		for _, cat := range categories {
			catObjId, err := primitive.ObjectIDFromHex(cat)
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(category)", err, config.CANT_DECODE))
			}
			brand.Categories = append(brand.Categories, catObjId)
		}
		if len(brand.Categories) == 0 {
			return c.JSON(errRes("NoCategories", errors.New("Categories not provided."), config.BODY_NOT_PROVIDED))
		} else {
			updateModel["categories"] = brand.Categories
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	brandsColl := config.MI.DB.Collection(config.BRANDS)
	updateResult := brandsColl.FindOneAndUpdate(ctx, bson.M{"_id": brandObjId}, bson.M{
		"$set": updateModel,
	})
	if err = updateResult.Err(); err != nil {
		if imageChanged {
			helpers.DeleteImageFile(*brand.Logo)
		}
		return c.JSON(errRes("FindOneAndUpdate()", err, config.CANT_UPDATE))
	}
	err = updateResult.Decode(&oldBrand)
	if err != nil {
		return c.JSON(errRes("SavedChanges BUT Decode()", err, config.CANT_DECODE))
	}
	if imageChanged {
		helpers.DeleteImageFile(*oldBrand.Logo)
	}
	return c.JSON(models.Response[models.NewBrand]{
		IsSuccess: true,
		Result:    brand,
	})
}
func GetBrand(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetBrand")
	var brandObjId primitive.ObjectID
	brandObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	brandsColl := config.MI.DB.Collection(config.BRANDS)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": brandObjId,
			},
		},
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
							"logo": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$parent",
				"preserveNullAndEmptyArrays": true,
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
						"$lookup": bson.M{
							"from":         "categories",
							"localField":   "parent",
							"foreignField": "_id",
							"as":           "parent",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"name":  "$name.en",
										"image": 1,
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path":                       "$parent",
							"preserveNullAndEmptyArrays": true,
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
	}
	cursor, err := brandsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var brand models.EditBrand
	if cursor.Next(ctx) {
		err = cursor.Decode(&brand)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if brand.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Brand not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[models.EditBrand]{
		IsSuccess: true,
		Result:    brand,
	})
}
