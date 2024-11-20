package ojocronlisteners

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/models"
	notificationmanager "github.com/devzatruk/bizhubBackend/notification_manager"
	"github.com/devzatruk/bizhubBackend/ojocronservice"
	"github.com/devzatruk/bizhubBackend/ojologger"
	ojoTr "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/gofiber/fiber/v2"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleAuctionFinished(job *ojocronservice.OjoCronJob) {
	logger := ojologger.LoggerService.Logger("AddOjoCronListeners()")
	log := logger.Group("handleAuctionFinished()")

	// auctionsColl.update{is_finished:true, status:finished}
	// update walletsColl.in_auction[].deleteByAuctionid()
	// update wallet_history.AddPayment()
	// sellers.transfer[]

	// asaky setirlerin haysy ishlar????
	// auctionId, err := primitive.ObjectIDFromHex(job.Payload.(string))

	auctionId := job.Payload["auction_id"].(primitive.ObjectID)

	// fmt.Printf("\nHandleAuctionFinished(): auctionId: %v\n", auctionId)

	// ???
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// begin transaction
	now := time.Now()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_auctionsColl := transaction_manager.Collection(config.AUCTIONS)
	update_model_auc := ojoTr.NewModel().
		SetFilter(bson.M{"_id": auctionId}).
		SetUpdate(bson.M{
			"$set": bson.M{
				"is_finished": true,
				"status":      config.STATUS_FINISHED,
			},
		}).
		SetRollbackUpdate(bson.M{
			"$set": bson.M{
				"is_finished": false,
				"status":      config.STATUS_PUBLISHED},
		})

	updateResult, err := tr_auctionsColl.FindOneAndUpdate(update_model_auc)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		log.Errorf("FindOneAndUpdate(post): %v - %v", err, config.CANT_UPDATE)
		job.Failed()
		return
	}
	var oldAuctionData models.NewAuction
	err = updateResult.Decode(&oldAuctionData)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		log.Errorf("Decode(oldAuctionData): %v - %v", err, config.CANT_DECODE)
		job.Failed()
		return
	}
	// var winnerModel models.AuctionDetailNewWinner
	tr_walletsColl := transaction_manager.Collection(config.WALLETS)
	tr_walletHistoryColl := transaction_manager.Collection(config.WALLETHISTORY)
	tr_sellersColl := transaction_manager.Collection(config.SELLERS)
	walletHistoryNote := models.Translation{
		En: os.Getenv("AuctionFinishedEn"),
		Tm: os.Getenv("AuctionFinishedTm"),
		Tr: os.Getenv("AuctionFinishedTr"),
		Ru: os.Getenv("AuctionFinishedRu"),
	}
	var winnerIdsForNotificationService = make([]primitive.ObjectID, 0)
	for _, winner := range oldAuctionData.Winners {
		winnerIdsForNotificationService = append(winnerIdsForNotificationService, winner.SellerId)
		update_model_wallet := ojoTr.NewModel().
			SetFilter(bson.M{"seller_id": winner.SellerId}).
			SetUpdate(bson.M{"$pull": bson.M{"in_auction": bson.M{"auction_id": auctionId}}}).
			SetRollbackUpdate(bson.M{"$push": bson.M{
				"in_auction": bson.M{
					"$each": bson.A{fiber.Map{
						"auction_id": auctionId,
						"amount":     winner.LastBid,
						"name":       oldAuctionData.Heading,
					}},
					// "$position": 0,
				},
			}})
			// auctionid 6301f432680dc66abc356dbd
			// seller 1 62ce628c8cae982f654a3578
			// seller 2 635285737705a72c768c2b08 dal 632d8b5caa08ebcafdfa4180
		updateResult, err := tr_walletsColl.FindOneAndUpdate(update_model_wallet)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			log.Errorf("FindOneAndUpdate(wallets.in_auction[]): %v - %v", err, config.CANT_UPDATE)
			job.Failed()
			return
		}
		var oldWallet models.SellerWallet
		err = updateResult.Decode(&oldWallet)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			log.Errorf("Decode(oldWallet): %v - %v", err, config.CANT_DECODE)
			job.Failed()
			return
		}
		// insert into wallet_history
		wh := models.MyWalletHistory{
			SellerId:    winner.SellerId,
			WalletId:    oldWallet.Id,
			OldBalance:  oldWallet.Balance,
			Amount:      float64(winner.LastBid),
			Intent:      config.INTENT_PAYMENT,
			Note:        &walletHistoryNote,
			Code:        nil,
			Status:      config.STATUS_COMPLETED,
			CompletedAt: &now,
			CreatedAt:   now,
			EmployeeId:  nil,
		}
		insert_models_wh := ojoTr.NewModel().SetDocument(wh)
		insertResult, err := tr_walletHistoryColl.InsertOne(insert_models_wh)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			log.Errorf("InsertOne(wallet_history): %v - %v", err, config.CANT_INSERT)
			job.Failed()
			return
		}
		// sellers.transfers[].push()
		update_model_seller := ojoTr.NewModel().
			SetFilter(bson.M{"_id": winner.SellerId}).
			SetUpdate(bson.M{
				"$push": bson.M{
					"transfers": bson.M{
						"$each":     bson.A{insertResult.InsertedID.(primitive.ObjectID)},
						"$slice":    2,
						"$position": 0,
					},
				},
			}).
			SetRollbackUpdate(bson.M{
				"$pull": bson.M{
					"transfers": insertResult.InsertedID.(primitive.ObjectID),
				},
			})

		_, err = tr_sellersColl.FindOneAndUpdate(update_model_seller)
		if err != nil {
			errTr := transaction_manager.Rollback()
			if errTr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
			}
			log.Errorf("FindOneAndUpdate(seller.transfers[]): %v - %v", err, config.CANT_UPDATE)
			job.Failed()
			return
		}
	}
	if len(winnerIdsForNotificationService) > 0 {
		// SendNotificationToWinners(winnerIdsForNotificationService)
		config.NotificationManager.AddNotificationEvent(&notificationmanager.NotificationEvent{
			Title:       "Buşluk! Auksionda yeňiji bolduňyz!",
			Description: fmt.Sprintf("Auksion: %v", oldAuctionData.Heading.Tm),
			ClientType: notificationmanager.NotificationEventClientType{
				Sellers: true,
			},
			ClientIds: winnerIdsForNotificationService,
		})
	}
	job.Finish()
}
