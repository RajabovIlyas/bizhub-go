package v1

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	notificationmanager "github.com/devzatruk/bizhubBackend/notification_manager"
	ojoTr "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetNotifications(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetNotifications")
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
	// filter := c.Query("filter", "all")
	// audience := c.Query("audience", "none")
	// aud := strings.Split(audience, ",")
	// aggregationArray := bson.A{}
	// if filter == "sent" {
	// 	aggregationArray = append(aggregationArray, bson.M{
	// 		"$match": bson.M{
	// 			"is_confirmed": true,
	// 		},
	// 	})
	// } else if filter == "notsent" {
	// 	aggregationArray = append(aggregationArray, bson.M{
	// 		"$match": bson.M{
	// 			"is_confirmed": false,
	// 		},
	// 	})
	// }
	// if audience != "none" {
	// 	if helpers.SliceContains(aud, "all") {
	// 		aggregationArray = append(aggregationArray, bson.M{
	// 			"$match": bson.M{
	// 				"audience.all": true,
	// 			},
	// 		})
	// 	} else {
	// 		aggregationArray = append(aggregationArray, bson.M{
	// 			"$match": bson.M{
	// 				"audience.users":   helpers.SliceContains(aud, "users"),
	// 				"audience.sellers": helpers.SliceContains(aud, "sellers"),
	// 			},
	// 		})
	// 	}
	// }
	// aggregationArray = append(aggregationArray, bson.A{
	// 	bson.M{
	// 		"$skip": pageIndex * limit,
	// 	},
	// 	bson.M{
	// 		"$limit": limit,
	// 	},
	// }...)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	notificationsColl := config.MI.DB.Collection(config.NOTIFICATIONS)
	cursor, err := notificationsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var notifications []models.Notification
	for cursor.Next(ctx) {
		var notif models.Notification
		err = cursor.Decode(&notif)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		notifications = append(notifications, notif)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if notifications == nil {
		notifications = make([]models.Notification, 0)
	}
	return c.JSON(models.Response[[]models.Notification]{
		IsSuccess: true,
		Result:    notifications,
	})
}
func NewNotification(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.NewNotification")
	var notifBody struct {
		Audience models.Audience `json:"audience" bson:"audience"`
		Text     string          `json:"text" bson:"text"`
	}
	err := c.BodyParser(&notifBody)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	var employeeObjId primitive.ObjectID
	err = helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	notif := models.Notification{
		Audience: notifBody.Audience,
		// Text: models.Translation{
		// 	En: notifBody.Text,
		// },
		Text:        notifBody.Text,
		CreatedBy:   employeeObjId,
		CreatedAt:   time.Now(),
		IsConfirmed: false,
		CheckedBy:   nil,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_notificationsColl := transaction_manager.Collection(config.NOTIFICATIONS)
	insertResult, err := tr_notificationsColl.InsertOne(ojoTr.NewModel().SetDocument(notif))
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("InsertOne(notification)", err, config.CANT_INSERT))
	}
	insertedID := insertResult.InsertedID.(primitive.ObjectID)
	// new_task := taskmanager.NewTask(config.MI.DB, 3)
	// new_task.Notification(insertedID, notif.Text.En)
	// err = new_task.Commit()
	// if err != nil {
	// 	trErr := transaction_manager.Rollback()
	// 	if trErr != nil {
	// 		err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
	// 	}
	// 	return c.JSON(errRes("InsertOne(task)", err, config.CANT_INSERT))
	// }
	if err = transaction_manager.Err(); err != nil { // i think it is unreachable code!
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	notif.Id = insertedID
	config.NotificationManager.AddNotificationEvent(&notificationmanager.NotificationEvent{
		Title:       "Bizhub",
		Description: notif.Text,
		ClientType: notificationmanager.NotificationEventClientType{
			All:       notif.Audience.All,
			Customers: notif.Audience.Users,
			Sellers:   notif.Audience.Sellers,
		},
	})
	return c.JSON(models.Response[models.Notification]{
		IsSuccess: true,
		Result:    notif,
	})
}

// func CheckNotification(c *fiber.Ctx) error {
// 	errRes := helpers.ErrorResponse("Admin.CheckNotification")
// 	var notifObjId primitive.ObjectID
// 	notifObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
// 	if err != nil {
// 		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
// 	}
// 	var notifText models.Translation
// 	err = c.BodyParser(&notifText)
// 	if err != nil {
// 		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
// 	}
// 	if notifText.HasEmptyFields() {
// 		return c.JSON(errRes("HasEmptyFields()", errors.New("Some data not provided."), config.BODY_NOT_PROVIDED))
// 	}
// 	// var employeeObjId primitive.ObjectID
// 	// err = helpers.GetCurrentEmployee(c, &employeeObjId)
// 	// if err != nil {
// 	// 	return c.JSON(errRes("GetCurrentEmployee()", err, config.NO_PERMISSION))
// 	// }
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
// 	tr_notificationsColl := transaction_manager.Collection(config.NOTIFICATIONS)
// 	update_model := ojoTr.NewModel().
// 		SetFilter(bson.M{"_id": notifObjId}).
// 		SetUpdate(bson.M{
// 			"$set": bson.M{
// 				"text": notifText,
// 				// "checked_by": employeeObjId,
// 				"is_confirmed": true,
// 			},
// 		}).SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
// 		oldData := i.(bson.M)
// 		return bson.M{
// 			"$set": bson.M{
// 				"text":         oldData["text"],
// 				"checked_by":   nil,
// 				"is_confirmed": false,
// 			},
// 		}
// 	})
// 	_, err = tr_notificationsColl.FindOneAndUpdate(update_model)
// 	if err != nil {
// 		trErr := transaction_manager.Rollback()
// 		if trErr != nil {
// 			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
// 		}
// 		return c.JSON(errRes("FindOneAndUpdate(notification)", err, config.CANT_UPDATE))
// 	}
// 	//TODO: tasks listesinden pozyan service

// 	if err = transaction_manager.Err(); err != nil { // i think it is unreachable code!
// 		fmt.Printf("\nInside transaction_manager.Err()\n")
// 		trErr := transaction_manager.Rollback()
// 		if trErr != nil {
// 			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
// 		}
// 		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
// 	}
// 	// TODO: everydayWorkService....
// 	// TODO: notificationService.Send()
// 	fmt.Printf("\nnotification text: %v\n", notifText)
// 	return c.JSON(models.Response[string]{
// 		IsSuccess: true,
// 		Result:    config.TRANSACTION_SUCCESSFUL,
// 	})
// }

// {
//     "$match": {
//       "is_confirmed": true,
//       "audience.all": false,
//       "audience.users": true,
//       "audience.sellers": true
//     }
//   }
