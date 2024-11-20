package checkertaskservice

import (
	"context"
	"time"

	"github.com/devzatruk/bizhubBackend/models"
	"github.com/devzatruk/bizhubBackend/ojologger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	MAXRETRYCOUNT = 5
)

type CheckerTask struct {
	Id          primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	TargetId    primitive.ObjectID  `json:"target_id" bson:"target_id"`
	Description string              `json:"description" bson:"description"`
	IsUrgent    bool                `json:"is_urgent" bson:"is_urgent"`
	Type        string              `json:"type" bson:"type"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	SellerId    *primitive.ObjectID `json:"seller_id" bson:"seller_id"`
}

type CheckerTaskWithRetry struct {
	CheckerTask `json:",inline" bson:",inline"`
	RetryCount  int `json:"-" bson:"-"`
}

type CheckerTaskServiceWriter struct {
	queue   chan *CheckerTaskWithRetry
	service *CheckerTaskService
}

func (w *CheckerTaskServiceWriter) run() {
	ctx := context.Background()
	log := ojologger.LoggerService.Logger("CheckerTaskServiceWriter").Group("run()")
	for {
		select {
		case task, ok := <-w.queue:
			if !ok {
				log.Log("channel closed.") // channel closed
				continue                   // TODO: return etmeli dalmi?
			}
			if task.RetryCount == configWriterMaxRetryCount {
				log.Errorf("reached max retry count.")
				continue
			}

			insertResult, err := w.service.coll.InsertOne(ctx, task.CheckerTask)
			if err != nil {
				log.Errorf("task err: %v", err)
				task.RetryCount++
				w.queue <- task
				continue
			}

			task.Id = insertResult.InsertedID.(primitive.ObjectID)
			// ctx := context.Background()
			cursor, err := w.service.coll.Aggregate(ctx, bson.A{
				bson.M{
					"$match": bson.M{
						"_id": task.Id,
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
			})
			if err != nil {
				log.Errorf("aggregate() error: %v", err)
				continue
			}
			if cursor.Next(ctx) {
				var taskForUI models.NewTask
				err := cursor.Decode(&taskForUI)
				if err != nil {
					log.Errorf("aggregate decode() error: %v", err)
					continue
				}
				w.service.announcer.task(taskForUI)
			}
		}
	}
}

func (w *CheckerTaskServiceWriter) Product(targetId primitive.ObjectID, des string, sellerId primitive.ObjectID) {
	w.queue <- &CheckerTaskWithRetry{
		CheckerTask: CheckerTask{
			TargetId:    targetId,
			Description: des,
			IsUrgent:    false,
			Type:        "product",
			CreatedAt:   time.Now(),
			SellerId:    &sellerId,
		},
		RetryCount: 0,
	}
}

func (w *CheckerTaskServiceWriter) Post(targetId primitive.ObjectID, isUrgent bool, des string, sellerId primitive.ObjectID) {
	w.queue <- &CheckerTaskWithRetry{
		CheckerTask: CheckerTask{
			TargetId:    targetId,
			Description: des,
			IsUrgent:    isUrgent,
			Type:        "post",
			CreatedAt:   time.Now(),
			SellerId:    &sellerId,
		},
		RetryCount: 0,
	}
}

func (w *CheckerTaskServiceWriter) Auction(targetId primitive.ObjectID, des string) {
	w.queue <- &CheckerTaskWithRetry{
		CheckerTask: CheckerTask{
			TargetId:    targetId,
			Description: des,
			IsUrgent:    true,
			Type:        "auction",
			CreatedAt:   time.Now(),
			SellerId:    nil,
		},
		RetryCount: 0,
	}
}

func (w *CheckerTaskServiceWriter) SellerProfile(targetId primitive.ObjectID, des string) {
	w.queue <- &CheckerTaskWithRetry{
		CheckerTask: CheckerTask{
			TargetId:    targetId,
			Description: des,
			IsUrgent:    false,
			Type:        "profile",
			CreatedAt:   time.Now(),
			SellerId:    &targetId,
		},
		RetryCount: 0,
	}
}
