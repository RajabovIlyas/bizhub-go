package v1

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/devzatruk/bizhubBackend/ojocronservice"
	ojoTr "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/context"
)

func GetWalletHistory(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("GetWalletHistory")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)

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
				"created_at": 1,
				"amount":     1,
				"intent":     1,
				"status":     1,
				"note": bson.M{
					"$ifNull": bson.A{
						fmt.Sprintf("$note.%v", culture.Lang), nil,
					},
				},

				"code": bson.M{
					"$ifNull": bson.A{
						"$code", nil,
					},
				},
				// "completed_at": 1,
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	whisColl := config.MI.DB.Collection(config.WALLETHISTORY)
	cursor, err := whisColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var whistories = []models.SellerWalletHistory{}
	for cursor.Next(ctx) {
		var history models.SellerWalletHistory
		err = cursor.Decode(&history)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		whistories = append(whistories, history)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.SellerWalletHistory]{
		IsSuccess: true,
		Result:    whistories,
	})
}
func Withdraw(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.Withdraw")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	var toBeWithdrawn struct {
		Sum float64 `json:"sum"`
	}
	err = c.BodyParser(&toBeWithdrawn)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	walletsColl := config.MI.DB.Collection(config.WALLETS)
	findResult := walletsColl.FindOne(ctx, bson.M{
		"seller_id": sellerObjId,
		"closed_at": nil,
		"status":    config.STATUS_ACTIVE,
	})
	if err = findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne()", err, config.NOT_FOUND))
	}
	var sellerWallet models.SellerWallet
	err = findResult.Decode(&sellerWallet)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	if sellerWallet.Balance < toBeWithdrawn.Sum {
		return c.JSON(errRes("BalanceNotEnough", errors.New("No enough funds."), config.NOT_ALLOWED))
	}
	generatedCode := helpers.RandomCodeGenerator()
	now := time.Now()
	wh := models.MyWalletHistory{
		SellerId:    sellerObjId,
		WalletId:    sellerWallet.Id,
		OldBalance:  sellerWallet.Balance,
		Amount:      toBeWithdrawn.Sum,
		Intent:      config.INTENT_WITHDRAW,
		Note:        nil,
		Code:        &generatedCode,
		Status:      config.STATUS_WAITING,
		EmployeeId:  nil,
		CompletedAt: nil,
		CreatedAt:   now,
	}
	// begin transaction
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	whistoriesColl := transaction_manager.Collection(config.WALLETHISTORY)
	insert_model := ojoTr.NewModel().SetDocument(wh)
	walletHistoryInsertResult, err := whistoriesColl.InsertOne(insert_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("InsertOne(wallet_history)", err, config.CANT_INSERT))
	}
	// fmt.Printf("\nInsertOne() ok, waiting for 5 seconds... and trying failing operation: UpdateOne()\n")
	// time.Sleep(time.Second * 5)
	walletsCollTrans := transaction_manager.Collection(config.WALLETS)
	update_model := ojoTr.NewModel().
		SetFilter(bson.M{"_id": sellerWallet.Id}).
		SetUpdate(bson.M{
			"$inc": bson.M{
				"balance": -wh.Amount,
			},
		}).
		SetRollbackUpdate(bson.M{
			"$inc": bson.M{
				"balance": wh.Amount,
			},
		})
	_, err = walletsCollTrans.UpdateOne(update_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("UpdateOne(wallets)", err, config.CANT_UPDATE))
	}

	newWalletHistoryId := walletHistoryInsertResult.InsertedID.(primitive.ObjectID)

	jobModel := ojocronservice.NewOjoCronJobModel()
	jobModel.Group(newWalletHistoryId)
	jobModel.ListenerName(config.CANCEL_WITHDRAW_ACTION).Payload(map[string]interface{}{
		"transaction_id": newWalletHistoryId,
	}).RunAt(wh.CreatedAt.Add(time.Hour * 24))

	err = config.OjoCronService.NewJob(jobModel)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("NewJob(cancel_withdraw_action)", err, config.CANT_INSERT))
	}

	if err = transaction_manager.Err(); err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	return c.JSON(models.Response[fiber.Map]{
		IsSuccess: true,
		Result: fiber.Map{
			"balance": wh.OldBalance - wh.Amount,
		},
	})
}
func GetMyWallet(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetMyWallet")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
				"closed_at": nil,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "sellers",
				"localField":   "seller_id",
				"foreignField": "_id",
				"as":           "seller",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$seller",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "packages",
				"localField":   "seller.package.type",
				"foreignField": "type",
				"as":           "package",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$package",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"in_auction": bson.M{
					"$map": bson.M{
						"input": "$in_auction",
						"as":    "auction",
						"in": bson.M{
							"auction_id": "$$auction.auction_id",
							"amount":     "$$auction.amount",
							"name":       fmt.Sprintf("$$auction.name.%v", culture.Lang),
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"balance":   1,
				"seller_id": 1,
				"package": bson.M{
					"expires_at":   "$seller.package.to",
					"type":         "$package.type",
					"name":         fmt.Sprintf("$package.name.%v", culture.Lang),
					"price":        "$package.price",
					"max_products": "$package.max_products",
				},
				"in_auction": 1,
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	walletColl := config.MI.DB.Collection(config.WALLETS)
	cursor, err := walletColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var sellerWallet models.MyWallet
	if cursor.Next(ctx) {
		err = cursor.Decode(&sellerWallet)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if sellerWallet.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Seller Wallet Not Found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[models.MyWallet]{
		IsSuccess: true,
		Result:    sellerWallet,
	})
}

func CancelWithdraw(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.CancelWithdraw")
	transactionObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	whistoryColl := transaction_manager.Collection(config.WALLETHISTORY)
	now := time.Now()
	update_model := ojoTr.NewModel().
		SetFilter(bson.M{
			"_id":    transactionObjId,
			"intent": config.INTENT_WITHDRAW,
			"status": config.STATUS_WAITING,
		}).
		SetUpdate(bson.M{
			"$set": bson.M{
				"status":       config.STATUS_CANCELLED,
				"completed_at": now,
			},
		}).
		SetRollbackUpdate(bson.M{
			"$set": bson.M{
				"status":       config.STATUS_WAITING,
				"completed_at": nil,
			},
		})
	updateResult, err := whistoryColl.FindOneAndUpdate(update_model)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		return c.JSON(errRes("UpdateOne(wallet_history)", err, config.CANT_UPDATE))
	}
	var oldHistory models.MyWalletHistory
	err = updateResult.Decode(&oldHistory)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		return c.JSON(errRes("Decode(wallet_history)", err, config.CANT_DECODE))
	}
	walletsColl := transaction_manager.Collection(config.WALLETS)
	update_model = ojoTr.NewModel().
		SetFilter(bson.M{"_id": oldHistory.WalletId}).
		SetUpdate(bson.M{"$inc": bson.M{
			"balance": oldHistory.Amount,
		}}).
		SetRollbackUpdate(bson.M{"$inc": bson.M{
			"balance": -oldHistory.Amount,
		}})
	walletResult, err := walletsColl.FindOneAndUpdate(update_model)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(wallet)", err, config.CANT_UPDATE))
	}
	var oldWallet models.SellerWallet
	err = walletResult.Decode(&oldWallet)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		return c.JSON(errRes("Decode(wallet)", err, config.CANT_DECODE))
	}

	err = config.OjoCronService.RemoveJobsByGroup(transactionObjId)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		return c.JSON(errRes("RemoveJobsByGroup(transactionObjId)", err, config.CANT_DECODE))
	}

	return c.JSON(models.Response[fiber.Map]{
		IsSuccess: true,
		Result: fiber.Map{
			"balance": oldWallet.Balance + oldHistory.Amount,
		},
	})
}

func ValidatePayload(p *models.PayPayload) error {
	if p == nil {
		return errors.New(" payload not provided.")
	}
	if !helpers.SliceContains(config.PACKAGE_PAY_CHANGE, p.Action) {
		return errors.New(" invalid package action.")
	}
	if !helpers.SliceContains(config.PACKAGE_TYPES, p.PackageType) {
		return errors.New(" invalid package type.")
	}
	return nil
}
func PayForPackage(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.PayForPackage")
	culture := helpers.GetCultureFromQuery(c)
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	var payPayload models.PayPayload
	err = c.BodyParser(&payPayload)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	err = ValidatePayload(&payPayload)
	if err != nil {
		return c.JSON(errRes("ValidatePayload()", err, config.NOT_ALLOWED))
	}
	// seller has enough money?
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	packagesColl := config.MI.DB.Collection(config.PACKAGES)
	packageResult := packagesColl.FindOne(ctx, bson.M{
		"type": payPayload.PackageType,
	})
	if err = packageResult.Err(); err != nil {
		return c.JSON(errRes("packageResult.Err()", err, config.NOT_FOUND))
	}
	var packageDetail models.PackageWithoutName
	err = packageResult.Decode(&packageDetail)
	if err != nil {
		return c.JSON(errRes("Decode(package)", err, config.CANT_DECODE))
	}
	sellersColl := config.MI.DB.Collection(config.SELLERS)
	cursor, err := sellersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": sellerObjId,
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
							"name":  "$name.en",
							"price": 1,
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
				"package.name":  "$package.detail.name",
				"package.price": "$package.detail.price",
			},
		},
		bson.M{
			"$project": bson.M{
				"package": bson.M{
					"history_id": "$package.package_history_id",
					"type":       "$package.type",
					"name":       "$package.name",
					"price":      "$package.price",
					"to":         "$package.to",
				},
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var sellerInfo struct {
		SellerId primitive.ObjectID `bson:"_id"`
		Package  struct {
			HistoryId primitive.ObjectID `bson:"history_id"`
			Type      string             `bson:"type"`
			Name      string             `bson:"name"`
			Price     float64            `bson:"price"`
			To        time.Time          `bson:"to"`
		} `bson:"package"`
	}
	if cursor.Next(ctx) {
		err = cursor.Decode(&sellerInfo)
		if err != nil {
			return c.JSON(errRes("Decode(sellerInfo)", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if sellerInfo.SellerId == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Seller not found."), config.NOT_FOUND))
	}
	walletsColl := config.MI.DB.Collection(config.WALLETS)
	findResult := walletsColl.FindOne(ctx, bson.M{"seller_id": sellerObjId})
	if err = findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(wallet)", err, config.NOT_FOUND))
	}
	var sellerWallet models.SellerWallet
	err = findResult.Decode(&sellerWallet)
	if err != nil {
		return c.JSON(errRes("Decode(sellerWallet)", err, config.CANT_DECODE))
	}
	if sellerWallet.Balance < packageDetail.Price {
		return c.JSON(errRes("BalanceNotEnough", errors.New("No enough funds."), config.NOT_ALLOWED))
	}
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_walletsColl := transaction_manager.Collection(config.WALLETS)
	tr_updateModel := ojoTr.NewModel().
		SetFilter(bson.M{"_id": sellerWallet.Id}).
		SetUpdate(bson.M{"$inc": bson.M{"balance": -packageDetail.Price}}).
		SetRollbackUpdate(bson.M{"$inc": bson.M{"balance": packageDetail.Price}})
	_, err = tr_walletsColl.FindOneAndUpdate(tr_updateModel)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(wallet)", err, config.CANT_UPDATE))
	}
	// wallet_history.InsertNewTransfer()
	transferNote := ""
	switch culture.Lang {
	case "en":
		transferNote = fmt.Sprintf(os.Getenv("MonthlyPaymentNoteEn"), payPayload.PackageType, packageDetail.Price)
	case "ru":
		transferNote = fmt.Sprintf(os.Getenv("MonthlyPaymentNoteRu"), payPayload.PackageType, packageDetail.Price)
	case "tm":
		transferNote = fmt.Sprintf(os.Getenv("MonthlyPaymentNoteTm"), payPayload.PackageType, packageDetail.Price)
	case "tr":
		transferNote = fmt.Sprintf(os.Getenv("MonthlyPaymentNoteTr"), payPayload.PackageType, packageDetail.Price)
	}
	now := time.Now()
	newTransfer := models.WalletTransfer{
		SellerId:    sellerWallet.SellerId,
		WalletId:    sellerWallet.Id,
		OldBalance:  sellerWallet.Balance,
		Amount:      packageDetail.Price,
		Intent:      config.INTENT_PAYMENT,
		Note:        &transferNote,
		Code:        nil,
		Status:      config.STATUS_COMPLETED,
		EmployeeId:  nil,
		CreatedAt:   now,
		CompletedAt: &now,
	}
	tr_whistoryColl := transaction_manager.Collection(config.WALLETHISTORY)
	tr_insertModel := ojoTr.NewModel().SetDocument(newTransfer)
	tr_insertResult, err := tr_whistoryColl.InsertOne(tr_insertModel)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		return c.JSON(errRes("InsertOne(transfer)", err, config.CANT_INSERT))
	}
	// seller.transfers.updateArray()
	packageExpiresAt := sellerInfo.Package.To.AddDate(0, 1, 0)

	// fmt.Printf("\n                  Package to: %v\n", sellerInfo.Package.To)
	// fmt.Printf("\n          Package Expires at: %v\n", packageExpiresAt)
	// fmt.Printf("\nPackage Expires at FORMATTED: %v\n", packageExpiresAt.Format(time.RFC3339))

	tr_sellersColl := transaction_manager.Collection(config.SELLERS)
	tr_updateModel = ojoTr.NewModel().
		SetFilter(bson.M{"_id": sellerWallet.SellerId}).
		SetUpdate(bson.M{
			"$set": bson.M{
				"package.to": packageExpiresAt, // 1 ay uzaldyas!
			},
			"$push": bson.M{
				"transfers": bson.M{
					"$each":     bson.A{tr_insertResult.InsertedID.(primitive.ObjectID)},
					"$slice":    2,
					"$position": 0,
				},
			},
		}).
		SetRollbackUpdate(bson.M{
			"$set": bson.M{
				"package.to": sellerInfo.Package.To,
			},
			"$pull": bson.M{
				"transfers": tr_insertResult.InsertedID.(primitive.ObjectID),
			},
		})
	_, err = tr_sellersColl.FindOneAndUpdate(tr_updateModel)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(seller)", err, config.CANT_UPDATE))
	}
	if payPayload.Action == config.PACKAGE_CHANGE {
		newPackageHistory := models.SellerPackageHistoryFull{
			SellerId:       sellerObjId,
			From:           now, // ya onki from-danmy?
			To:             packageExpiresAt,
			AmountPaid:     packageDetail.Price,
			Action:         config.PACKAGE_CHANGE,
			OldPackage:     &sellerInfo.Package.Type,
			CurrentPackage: payPayload.PackageType,
		}
		tr_packHisColl := transaction_manager.Collection(config.PACKAGEHISTORY)
		tr_packHisModel := ojoTr.NewModel().SetDocument(newPackageHistory)
		tr_insertResult, err = tr_packHisColl.InsertOne(tr_packHisModel)
		if err != nil {
			errTr := transaction_manager.Rollback()
			if errTr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
			}
			return c.JSON(errRes("InsertOne(package_history)", err, config.CANT_INSERT))
		}
		tr_updateModel = ojoTr.NewModel().
			SetFilter(bson.M{"_id": sellerInfo.SellerId}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"package": bson.M{
						"package_history_id": tr_insertResult.InsertedID.(primitive.ObjectID),
						"to":                 packageExpiresAt,
						"type":               payPayload.PackageType,
					},
				},
			}).
			SetRollbackUpdate(bson.M{
				"$set": bson.M{
					"package": bson.M{
						"package_history_id": sellerInfo.Package.HistoryId,
						"to":                 sellerInfo.Package.To,
						"type":               sellerInfo.Package.Type,
					},
				},
			})
		_, err = tr_sellersColl.FindOneAndUpdate(tr_updateModel)
		if err != nil {
			errTr := transaction_manager.Rollback()
			if errTr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(seller.package)", err, config.CANT_UPDATE))
		}
	}
	return c.JSON(models.Response[fiber.Map]{
		IsSuccess: true,
		Result: fiber.Map{
			"balance":            sellerWallet.Balance - packageDetail.Price,
			"package_expires_at": packageExpiresAt,
		},
	})
}

// func PayForPackage(c *fiber.Ctx) error {
// 	errRes := helpers.ErrorResponse("Mobile.PayForPackage")
// 	var sellerObjId primitive.ObjectID
// 	err := helpers.GetCurrentSeller(c, &sellerObjId)
// 	if err != nil {
// 		return c.JSON(errRes("GetCurrentSeller()", err, ""))
// 	}
// 	var payPayload models.PayPayload
// 	err = c.BodyParser(&payPayload)
// 	if err != nil {
// 		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
// 	}
// 	err = ValidatePayload(&payPayload)
// 	if err != nil {
// 		return c.JSON(errRes("ValidatePayload()", err, ""))
// 	}
// 	// seller has enough money?
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	sellersColl := config.MI.DB.Collection(config.SELLERS)
// 	cursor, err := sellersColl.Aggregate(ctx, bson.A{
// 		bson.M{
// 			"$match": bson.M{
// 				"_id": sellerObjId,
// 			},
// 		},
// 		bson.M{
// 			"$lookup": bson.M{
// 				"from":         "packages",
// 				"localField":   "package.package_id",
// 				"foreignField": "_id",
// 				"as":           "package.detail",
// 				"pipeline": bson.A{
// 					bson.M{
// 						"$project": bson.M{
// 							"price": 1,
// 							"name":  "$name.en",
// 						},
// 					},
// 				},
// 			},
// 		},
// 		bson.M{
// 			"$unwind": bson.M{
// 				"path": "$package.detail",
// 			},
// 		},
// 		bson.M{
// 			"$addFields": bson.M{
// 				"package.name":  "$package.detail.name",
// 				"package.price": "$package.detail.price",
// 			},
// 		},
// 		bson.M{
// 			"$project": bson.M{
// 				"package": bson.M{
// 					"type":  "$package.type",
// 					"name":  "$package.name",
// 					"price": "$package.price",
// 					"to":    "$package.to",
// 				},
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return c.JSON(errRes("Aggregate()", err, config.SERVER_ERROR))
// 	}
// 	defer cursor.Close(ctx)
// 	var sellerInfo struct {
// 		Id      primitive.ObjectID `bson:"_id"`
// 		Package struct {
// 			Type  string    `bson:"type"`
// 			Name  string    `bson:"name"`
// 			Price int64     `bson:"price"`
// 			To    time.Time `bson:"to"`
// 		} `bson:"package"`
// 	}
// 	if cursor.Next(ctx) {
// 		err = cursor.Decode(&sellerInfo)
// 		if err != nil {
// 			return c.JSON(errRes("Decode(sellerInfo)", err, config.CANT_DECODE))
// 		}
// 	}
// 	if err = cursor.Err(); err != nil {
// 		return c.JSON(errRes("cursor.Err()", err, config.NOT_FOUND))
// 	}
// 	if sellerInfo.Id == primitive.NilObjectID {
// 		return c.JSON(errRes("NilObjectID", errors.New("Seller info not found."), config.NOT_FOUND))
// 	}
// 	walletsColl := config.MI.DB.Collection(config.WALLETS)
// 	findResult := walletsColl.FindOne(ctx, bson.M{"seller_id": sellerObjId})
// 	if err = findResult.Err(); err != nil {
// 		return c.JSON(errRes("FindOne(wallet)", err, config.NOT_FOUND))
// 	}
// 	var sellerWallet models.SellerWallet
// 	err = findResult.Decode(&sellerWallet)
// 	if err != nil {
// 		return c.JSON(errRes("Decode(sellerWallet)", err, config.CANT_DECODE))
// 	}
// 	if sellerWallet.Balance < float64(payPayload.Price) {
// 		return c.JSON(errRes("wallet.Balance < package.Price", errors.New("No enough funds."), config.NOT_ALLOWED))
// 	}
// 	updateResult := walletsColl.FindOneAndUpdate(ctx, bson.M{"_id": sellerWallet.Id}, bson.M{
// 		"$inc": bson.M{"balance": -payPayload.Price},
// 	})
// 	if err = updateResult.Err(); err != nil {
// 		return c.JSON(errRes("FindOneAndUpdate(wallet)", err, config.CANT_UPDATE))
// 	}
// 	// wallet_history.InsertNewTransfer()
// 	transferNote := fmt.Sprintf("Monthly payment, %v %v TMT per month.", payPayload.PackageType, payPayload.Price)
// 	now := time.Now()
// 	newTransfer := models.WalletTransfer{
// 		SellerId:    sellerWallet.SellerId,
// 		WalletId:    sellerWallet.Id,
// 		OldBalance:  int64(sellerWallet.Balance),
// 		Amount:      payPayload.Price,
// 		Intent:      "payment",
// 		Note:        &transferNote,
// 		Code:        nil,
// 		Status:      config.COMPLETED,
// 		EmployeeId:  nil,
// 		CreatedAt:   now,
// 		CompletedAt: &now,
// 	}
// 	whistoryColl := config.MI.DB.Collection(config.WALLETHISTORY)
// 	insertResult, err := whistoryColl.InsertOne(ctx, newTransfer)
// 	if err != nil {
// 		return c.JSON(errRes("InsertOne(transfer)", err, config.CANT_INSERT))
// 	}
// 	// seller.transfers.updateArray()
// 	package_expires_at := sellerInfo.Package.To.AddDate(0, 1, 0)
// 	updateResult = sellersColl.FindOneAndUpdate(ctx, bson.M{"_id": sellerWallet.SellerId},
// 		bson.M{
// 			"$set": bson.M{
// 				"package.to": package_expires_at, // 1 ay uzaldyas!
// 			},
// 			"$push": bson.M{
// 				"transfers": bson.M{
// 					"$each":     bson.A{insertResult.InsertedID},
// 					"$slice":    2,
// 					"$position": 0,
// 				},
// 			},
// 		})
// 	if err = updateResult.Err(); err != nil {
// 		return c.JSON(errRes("FindOneAndUpdate(seller)", err, config.CANT_UPDATE))
// 	}
// 	if payPayload.Action == config.PACKAGE_CHANGE {
// 		newPackageHistory := models.SellerPackageHistoryFull{
// 			SellerId:       sellerObjId,
// 			From:           now, // ya onki from-danmy?
// 			To:             package_expires_at,
// 			AmountPaid:     float64(payPayload.Price),
// 			Action:         config.PACKAGE_CHANGE,
// 			OldPackage:     &sellerInfo.Package.Type,
// 			CurrentPackage: payPayload.PackageType,
// 		}
// 		packHisColl := config.MI.DB.Collection(config.PACKAGEHISTORY)
// 		insertResult, err := packHisColl.InsertOne(ctx, newPackageHistory)
// 		if err != nil {
// 			return c.JSON(errRes("InsertOne(package_history)", err, config.CANT_INSERT))
// 		}
// 		updateResult = sellersColl.FindOneAndUpdate(ctx, bson.M{"_id": sellerWallet.SellerId},
// 			bson.M{
// 				"$set": bson.M{
// 					"package": bson.M{
// 						"package_history_id": insertResult.InsertedID,
// 						"to":                 package_expires_at,
// 						"type":               payPayload.PackageType,
// 					},
// 				},
// 			})
// 		if err = updateResult.Err(); err != nil {
// 			return c.JSON(errRes("FindOneAndUpdate(seller.package)", err, config.CANT_UPDATE))
// 		}
// 	}
// 	return c.JSON(models.Response[fiber.Map]{
// 		IsSuccess: true,
// 		Result: fiber.Map{
// 			"balance":            sellerWallet.Balance - float64(payPayload.Price),
// 			"package_expires_at": package_expires_at,
// 		},
// 	})
// }
