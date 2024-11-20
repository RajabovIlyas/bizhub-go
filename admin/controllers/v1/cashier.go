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
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionWithCode struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Status      string             `json:"status" bson:"status"`
	CompletedAt *time.Time         `json:"completed_at" bson:"completed_at"`
	Amount      float64            `json:"amount" bson:"amount"`
}

func Withdraw(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.Withdraw")
	var payload struct {
		TransactionId primitive.ObjectID `json:"transaction_id" bson:"transaction_id"`
	}
	err := c.BodyParser(&payload)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	// tr_id, err := primitive.ObjectIDFromHex(c.Query("tr_id"))
	// if err != nil {
	// 	return c.JSON(errRes("Query(trID)", err, config.QUERY_NOT_PROVIDED))
	// }
	// payload.TransactionId = tr_id
	var employeeObjId primitive.ObjectID
	err = helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// cashier withdraw etjek bolup durka, seller cancel eden bolsa, abort etmeli!
	walletHistoryColl := config.MI.DB.Collection(config.WALLETHISTORY)
	findResult := walletHistoryColl.FindOne(ctx, bson.M{
		"_id":          payload.TransactionId,
		"status":       config.STATUS_WAITING,
		"completed_at": nil,
	})
	if err := findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(transaction)", err, config.NOT_FOUND))
	}
	now := time.Now()
	// y, m, d := now.Date()
	// today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_walletHisColl := transaction_manager.Collection(config.WALLETHISTORY)
	update_model := ojoTr.NewModel().
		SetFilter(bson.M{"_id": payload.TransactionId}).
		SetUpdate(bson.M{
			"$set": bson.M{
				"status":       config.STATUS_COMPLETED,
				"employee_id":  employeeObjId,
				"completed_at": now,
			},
		}).
		SetRollbackUpdate(bson.M{
			"$set": bson.M{
				"status":       config.STATUS_WAITING,
				"employee_id":  nil,
				"completed_at": nil,
			},
		})
	walletHis_updateResult, err := tr_walletHisColl.FindOneAndUpdate(update_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(wallet_history)", err, config.CANT_UPDATE))
	}
	var old_walletHistory models.MyWalletHistory
	err = walletHis_updateResult.Decode(&old_walletHistory)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("Decode(wallet_history)", err, config.CANT_DECODE))
	}
	// cashier_works.Add(completed_task)
	completedTaskId := primitive.NewObjectID()
	taskCode := old_walletHistory.Code
	completedTask := models.CashierWork{
		Id:         completedTaskId,
		EmployeeId: employeeObjId,
		SellerId:   old_walletHistory.SellerId,
		Intent:     old_walletHistory.Intent,
		Amount:     old_walletHistory.Amount,
		Code:       taskCode,
		CreatedAt:  now,
	}
	tr_cashierWorksColl := transaction_manager.Collection(config.CASHIERWORKS)
	insert_model := ojoTr.NewModel().SetDocument(completedTask)
	cashierWork_insertResult, err := tr_cashierWorksColl.InsertOne(insert_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("InsertOne(cashier_work)", err, config.CANT_INSERT))
	}
	employeeManager := config.EverydayWorkService.Of(employeeObjId)
	employeeManager.CashierActivity(cashierWork_insertResult.InsertedID.(primitive.ObjectID))
	config.OjoCronService.RemoveJobsByGroup(payload.TransactionId)

	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.TRANSACTION_SUCCESSFUL,
	})
}
func Deposit(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.Deposit")
	var employeeObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &employeeObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	var payload struct {
		SellerId primitive.ObjectID `json:"seller_id"`
		Amount   float64            `json:"amount"`
	}
	err = c.BodyParser(&payload)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	// for test
	// employeeObjId, err := primitive.ObjectIDFromHex(c.Query("empid"))
	// if err != nil {
	// 	return c.JSON(errRes("Query(empid)", err, config.NOT_FOUND))
	// }
	// sellerId, err := primitive.ObjectIDFromHex(c.Query("id"))
	// if err != nil {
	// 	return c.JSON(errRes("Query(id)", err, config.NOT_FOUND))
	// }
	// payload.SellerId = sellerId
	// amount, err := strconv.ParseFloat(c.Query("amount"), 64)
	// if err != nil {
	// 	return c.JSON(errRes("ParseFloat(amount)", err, config.CANT_DECODE))
	// }
	// payload.Amount = amount
	// end of test
	if payload.Amount < 0 {
		return c.JSON(errRes("NegativeAmount", errors.New("Negative amount not allowed."), config.NOT_ALLOWED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	walletsCollTrans := transaction_manager.Collection(config.WALLETS)
	update_model := ojoTr.NewModel().
		SetFilter(bson.M{"seller_id": payload.SellerId}).
		SetUpdate(bson.M{
			"$inc": bson.M{
				"balance": payload.Amount,
			},
		}).
		SetRollbackUpdate(bson.M{
			"$inc": bson.M{
				"balance": -payload.Amount,
			},
		})
	walletUpdateResult, err := walletsCollTrans.FindOneAndUpdate(update_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("UpdateOne(wallets)", err, config.CANT_UPDATE))
	}
	var walletBeforeUpdate models.SellerWallet
	err = walletUpdateResult.Decode(&walletBeforeUpdate)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(wallet)", err, config.CANT_UPDATE))
	}
	whistoriesColl := transaction_manager.Collection(config.WALLETHISTORY)
	now := time.Now()
	wh := models.MyWalletHistory{
		SellerId:    payload.SellerId,
		WalletId:    walletBeforeUpdate.Id,
		OldBalance:  walletBeforeUpdate.Balance,
		Amount:      payload.Amount,
		Intent:      config.INTENT_DEPOSIT,
		Note:        nil,
		Code:        nil,
		Status:      config.STATUS_COMPLETED,
		CompletedAt: &now,
		EmployeeId:  &employeeObjId,
		CreatedAt:   now,
	}
	insert_model := ojoTr.NewModel().SetDocument(wh)
	wh_insertResult, err := whistoriesColl.InsertOne(insert_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("InsertOne(wallet_history)", err, config.CANT_INSERT))
	}
	// seller.transfers[].push(wallet_history_id) etmeli
	new_wh_transfer := wh_insertResult.InsertedID.(primitive.ObjectID)
	tr_sellersColl := transaction_manager.Collection(config.SELLERS)
	update_model = ojoTr.NewModel().
		SetFilter(bson.M{"_id": wh.SellerId}).
		SetUpdate(bson.M{
			"$push": bson.M{
				"transfers": bson.M{
					"$each":     bson.A{new_wh_transfer},
					"$slice":    2,
					"$position": 0,
				},
			},
		}).
		SetRollbackUpdate(bson.M{
			"$pull": bson.M{
				"transfers": new_wh_transfer,
			},
		})
	_, err = tr_sellersColl.FindOneAndUpdate(update_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(seller.transfers)", err, config.CANT_UPDATE))
	}
	cashierActivityId := primitive.NewObjectID()
	cashierDepositActivity := models.CashierWork{
		Id:         cashierActivityId,
		EmployeeId: employeeObjId,
		SellerId:   wh.SellerId,
		Intent:     config.INTENT_DEPOSIT,
		Amount:     payload.Amount,
		Code:       nil,
		CreatedAt:  now,
	}
	insert_model = ojoTr.NewModel().SetDocument(cashierDepositActivity)
	tr_cashierWorksColl := transaction_manager.Collection(config.CASHIERWORKS)
	_, err = tr_cashierWorksColl.InsertOne(insert_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("InsertOne(cashier_work)", err, config.CANT_INSERT))
	}
	if err = transaction_manager.Err(); err != nil && err != mongo.ErrNoDocuments {
		fmt.Printf("\nInside transaction_manager.Err()\n")
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	employeeManager := config.EverydayWorkService.Of(employeeObjId)
	employeeManager.CashierActivity(cashierActivityId)

	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.TRANSACTION_SUCCESSFUL,
	})
}
func Code(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.Code")
	var payload struct {
		SellerId primitive.ObjectID `json:"seller_id"`
		Code     string             `json:"code"`
	}
	err := c.BodyParser(&payload)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	// sellerId, err := primitive.ObjectIDFromHex(c.Query("seller_id"))
	// if err != nil {
	// 	return c.JSON(errRes("Query(id)", err, config.PARAM_NOT_PROVIDED))
	// }
	// payload.SellerId = sellerId
	// payload.Code = c.Query("code")

	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": payload.SellerId,
				"intent":    config.INTENT_WITHDRAW,
				"code":      payload.Code,
			},
		},
		bson.M{
			"$project": bson.M{
				"status":       1,
				"completed_at": 1,
				"amount":       1,
			},
		},
	}
	var walletHistory TransactionWithCode
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	walletHistoryColl := config.MI.DB.Collection(config.WALLETHISTORY)
	cursor, err := walletHistoryColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	if cursor.Next(ctx) {
		err = cursor.Decode(&walletHistory)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if walletHistory.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Transaction not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[TransactionWithCode]{
		IsSuccess: true,
		Result:    walletHistory,
	})
}
func CompletedTasks(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.CompletedTasks")
	var cashierObjId primitive.ObjectID
	err := helpers.GetCurrentEmployee(c, &cashierObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentEmployee()", err, config.AUTH_REQUIRED))
	}
	// employeeRole := c.Locals(config.EMPLOYEE_JOB).(string)
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	filter := c.Query("filter", "all")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cashierworksColl := config.MI.DB.Collection(config.CASHIERWORKS)
	now := time.Now()
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	match := bson.M{
		"created_at": bson.M{
			"$gt": today,
		},
		"employee_id": cashierObjId,
	}
	if filter == "withdraw" || filter == "deposit" {
		match["intent"] = filter
	}
	aggregationArray := bson.A{
		bson.M{
			"$match": match,
		},
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
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
	}
	cursor, err := cashierworksColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate(cashier_works)", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var works []models.CompletedCashierWork
	for cursor.Next(ctx) {
		var work models.CompletedCashierWork
		err = cursor.Decode(&work)
		if err != nil {
			return c.JSON(errRes("Decode(cashier_work)", err, config.CANT_DECODE))
		}
		works = append(works, work)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err(cashier_work)", err, config.DBQUERY_ERROR))
	}
	if len(works) == 0 {
		works = make([]models.CompletedCashierWork, 0)
	}
	return c.JSON(models.Response[[]models.CompletedCashierWork]{
		IsSuccess: true,
		Result:    works,
	})
}
