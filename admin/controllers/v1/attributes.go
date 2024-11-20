package v1

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetAttributes(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetAttributes")
	queryString := c.Query("q")
	match := bson.M{}
	if len(queryString) > 0 {
		match["$match"] = bson.M{
			"name.en": primitive.Regex{
				Pattern: queryString,
				Options: "i",
			},
		}
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
	attributesColl := config.MI.DB.Collection(config.ATTRIBUTES)
	aggregationArray := bson.A{}
	if len(queryString) > 0 {
		aggregationArray = append(aggregationArray, match)
	}
	aggregationArray = append(aggregationArray,
		bson.A{
			bson.M{
				"$project": bson.M{
					"name": "$name.en",
				},
			},
			bson.M{"$skip": pageIndex * limit},
			bson.M{"$limit": limit},
		}...)

	var attributes []models.AttributeName
	cursor, err := attributesColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var attr models.AttributeName
		err = cursor.Decode(&attr)
		if err != nil {
			return c.JSON(errRes("Decode(attribute)", err, config.CANT_DECODE))
		}
		attributes = append(attributes, attr)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.CANT_DECODE))
	}
	if attributes == nil {
		attributes = make([]models.AttributeName, 0)
	}
	return c.JSON(models.Response[[]models.AttributeName]{
		IsSuccess: true,
		Result:    attributes,
	})
}
func GetAttribute(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetAttribute")
	attrObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	attributesColl := config.MI.DB.Collection(config.ATTRIBUTES)
	findResult := attributesColl.FindOne(ctx, bson.M{"_id": attrObjId})
	if err = findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne()", err, config.NOT_FOUND))
	}
	var attr models.NewAttribute
	err = findResult.Decode(&attr)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	attr.Units = strings.Join(attr.UnitsArray, ", ")
	return c.JSON(models.Response[models.NewAttribute]{
		IsSuccess: true,
		Result:    attr,
	})
}
func EditAttribute(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.EditAttribute")
	attrObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var attr models.NewAttribute
	err = c.BodyParser(&attr)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	attr.UnitsArray = make([]string, 0)
	splitUnits := strings.Split(attr.Units, ",")
	for _, u := range splitUnits {
		unit := strings.TrimSpace(u)
		if len(unit) != 0 {
			attr.UnitsArray = append(attr.UnitsArray, unit)
		}
	}
	if attr.HasEmptyFields() {
		return c.JSON(errRes("HasEmptyFields()", errors.New("Some data not provided."), config.BODY_NOT_PROVIDED))
	}
	attr.Id = attrObjId
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	attributesColl := config.MI.DB.Collection(config.ATTRIBUTES)
	updateResult, err := attributesColl.UpdateOne(ctx, bson.M{"_id": attr.Id}, bson.M{
		"$set": bson.M{
			"name":        attr.Name,
			"units_array": attr.UnitsArray,
			"is_number":   attr.IsNumber,
		},
	})
	if err != nil {
		return c.JSON(errRes("UpdateOne(attribute)", err, config.CANT_UPDATE))
	}
	if updateResult.MatchedCount == 0 {
		return c.JSON(errRes("UpdateOne(attribute)", errors.New("Attribute not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.UPDATED,
	})
}
func AddNewAttribute(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.AddNewAttribute")
	var newAttr models.NewAttribute
	err := c.BodyParser(&newAttr)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	newAttr.UnitsArray = make([]string, 0)
	splitUnits := strings.Split(newAttr.Units, ",")
	for _, u := range splitUnits {
		unit := strings.TrimSpace(u)
		if len(unit) != 0 {
			newAttr.UnitsArray = append(newAttr.UnitsArray, unit)
		}
	}
	if newAttr.HasEmptyFields() {
		return c.JSON(errRes("HasEmptyFields()", errors.New("Some data not provided."), config.BODY_NOT_PROVIDED))
	}
	var employeeObjId primitive.ObjectID
	err = helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	attributesColl := config.MI.DB.Collection(config.ATTRIBUTES)
	insertResult, err := attributesColl.InsertOne(ctx, newAttr)
	if err != nil {
		return c.JSON(errRes("InsertOne(attribute)", err, config.CANT_INSERT))
	}
	newAttr.Id = insertResult.InsertedID.(primitive.ObjectID)
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.CREATED,
	})
}
