package statisticsservice

import (
	"context"
	"fmt"
	"time"

	"github.com/devzatruk/bizhubBackend/ojologger"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StatisticWriterEventType struct {
	Name string
}

func (t *StatisticWriterEventType) setType(type_ string) {
	t.Name = type_
}

func (t *StatisticWriterEventType) newUserEventString() string {
	return "new_user"
}

func (t *StatisticWriterEventType) newDeletedUserEventString() string {
	return "new_deleted_user"
}

func (t *StatisticWriterEventType) newActiveUserEventString() string {
	return "new_active_user"
}

func (t *StatisticWriterEventType) newInactiveUserEventString() string {
	return "new_inactive_user"
}

func (t *StatisticWriterEventType) newSellerEventString() string {
	return "new_seller"
}

func (t *StatisticWriterEventType) newDeletedSellerEventString() string {
	return "new_deleted_seller"
}

func (t *StatisticWriterEventType) newActiveSellerEventString() string {
	return "new_active_seller"
}

func (t *StatisticWriterEventType) newInactiveSellerEventString() string {
	return "new_inactive_seller"
}

func (t *StatisticWriterEventType) newExpenseEventString() string {
	return "new_expense"
}

func (t *StatisticWriterEventType) removeExpenseEventString() string {
	return "remove_expense"
}

func (t *StatisticWriterEventType) publishedProductEventString() string {
	return "published_product"
}

func (t *StatisticWriterEventType) publishedPostEventString() string {
	return "published_post"
}

func (t *StatisticWriterEventType) moneyDepositEventString() string {
	return "money_deposit"
}

func (t *StatisticWriterEventType) moneyWithdrawEventString() string {
	return "money_withdraw"
}

// event

type StatisticWriterEvent struct {
	Type       StatisticWriterEventType
	Payload    interface{}
	Error      error
	RetryCount int64
}

// writer

type StatisticsServiceWriter struct {
	service *StatisticsService
	logger  *ojologger.OjoLogGroup
	queue   chan *StatisticWriterEvent
}

func (w *StatisticsServiceWriter) errorEvent(event *StatisticWriterEvent, err error) {
	log := w.logger.Group("errorEvent()")
	event.Error = err
	event.RetryCount++

	if event.RetryCount >= 3 {
		log.Error(errors.Errorf("failed statistic event; type: %v", event.Type.Name))
		return
	}
	w.queue <- event
}

func (w *StatisticsServiceWriter) withHourAsStringFunc(hour int) func(string) string {
	return func(str string) string {
		return fmt.Sprintf(str, hour)
	}
}

func (w *StatisticsServiceWriter) run() {
	log := w.logger.Group("Run()")
	log.Logf("StatisticServiceWriter started at: %v", time.Now())

	for {
		select {
		case event, ok := <-w.queue:
			if !ok {
				log.Log("Not found event")
				continue
			}

			bsonFilter := bson.M{
				"_id": w.service.recentStatisticId,
			}

			hour := time.Now().Hour()
			withHour := w.withHourAsStringFunc(hour)

			if event.Type.Name == event.Type.moneyDepositEventString() {
				log := w.logger.Group("Run.MoneyDeposited()")

				amount := event.Payload.(float64)
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{"money.deposited": amount, "money.total": amount},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Logf("amount deposited: +%v", amount)
			} else if event.Type.Name == event.Type.moneyWithdrawEventString() {
				log := w.logger.Group("Run.MoneyWithdrew()")

				amount := event.Payload.(float64)
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{"money.withdrew": amount, "money.total": -amount},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Logf("amount withdrawn: -%v", amount)
			} else if event.Type.Name == event.Type.newActiveSellerEventString() {
				log := w.logger.Group("Run.NewActiveSeller()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{
						"sellers.active":                     1,
						withHour("sellers_detail.%v.active"): 1,
						"sellers.active_difference.up":       1,
					},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("active seller +1")
			} else if event.Type.Name == event.Type.newActiveUserEventString() {
				log := w.logger.Group("Run.NewActiveUser()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{
						"users.active":                     1,
						withHour("users_detail.%v.active"): 1,
						"users.active_difference.up":       1,
					},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("active user +1")
			} else if event.Type.Name == event.Type.newInactiveSellerEventString() {
				log := w.logger.Group("Run.NewInactiveSeller()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{
						"sellers.active":                     -1,
						withHour("sellers_detail.%v.active"): -1,
						"sellers.active_difference.down":     -1,
					},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("active seller -1")
			} else if event.Type.Name == event.Type.newInactiveUserEventString() {
				log := w.logger.Group("Run.NewInactiveUser()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{
						"users.active":                     -1,
						withHour("users_detail.%v.active"): -1,
						"users.active_difference.down":     -1,
					},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("active user -1")
			} else if event.Type.Name == event.Type.newDeletedSellerEventString() {
				log := w.logger.Group("Run.NewDeletedSeller()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{
						"sellers.deleted":                     1,
						withHour("sellers_detail.%v.deleted"): 1,
					},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("seller -1")
			} else if event.Type.Name == event.Type.newDeletedUserEventString() {
				log := w.logger.Group("Run.NewDeletedUser()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{
						"users.deleted":                     1,
						withHour("users_detail.%v.deleted"): 1,
					},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("user -1")
			} else if event.Type.Name == event.Type.newExpenseEventString() {
				log := w.logger.Group("Run.NewExpense()")

				exp := event.Payload.(StatisticExpense)
				exp.Id = primitive.NewObjectID()

				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$push": bson.M{"expenses": exp},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("expense +1")
			} else if event.Type.Name == event.Type.newSellerEventString() {
				log := w.logger.Group("Run.NewSeller()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{
						"sellers.all":                     1,
						withHour("sellers_detail.%v.all"): 1,
					},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("seller +1")
			} else if event.Type.Name == event.Type.newUserEventString() {
				log := w.logger.Group("Run.NewUser()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{
						"users.all":                     1,
						withHour("users_detail.%v.all"): 1,
					},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("user +1")
			} else if event.Type.Name == event.Type.publishedPostEventString() {
				log := w.logger.Group("Run.NewPublishedPost()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{"published_posts": 1},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("published post +1")
			} else if event.Type.Name == event.Type.publishedProductEventString() {
				log := w.logger.Group("Run.NewPublishedProduct()")
				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$inc": bson.M{"published_products": 1},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("published product +1")
			} else if event.Type.Name == event.Type.removeExpenseEventString() {
				log := w.logger.Group("Run.RemoveExpense()")

				_, err := w.service.coll.UpdateOne(context.Background(), bsonFilter, bson.M{
					"$pull": bson.M{"expenses": bson.M{
						"_id": event.Payload,
					}},
				})

				if err != nil {
					log.Error(err)
					w.errorEvent(event, err)
					continue
				}

				log.Log("expense -1")
			}
		}
	}
}

// users 4

func (w *StatisticsServiceWriter) NewUser() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.newUserEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

func (w *StatisticsServiceWriter) NewDeletedUser() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.newDeletedUserEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

func (w *StatisticsServiceWriter) NewActiveUser() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.newActiveUserEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

func (w *StatisticsServiceWriter) NewInactiveUser() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.newInactiveUserEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

// sellers 4

func (w *StatisticsServiceWriter) NewSeller() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.newSellerEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

func (w *StatisticsServiceWriter) NewDeletedSeller() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.newDeletedSellerEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

func (w *StatisticsServiceWriter) NewActiveSeller() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.newActiveSellerEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

func (w *StatisticsServiceWriter) NewInactiveSeller() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.newInactiveSellerEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

// expenses 2

func (w *StatisticsServiceWriter) NewExpense(e StatisticExpense) {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.newExpenseEventString())

	w.queue <- &StatisticWriterEvent{
		Type:    type_,
		Payload: e,
	}
}

func (w *StatisticsServiceWriter) RemoveExpense(e primitive.ObjectID) {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.removeExpenseEventString())

	w.queue <- &StatisticWriterEvent{
		Type:    type_,
		Payload: e,
	}
}

// Published Products/Posts 2

func (w *StatisticsServiceWriter) NewPublishedProduct() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.publishedProductEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

func (w *StatisticsServiceWriter) NewPublishedPost() {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.publishedPostEventString())

	w.queue <- &StatisticWriterEvent{
		Type: type_,
	}
}

// money deposit/withdraw 2

func (w *StatisticsServiceWriter) MoneyDeposited(amount float64) {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.moneyDepositEventString())

	w.queue <- &StatisticWriterEvent{
		Type:    type_,
		Payload: amount,
	}
}

func (w *StatisticsServiceWriter) MoneyWithdrew(amount float64) {
	type_ := StatisticWriterEventType{}
	type_.setType(type_.moneyWithdrawEventString())

	w.queue <- &StatisticWriterEvent{
		Type:    type_,
		Payload: amount,
	}
}
