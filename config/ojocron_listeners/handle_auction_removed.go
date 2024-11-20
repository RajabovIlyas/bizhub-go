package ojocronlisteners

import (
	"context"
	"fmt"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/ojocronservice"
	"github.com/devzatruk/bizhubBackend/ojologger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func HandleAuctionRemoved(job *ojocronservice.OjoCronJob) {
	logger := ojologger.LoggerService.Logger("AddOjoCronListeners()")
	log := logger.Group("handleAuctionRemoved()")

	auctionId := job.Payload["auction_id"].(primitive.ObjectID)
	// fmt.Printf("\nHandleAuctionRemoved(): auctionId: %v\n", auctionId)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	auctionsColl := config.MI.DB.Collection(config.AUCTIONS)
	deleteResult, err := auctionsColl.DeleteOne(ctx, bson.M{"_id": auctionId})
	if err != nil {
		log.Errorf("DeleteOne(auction): %v - %v", err, config.CANT_DELETE)
		job.Failed()
		return
	}
	if deleteResult.DeletedCount == 0 {
		log.Errorf("DeleteOne(auction): %v - %v", fmt.Errorf("Auction %v not found.", auctionId), config.NOT_FOUND)
		job.Failed()
		return
	}
	log.Logf("Auction %v removed successfully.", auctionId)
	job.Finish()
}
