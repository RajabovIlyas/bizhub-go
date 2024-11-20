package everydayworkservice

import (
	"context"
	"time"

	"github.com/devzatruk/bizhubBackend/ojologger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EverydayWorkTask struct {
	Type   string
	TaskId primitive.ObjectID
	Note   string
}

type EverydayWorkServiceClient struct {
	employeeId primitive.ObjectID
	service    *EverydayWorkService
	workId     *primitive.ObjectID
	date       time.Time
	queue      chan *EverydayWorkTask
	logger     *ojologger.OjoLogGroup
}

// private

func (c *EverydayWorkServiceClient) run() {
	log := c.logger.Group("run()")
	err := c.checkDocument()
	if err != nil {
		log.Error(err)

		c.service.mu.Lock()
		delete(c.service.employees, c.employeeId.Hex())
		c.service.mu.Unlock()

		return
	}

	for {
		select {
		case task, ok := <-c.queue:
			if !ok {
				continue
			}

			bsonUpdate := bson.M{}
			if task.Type != "note" {
				bsonUpdate = bson.M{
					"$inc": bson.M{
						"completed_tasks_count": 1,
					},
				}

				if task.Type == "product" {
					bsonUpdate["$push"] = bson.M{
						"tasks.products": task.TaskId,
					}
				} else if task.Type == "post" {
					bsonUpdate["$push"] = bson.M{
						"tasks.posts": task.TaskId,
					}
				} else if task.Type == "seller_profile" {
					bsonUpdate["$push"] = bson.M{
						"tasks.seller_profiles": task.TaskId,
					}
				} else if task.Type == "notification" {
					bsonUpdate["$push"] = bson.M{
						"tasks.notifications": task.TaskId,
					}
				} else if task.Type == "auction" {
					bsonUpdate["$push"] = bson.M{
						"tasks.auctions": task.TaskId,
					}
				} else if task.Type == "cashier_activity" {
					bsonUpdate["$push"] = bson.M{
						"tasks.cashier_activities": task.TaskId,
					}
				} else {
					continue
				}
			} else {
				bsonUpdate = bson.M{
					"$set": bson.M{
						"note": task.Note,
					},
				}
			}

			_, err := c.service.coll.UpdateOne(context.Background(), bson.M{
				"_id": c.workId,
			}, bsonUpdate)
			if err != nil {
				log.Error(err)
				continue
			}

		}
	}
}

func (c *EverydayWorkServiceClient) createNewEverydayWorkDocument(date time.Time) error {

	insertResult, err := c.service.coll.InsertOne(context.Background(), EverydayWork{
		EmployeeId: c.employeeId,
		Date:       date,
		Tasks: EverydayWorkTasks{
			Products:          []primitive.ObjectID{},
			Posts:             []primitive.ObjectID{},
			SellerProfiles:    []primitive.ObjectID{},
			Notifications:     []primitive.ObjectID{},
			Auctions:          []primitive.ObjectID{},
			CashierActivities: []primitive.ObjectID{},
		},
	})
	if err != nil {
		return err
	}

	workId := insertResult.InsertedID.(primitive.ObjectID)
	c.workId = &workId
	c.date = date

	return nil
}

func (c *EverydayWorkServiceClient) checkDocument() error {
	c.service.mu.Lock()
	emp := c.service.employees[c.employeeId.Hex()]
	c.service.mu.Unlock()
	date := c.service.getOnlyDate(time.Now())

	if emp.workId == nil {
		err := c.createNewEverydayWorkDocument(date)
		if err != nil {
			return err
		}

		return nil
	}
	// eger emp.workId != nil ise, asaky code ishleyar, onda
	ctx := context.Background()
	findResult := c.service.coll.FindOne(ctx, bson.M{
		"employee_id": c.employeeId, // TODO: c.workId gozlesek has gowy bolardy!!
		"date":        date,
	})

	if err := findResult.Err(); err != nil {
		return err
	}

	var workIdStruct struct {
		Id primitive.ObjectID `bson:"_id"`
	}

	err := findResult.Decode(&workIdStruct)
	if err != nil {
		return err
	}

	c.workId = &workIdStruct.Id
	c.date = date

	return nil
}

// public

func (c *EverydayWorkServiceClient) Product(id primitive.ObjectID) {
	c.queue <- &EverydayWorkTask{
		Type:   "product",
		TaskId: id,
	}
}

func (c *EverydayWorkServiceClient) Post(id primitive.ObjectID) {
	c.queue <- &EverydayWorkTask{
		Type:   "post",
		TaskId: id,
	}
}

func (c *EverydayWorkServiceClient) SellerProfile(id primitive.ObjectID) {
	c.queue <- &EverydayWorkTask{
		Type:   "seller_profile",
		TaskId: id,
	}
}

func (c *EverydayWorkServiceClient) CashierActivity(id primitive.ObjectID) {
	c.queue <- &EverydayWorkTask{
		Type:   "cashier_activity",
		TaskId: id,
	}
}

func (c *EverydayWorkServiceClient) Notification(id primitive.ObjectID) {
	c.queue <- &EverydayWorkTask{
		Type:   "notification",
		TaskId: id,
	}
}

func (c *EverydayWorkServiceClient) Auction(id primitive.ObjectID) {
	c.queue <- &EverydayWorkTask{
		Type:   "auction",
		TaskId: id,
	}
}

// note

func (c *EverydayWorkServiceClient) Note(note string) {
	c.queue <- &EverydayWorkTask{
		Type: "note",
		Note: note,
	}
}
