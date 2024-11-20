package v1

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	notificationmanager "github.com/devzatruk/bizhubBackend/notification_manager"
	"github.com/devzatruk/bizhubBackend/ojocronservice"
	ojoTr "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetTasks(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetTasks")
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
	tasksColl := config.MI.DB.Collection(config.TASKS)
	aggregateTasks := bson.A{
		bson.M{"$sort": bson.M{"is_urgent": -1}},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
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
							"path":                       "$city",
							"preserveNullAndEmptyArrays": true,
						},
					},
					bson.M{
						"$project": bson.M{
							"name": 1,
							"type": 1,
							"city": bson.M{
								"$ifNull": bson.A{"$city", nil},
							},
							"logo": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$seller",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$project": bson.M{
				"description": 1,
				"target_id":   1,
				"type":        1,
				"is_urgent":   1,
				"seller": bson.M{
					"$ifNull": bson.A{"$seller", nil},
				},
			},
		},
	}
	cursor, err := tasksColl.Aggregate(ctx, aggregateTasks)
	if err != nil {
		return c.JSON(errRes("Aggregate(Tasks)", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var tasks []models.NewTask
	for cursor.Next(ctx) {
		var t models.NewTask
		err = cursor.Decode(&t)
		if err != nil {
			return c.JSON(errRes("Decode(Task)", err, config.CANT_DECODE))
		}
		tasks = append(tasks, t)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err(Task)", err, config.DBQUERY_ERROR))
	}
	if len(tasks) == 0 {
		tasks = make([]models.NewTask, 0)
	}
	return c.JSON(models.Response[[]models.NewTask]{
		IsSuccess: true,
		Result:    tasks,
	})
}
func GetTaskDetail(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetTaskDetail")
	taskId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tasksColl := config.MI.DB.Collection(config.TASKS)
	var task models.Task
	findResult := tasksColl.FindOne(ctx, bson.M{"_id": taskId})
	if err = findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(Task)", err, config.NOT_FOUND))
	}
	err = findResult.Decode(&task)
	if err != nil {
		return c.JSON(errRes("Decode(Task)", err, config.CANT_DECODE))
	}
	if task.Id.IsZero() {
		return c.JSON(errRes("NilObjectID", errors.New("Task not found."), config.NOT_FOUND))
	}
	var selectedModel interface{}
	var collectionName string
	var aggregationArray bson.A
	switch task.Type {
	case config.TASK_AUCTION:
		collectionName = config.AUCTIONS
		selectedModel = new(models.AuctionForAdminChecker)
		aggregationArray = bson.A{

			bson.M{
				"$match": bson.M{
					"_id": task.TargetId,
				},
			},
			bson.M{
				"$project": bson.M{
					"heading":     1,
					"description": 1,
				},
			},
		}
	case config.TASK_POST:
		collectionName = config.POSTS
		selectedModel = new(models.PostForAdminChecker)
		aggregationArray = bson.A{
			bson.M{
				"$match": bson.M{
					"_id": task.TargetId,
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
										"$addFields": bson.M{
											"name": "$name.en",
										},
									},
								},
							},
						},
						bson.M{
							"$unwind": bson.M{
								"path":                       "$city",
								"preserveNullAndEmptyArrays": true,
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
				"$project": bson.M{
					"seller":           1,
					"title":            1,
					"body":             1,
					"image":            1,
					"related_products": 1,
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
								"image": bson.M{
									"$first": "$images",
								},
							},
						},
					},
				},
			},
		}
	case config.TASK_NOTIFICATION:
		collectionName = config.NOTIFICATIONS
		selectedModel = new(models.NotificationForAdminChecker)
		aggregationArray = bson.A{
			bson.M{
				"$match": bson.M{
					"_id": task.TargetId,
				},
			},
			bson.M{
				"$project": bson.M{
					"text": 1,
				},
			},
		}
	case config.TASK_PRODUCT:
		collectionName = config.PRODUCTS
		selectedModel = new(models.ProductDetailForAdminChecker)
		aggregationArray = bson.A{
			bson.M{
				"$match": bson.M{
					"_id": task.TargetId,
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from":         "attributes",
					"localField":   "attrs.attr_id",
					"foreignField": "_id",
					"as":           "attrs_",
					"pipeline": bson.A{
						bson.M{
							"$project": bson.M{
								"name":        "$name.en",
								"placeholder": "$placeholder.en",
								"units_array": 1,
								"is_number":   1,
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
							"as":    "attr",
							"in": bson.M{
								"$mergeObjects": bson.A{
									"$$attr",
									bson.M{
										"attr_detail": bson.M{
											"$first": bson.M{
												"$filter": bson.M{
													"input": "$attrs_",
													"cond": bson.M{
														"$eq": bson.A{"$$this._id", "$$attr.attr_id"},
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
							},
						},
						bson.M{
							"$unwind": bson.M{
								"path": "$parent",
							},
						},
						bson.M{
							"$project": bson.M{
								"name": "$name.en",
								"parent": bson.M{
									"_id":  "$parent._id",
									"name": "$parent.name.en",
								},
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
							},
						},
						bson.M{
							"$unwind": bson.M{
								"preserveNullAndEmptyArrays": true,
								"path":                       "$parent",
							},
						},
						bson.M{
							"$project": bson.M{
								"name": 1,
								"parent": bson.M{
									"_id":  "$parent._id",
									"name": 1,
								},
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
							},
						},
						bson.M{
							"$unwind": bson.M{
								"path": "$city",
							},
						},
						bson.M{
							"$project": bson.M{
								"name":      1,
								"logo":      1,
								"type":      1,
								"city._id":  "$city._id",
								"city.name": "$city.name.en",
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
					"heading":      1,
					"more_details": 1,
					"price":        1,
					"discount":     1,
					"images":       1,
					"brand":        1,
					"category":     1,
					"attrs":        1,
					"seller":       1,
				},
			}}
	case config.TASK_PROFILE:
		collectionName = config.SELLERS
		selectedModel = new(models.SellerProfileForAdminChecker)
		aggregationArray = bson.A{
			bson.M{
				"$match": bson.M{
					"_id": task.TargetId,
				},
			},
			// bson.M{
			// 	"$lookup": bson.M{
			// 		"from":         "cities",
			// 		"localField":   "city_id",
			// 		"foreignField": "_id",
			// 		"as":           "city",
			// 	},
			// },
			// bson.M{
			// 	"$unwind": bson.M{
			// 		"path": "$city",
			// 	},
			// },
			bson.M{
				"$project": bson.M{
					"logo": 1,
					"name": 1,
					// "city":    1,
					"address": 1,
					"bio":     1,
				},
			},
		}
	default:
		return c.JSON(errRes("InvalidTaskType", errors.New("Task type invalid."), config.NOT_FOUND))
	}
	collection := config.MI.DB.Collection(collectionName)
	cursor, err := collection.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate(TaskTarget)", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	if cursor.Next(ctx) {
		err = cursor.Decode(selectedModel)
		if err != nil {
			return c.JSON(errRes("Decode(TaskTarget)", err, config.CANT_DECODE))
		}
		/*
			related product [ 62cf9f30c48e57fb6702a74b]
		*/
	}

	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err(TaskTarget)", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[fiber.Map]{
		IsSuccess: true,
		Result: fiber.Map{
			"type":   task.Type,
			"target": selectedModel,
		},
	})
}

func ConfirmTask(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.ConfirmTask()")
	var employeeObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	// now := time.Now()
	// y, m, d := now.Date()
	// today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	taskId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tasksColl := config.MI.DB.Collection(config.TASKS)
	findResult := tasksColl.FindOne(ctx, bson.M{"_id": taskId})
	if err = findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(Task)", err, config.NOT_FOUND))
	}
	var dbTask models.Task
	err = findResult.Decode(&dbTask)
	if err != nil {
		return c.JSON(errRes("Decode(Task)", err, config.CANT_DECODE))
	}
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	switch dbTask.Type {
	case config.TASK_POST:
		var postData struct {
			Title models.Translation `json:"title"`
			Body  models.Translation `json:"body"`
		}
		err = c.BodyParser(&postData)
		if err != nil {
			return c.JSON(errRes("BodyParser(postData)", err, config.CANT_DECODE))
		}
		tr_postsColl := transaction_manager.Collection(config.POSTS)
		update_model_post := ojoTr.NewModel().
			SetFilter(bson.M{"_id": dbTask.TargetId}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"title":  postData.Title,
					"body":   postData.Body,
					"status": config.STATUS_PUBLISHED,
				},
			}).
			SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
				oldData := i.(bson.M)
				return bson.M{
					"$set": bson.M{
						"title":  oldData["title"],
						"body":   oldData["body"],
						"status": oldData["status"],
					},
				}
			})
		_, err = tr_postsColl.FindOneAndUpdate(update_model_post)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(post)", err, config.CANT_UPDATE))
		}

	case config.TASK_PROFILE:
		var profileData struct {
			Address models.Translation `json:"address"`
			Bio     models.Translation `json:"bio"`
		}
		err = c.BodyParser(&profileData)
		if err != nil {
			return c.JSON(errRes("BodyParser(profileData)", err, config.CANT_DECODE))
		}
		tr_sellersColl := transaction_manager.Collection(config.SELLERS)
		update_model_profile := ojoTr.NewModel().
			SetFilter(bson.M{"_id": dbTask.TargetId}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"address": profileData.Address,
					"bio":     profileData.Bio,
					"status":  config.STATUS_PUBLISHED,
				},
			}).
			SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
				oldData := i.(bson.M)
				return bson.M{
					"$set": bson.M{
						"address": oldData["address"],
						"bio":     oldData["bio"],
						"status":  oldData["status"],
					},
				}
			})
		_, err = tr_sellersColl.FindOneAndUpdate(update_model_profile)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(seller_profile)", err, config.CANT_UPDATE))
		}
	case config.TASK_PRODUCT:
		type ProductAttrs struct {
			Id    primitive.ObjectID `json:"attr_id" bson:"attr_id"`
			Value string             `json:"value" bson:"value"`
			// IsVisible bool               `json:"is_visible" bson:"is_visible"`
			UnitIndex int64 `json:"unit_index" bson:"unit_index"`
		}

		var productData struct {
			Heading     models.Translation `json:"heading" bson:"heading"`
			MoreDetails models.Translation `json:"more_details" bson:"more_details"`
			Attrs       []ProductAttrs     `json:"attrs" bson:"attrs"`
		}
		err = c.BodyParser(&productData)
		if err != nil {
			return c.JSON(errRes("BodyParser(productData)", err, config.CANT_DECODE))
		}
		tr_productsColl := transaction_manager.Collection(config.PRODUCTS)
		update_model_product := ojoTr.NewModel().
			SetFilter(bson.M{"_id": dbTask.TargetId}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"heading":      productData.Heading,
					"more_details": productData.MoreDetails,
					"attrs":        productData.Attrs,
					"status":       config.STATUS_PUBLISHED,
				},
			}).
			SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
				oldData := i.(bson.M)
				return bson.M{
					"$set": bson.M{
						"heading":      oldData["heading"],
						"more_details": oldData["more_details"],
						"attrs":        oldData["attrs"],
						"status":       oldData["status"],
					},
				}
			})
		_, err = tr_productsColl.FindOneAndUpdate(update_model_product)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(product)", err, config.CANT_UPDATE))
		}
		err := AutoPostNewProductAdded(dbTask.TargetId)
		if err != nil {
			// logger bilen log etsek gowy bolar!
			//TODO: transaction.Rollback() edelimi ya bos verelimi?
		}
	// case config.TASK_NOTIFICATION:
	// 	var notifData struct {
	// 		Text models.Translation `json:"text"`
	// 	}
	// 	err = c.BodyParser(&notifData)
	// 	if err != nil {
	// 		return c.JSON(errRes("BodyParser(notifData)", err, config.CANT_DECODE))
	// 	}
	// 	selectedField = config.NOTIFICATIONS
	// 	tr_notifColl := transaction_manager.Collection(config.NOTIFICATIONS)
	// 	update_model_notif := ojoTr.NewModel().
	// 		SetFilter(bson.M{"_id": dbTask.TargetId}).
	// 		SetUpdate(bson.M{"$set": bson.M{
	// 			"text":         notifData.Text,
	// 			"is_confirmed": true,
	// 			"checked_by":   employeeObjId,
	// 		}}).
	// 		SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
	// 			oldData := i.(bson.M)
	// 			return bson.M{
	// 				"$set": bson.M{
	// 					"text":         oldData["text"],
	// 					"is_confirmed": false,
	// 					"checked_by":   oldData["checked_by"],
	// 				},
	// 			}
	// 		})
	// 	_, err = tr_notifColl.FindOneAndUpdate(update_model_notif)
	// 	if err != nil {
	// 		trErr := transaction_manager.Rollback()
	// 		if trErr != nil {
	// 			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
	// 		}
	// 		return c.JSON(errRes("FindOneAndUpdate(notification)", err, config.CANT_UPDATE))
	// 	}
	//TODO: StartNotificationService() bu transaction-dan son bolsa gowy bolar!

	case config.TASK_AUCTION:
		var auctionData struct {
			Heading     models.Translation `json:"heading"`
			Description models.Translation `json:"description"`
		}
		err = c.BodyParser(&auctionData)
		if err != nil {
			return c.JSON(errRes("BodyParser(auctionData)", err, config.CANT_DECODE))
		}
		tr_auctionsColl := transaction_manager.Collection(config.AUCTIONS)
		update_model_auc := ojoTr.NewModel().
			SetFilter(bson.M{"_id": dbTask.TargetId}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"heading":     auctionData.Heading,
					"description": auctionData.Description,
					"status":      config.STATUS_PUBLISHED,
				},
			}).
			SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
				oldData := i.(bson.M)
				return bson.M{
					"$set": bson.M{
						"heading":     oldData["heading"],
						"description": oldData["description"],
						"status":      oldData["status"],
					},
				}
			})
		updateResult, err := tr_auctionsColl.FindOneAndUpdate(update_model_auc)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(auction)", err, config.CANT_UPDATE))
		}
		// CreateCronJobsForAuction(){ AuctionFinishedCronJob(); DeleteFinishedAuction();}
		var auctionTimes struct {
			StartedAt  time.Time `bson:"started_at"`
			FinishedAt time.Time `bson:"finished_at"`
		}
		err = updateResult.Decode(&auctionTimes)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("Decode(finishTime)", err, config.CANT_DECODE))
		}
		var batch []interface{}
		// create a cron job that activates auction at started_at

		// create a cron job for when auction is finished
		modelF := ojocronservice.NewOjoCronJobModel()
		modelF.ListenerName(config.AUCTION_FINISHED)
		modelF.Payload(map[string]interface{}{"auction_id": dbTask.TargetId})
		modelF.RunAt(auctionTimes.FinishedAt)
		batch = append(batch, modelF)
		// create a cron job for when auction is to be deleted automatically
		modelR := ojocronservice.NewOjoCronJobModel()
		modelR.ListenerName(config.AUCTION_REMOVED)
		modelR.Payload(map[string]interface{}{"auction_id": dbTask.TargetId})
		modelR.RunAt(auctionTimes.FinishedAt.Add(time.Hour * 24))

		batch = append(batch, modelR)
		tr_cronJobsColl := transaction_manager.Collection(config.CRON_JOBS)
		insertMany_model := ojoTr.NewModel().SetManyDocuments(batch)
		_, err = tr_cronJobsColl.InsertMany(insertMany_model)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("InsertMany(cron_jobs)", err, config.CANT_INSERT))
		}
	}
	service := config.EverydayWorkService.Of(employeeObjId)
	switch dbTask.Type {
	case config.TASK_AUCTION:
		service.Auction(dbTask.TargetId)
	case config.TASK_POST:
		service.Post(dbTask.TargetId)
		config.StatisticsService.Writer.NewPublishedPost()

	case config.TASK_PRODUCT:
		service.Product(dbTask.TargetId)
		config.StatisticsService.Writer.NewPublishedProduct()
	case config.TASK_PROFILE:
		service.SellerProfile(dbTask.TargetId)
	}
	tr_tasksColl := transaction_manager.Collection(config.TASKS)
	delete_model_task := ojoTr.NewModel().SetFilter(bson.M{"_id": taskId})
	_, err = tr_tasksColl.FindOneAndDelete(delete_model_task)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndDelete(task)", err, config.CANT_DELETE))
	}

	if err = transaction_manager.Err(); err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	// TODO: yokarda tasks collection-dan task-y pozyas, we asakda hem RemoveTask() edyas. gerekmi?
	config.CheckerTaskService.RemoveTask(taskId)

	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.STATUS_COMPLETED,
	})
}

// TODO: admin checker rejects a task with notification
func RejectTask(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.RejectTask")
	/*
		seller.profile, seller.post, seller.product new,update edende adminchecker reject edip
		biler. seller.post.status=rejected etmeli.
		payload {
			params.taskId,
			notification.message
		}
	*/

	// eger task owner == seller ise { send notification with message }
	// update seller.status = "rejected"
	// var taskObjId primitive.ObjectID
	taskObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var notification struct {
		Message string `json:"message"`
		// SellerId primitive.ObjectID `json:"seller_id"`
	}
	err = c.BodyParser(&notification)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tasksColl := config.MI.DB.Collection(config.TASKS)
	findResult := tasksColl.FindOne(ctx, bson.M{"_id": taskObjId})
	if err = findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(task)", err, config.NOT_FOUND))
	}
	var task models.Task
	err = findResult.Decode(&task)
	if err != nil {
		return c.JSON(errRes("Decode(task)", err, config.CANT_DECODE))
	}
	var collectionName string
	title := ""
	switch task.Type {
	case config.TASK_POST:
		collectionName = config.POSTS
		title = fmt.Sprintf("Habar kabul edilmedi(REJECTED)")
	case config.TASK_PRODUCT:
		collectionName = config.PRODUCTS
		title = fmt.Sprintf("Haryt kabul edilmedi(REJECTED)")
	case config.TASK_PROFILE:
		collectionName = config.SELLERS
		title = fmt.Sprintf("Profil kabul edilmedi(REJECTED)")
	default:
		return c.JSON(errRes("task.Type", errors.New("Task type invalid."), config.NOT_ALLOWED))
	}
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_Coll := transaction_manager.Collection(collectionName)
	update_model := ojoTr.NewModel().SetFilter(bson.M{"_id": task.TargetId}).
		SetUpdate(bson.M{
			"$set": bson.M{
				"status": config.STATUS_REJECTED,
			},
		}).
		SetRollbackUpdate(bson.M{
			"$set": bson.M{
				"status": config.STATUS_CHECKING,
			},
		})
	_, err = tr_Coll.FindOneAndUpdate(update_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes(fmt.Sprintf("FindOneAndUpdate(%v)", task.Type), err, config.CANT_UPDATE))
	}
	tr_tasksColl := transaction_manager.Collection(config.TASKS)
	_, err = tr_tasksColl.FindOneAndDelete(ojoTr.NewModel().SetFilter(bson.M{"_id": taskObjId}))
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndDelete(task)", err, config.CANT_DELETE))
	}
	if err = transaction_manager.Err(); err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	config.CheckerTaskService.RemoveTask(taskObjId)

	sellerId := *task.SellerId

	config.NotificationManager.AddNotificationEvent(&notificationmanager.NotificationEvent{
		Title:       title,
		Description: notification.Message,
		ClientIds:   []primitive.ObjectID{sellerId},
		ClientType: notificationmanager.NotificationEventClientType{
			Sellers: true,
		},
	})
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.STATUS_COMPLETED,
	})
}

func AutoPostNewProductAdded(productId primitive.ObjectID) error {

	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": productId,
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
				"path": "$seller",
			},
		},
		bson.M{
			"$project": bson.M{
				"seller": bson.M{
					"_id":  "$seller._id",
					"name": "$seller.name",
					"logo": "$seller.logo",
					"city": "$seller.city",
				},
				"image": bson.M{
					"$first": "$images",
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	cursor, err := productsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return fmt.Errorf("[Aggregate(product)] - db error.")
	}
	defer cursor.Close(ctx)
	var product models.NewProductForAutoPost
	if cursor.Next(ctx) {
		err = cursor.Decode(&product)
		if err != nil {
			return fmt.Errorf("[Decode(product)] - %v", config.CANT_DECODE)
		}
	} else {
		return fmt.Errorf("[cursor.Next(product)] - %v", config.CANT_DECODE)
	}
	if err = cursor.Err(); err != nil {
		return fmt.Errorf("[cursor.Err()] - %v", config.CANT_DECODE)
	}
	if product.Id == primitive.NilObjectID {
		return fmt.Errorf("[NilObjectID] - %v", config.NOT_FOUND)
	}
	var title = models.Translation{
		En: fmt.Sprintf(os.Getenv("NewProductEn"), product.Seller.Name),
		Ru: fmt.Sprintf(os.Getenv("NewProductRu"), product.Seller.Name),
		Tm: fmt.Sprintf(os.Getenv("NewProductTm"), product.Seller.Name),
		Tr: fmt.Sprintf(os.Getenv("NewProductTr"), product.Seller.Name),
	}
	post := models.PostUpsert{
		Image:    product.Image,
		SellerId: product.Seller.Id,
		Title:    title,
		Body: models.Translation{
			Tm: "",
			Ru: "",
			En: "",
			Tr: "",
		},
		RelatedProducts: []primitive.ObjectID{
			product.Id,
		},
		Viewed: 0,
		Likes:  0,
		Auto:   true,
		Status: config.STATUS_PUBLISHED,
	}
	postsColl := config.MI.DB.Collection(config.POSTS)
	_, err = postsColl.InsertOne(ctx, post)
	if err != nil {
		return fmt.Errorf("InsertOne(post): %v", err)
	}
	return nil
}
