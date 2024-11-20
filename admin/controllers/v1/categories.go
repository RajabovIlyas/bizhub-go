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

func AddNewCategory(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.AddNewCategory")
	var newCat models.NewCategory
	// parent=null ise ozi parent, diymek image hem bolmaly
	// parent!=null ise attributes[].notEmpty bolmaly
	newCat.Name.Tm = c.FormValue("name.tm")
	newCat.Name.Tr = c.FormValue("name.tr")
	newCat.Name.Ru = c.FormValue("name.ru")
	newCat.Name.En = c.FormValue("name.en")
	newCat.Attributes = []primitive.ObjectID{}
	parent, err := primitive.ObjectIDFromHex(c.FormValue("parent"))
	if err == nil {
		newCat.Parent = &parent
	}

	if newCat.Parent == nil {
		image, err := helpers.SaveImageFile(c, "image", config.FOLDER_CATEGORIES)
		if err != nil {
			return c.JSON(errRes("SaveImageFile()", err, config.BODY_NOT_PROVIDED))
		}
		newCat.Image = &image
	} else {
		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(errRes("MultipartForm()", err, config.CANT_DECODE))
		}
		attributes := form.Value["attributes"]
		for _, attr := range attributes {
			attrObjId, err := primitive.ObjectIDFromHex(attr)
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(attribute)", err, config.CANT_DECODE))
			}
			newCat.Attributes = append(newCat.Attributes, attrObjId)
		}
		if len(newCat.Attributes) == 0 {
			return c.JSON(errRes("NoAttributes", errors.New("Attributes not provided."), config.BODY_NOT_PROVIDED))
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	categoriesColl := config.MI.DB.Collection(config.CATEGORIES)
	insertResult, err := categoriesColl.InsertOne(ctx, newCat)
	if err != nil {
		if newCat.Image != nil {
			helpers.DeleteImageFile(*newCat.Image)
		}
		return c.JSON(errRes("InsertOne(category)", err, config.CANT_INSERT))
	}
	newCat.Id = insertResult.InsertedID.(primitive.ObjectID)
	return c.JSON(models.Response[models.NewCategory]{
		IsSuccess: true,
		Result:    newCat,
	})
}
func EditCategory(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.EditCategory")
	catObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var newCat, oldCat models.NewCategory
	// parent=null ise ozi parent, diymek image hem bolmaly
	// parent!=null ise attributes[].notEmpty bolmaly
	newCat.Id = catObjId
	newCat.Name.Tm = c.FormValue("name.tm")
	newCat.Name.Tr = c.FormValue("name.tr")
	newCat.Name.Ru = c.FormValue("name.ru")
	newCat.Name.En = c.FormValue("name.en")
	if len(newCat.Name.En) == 0 || len(newCat.Name.Ru) == 0 || len(newCat.Name.Tm) == 0 || len(newCat.Name.Tr) == 0 {
		return c.JSON(errRes("HasEmptyFields", errors.New("Some fields not provided."), config.BODY_NOT_PROVIDED))
	}
	newCat.Attributes = []primitive.ObjectID{}
	updateModel := bson.M{"name": newCat.Name}
	parent, err := primitive.ObjectIDFromHex(c.FormValue("parent"))
	if err == nil {
		newCat.Parent = &parent
		updateModel["parent"] = newCat.Parent
	}
	imageChanged, err := strconv.ParseBool(c.FormValue("is_image_changed"))
	if err != nil {
		return c.JSON(errRes("ParseBool()", err, config.BODY_NOT_PROVIDED))
	}
	if newCat.Parent == nil {
		if imageChanged {
			image, err := helpers.SaveImageFile(c, "image", config.FOLDER_CATEGORIES)
			if err != nil {
				return c.JSON(errRes("SaveImageFile()", err, config.BODY_NOT_PROVIDED))
			}
			newCat.Image = &image
			updateModel["image"] = newCat.Image
		}
	} else {
		form, err := c.MultipartForm()
		if err != nil {
			return c.JSON(errRes("MultipartForm()", err, config.CANT_DECODE))
		}
		attributes := form.Value["attributes"]
		for _, attr := range attributes {
			attrObjId, err := primitive.ObjectIDFromHex(attr)
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(attribute)", err, config.CANT_DECODE))
			}
			newCat.Attributes = append(newCat.Attributes, attrObjId)
		}
		if len(newCat.Attributes) == 0 {
			return c.JSON(errRes("NoAttributes", errors.New("Attributes not provided."), config.BODY_NOT_PROVIDED))
		} else {
			updateModel["attributes"] = newCat.Attributes
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	categoriesColl := config.MI.DB.Collection(config.CATEGORIES)
	updateResult := categoriesColl.FindOneAndUpdate(ctx, bson.M{"_id": catObjId}, bson.M{
		"$set": updateModel,
	})
	if err = updateResult.Err(); err != nil {
		if imageChanged {
			helpers.DeleteImageFile(*newCat.Image)
		}
		return c.JSON(errRes("InsertOne(category)", err, config.CANT_INSERT))
	}
	err = updateResult.Decode(&oldCat)
	if err != nil {
		// if imageChanged { // eger db-de update edilen bolsa, taze surat pozulmasyn.
		// 	helpers.DeleteImageFile(*newCat.Image)
		// }
		return c.JSON(errRes("Decode(old_category)", err, config.CANT_DECODE))
	}
	if newCat.Parent == nil && imageChanged { // eger update edilen bolsa, kone suraty pozay!
		helpers.DeleteImageFile(*oldCat.Image)
	}
	return c.JSON(models.Response[models.NewCategory]{
		IsSuccess: true,
		Result:    newCat,
	})
}

func GetCategories(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("GetCategories")
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
	categoriesColl := config.MI.DB.Collection(config.CATEGORIES)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"parent": nil,
			},
		},
		bson.M{"$sort": bson.M{"name.en": 1}},
		bson.M{
			"$skip": limit * pageIndex,
		},
		bson.M{
			"$limit": limit,
		},

		bson.M{
			"$lookup": bson.M{
				"from":         "categories",
				"localField":   "_id",
				"foreignField": "parent",
				"as":           "sub_categories",
				"pipeline": bson.A{
					bson.M{
						"$sort": bson.M{
							"order": 1,
						},
					},
					bson.M{
						"$project": bson.M{
							"name": "$name.en",
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"name":           "$name.en",
				"image":          1,
				"sub_categories": 1,
			},
		},
	}
	var categories = []models.CategoryWithSubCats{}
	cursor, err := categoriesColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var cat models.CategoryWithSubCats
		err = cursor.Decode(&cat)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		categories = append(categories, cat)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.CategoryWithSubCats]{
		IsSuccess: true,
		Result:    categories,
	})
}
func GetCategory(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetCategory")
	catObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	categoriesColl := config.MI.DB.Collection(config.CATEGORIES)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": catObjId,
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
						"$sort": bson.M{
							"order": 1,
						},
					},
					bson.M{
						"$project": bson.M{
							"name": "$name.en",
						},
					},
				},
			},
		},
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
	}
	cursor, err := categoriesColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var cat models.EditCategory
	if cursor.Next(ctx) {
		err = cursor.Decode(&cat)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}

	if cat.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Category not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[models.EditCategory]{
		IsSuccess: true,
		Result:    cat,
	})
}
