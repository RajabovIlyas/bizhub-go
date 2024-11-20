package ojocronlisteners

import (
	"context"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/devzatruk/bizhubBackend/ojocronservice"
	"github.com/devzatruk/bizhubBackend/ojologger"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductPayload struct {
	ProductId primitive.ObjectID `mapstructure:"product_id"`
	Headings  models.Translation `mapstructure:"headings"`
}

func AutoPostDiscountRemoved(job *ojocronservice.OjoCronJob) {
	logger := ojologger.LoggerService.Logger("AddOjoCronListeners()")
	log := logger.Group("autoPostDiscountRemoved()")

	// ilki cron_jobs-dan bar bolsa ayyr

	var payload ProductPayload

	err := mapstructure.Decode(job.Payload, &payload)
	if err != nil {
		log.Error(err)
		job.Failed()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = config.OjoCronService.RemoveJobsByGroup(job.Group)
	if err != nil {
		log.Errorf("Couldn't delete from cron jobs...")
		job.Failed()
		return
	}
	productsColl := config.MI.DB.Collection(config.PRODUCTS)

	// indi products collection-da product-y tapyp, discount = 0 etmeli, discountDetails =nil etmeli
	_, err = productsColl.UpdateOne(ctx, bson.M{"_id": payload.ProductId}, bson.M{
		"$set": bson.M{
			"discount":      0,
			"discount_data": nil,
		},
	})
	if err != nil {
		job.Failed()
		log.Log("AutoPost update product failed.")
		return
	}

	log.Logf("Product discount removed...")

	pResult, err := productsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": payload.ProductId,
			},
		},
		bson.M{
			"$project": bson.M{
				"image": bson.M{
					"$first": "$images",
				},
				"seller_id": 1,
			},
		},
	})

	if err != nil {
		log.Errorf("AutoPost aggregate error: %v", err)
		job.Failed()
		return
	}

	var product struct {
		SellerId primitive.ObjectID `bson:"seller_id"`
		Image    string             `bson:"image"`
	}
	if pResult.Next(ctx) {
		err = pResult.Decode(&product)
		if err != nil {
			log.Errorf("AutoPost decode error: %v", err)
			job.Failed()
			return
		}
	}

	if err = pResult.Err(); err != nil {
		log.Errorf("AutoPost result error: %v", err)
		job.Failed()
		return
	}
	post := models.PostUpsert{
		Image:           product.Image,
		SellerId:        product.SellerId,
		Title:           payload.Headings,
		Body:            models.Translation{Tm: "", Ru: "", En: "", Tr: ""},
		RelatedProducts: []primitive.ObjectID{payload.ProductId},
		Viewed:          0,
		Likes:           0,
		Auto:            true,
	}
	postsColl := config.MI.DB.Collection(config.POSTS)
	result, err := postsColl.InsertOne(ctx, post)
	if err != nil {
		log.Errorf("AutoPost create post error: %v", err)
		job.Failed()
		return
	}
	// gercek posts collection-a gosmaly
	log.Logf("AutoPost - post published -> %v", result.InsertedID)

	job.Finish()
}
