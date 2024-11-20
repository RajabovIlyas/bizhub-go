package taskmanager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskManager struct {
	database   *mongo.Database
	retryCount int
	TaskType   string
	Task       *models.Task
}

func NewTask(db *mongo.Database, retryCount int) *TaskManager {
	return &TaskManager{database: db, retryCount: retryCount, TaskType: ""}
}
func (m *TaskManager) Auction(auctionId primitive.ObjectID, heading string) *TaskManager {
	m.TaskType = config.TASK_AUCTION
	m.Task = &models.Task{
		IsUrgent:    true,
		Description: heading,
		TargetId:    auctionId,
		Type:        config.TASK_AUCTION,
		CreatedAt:   time.Now(),
	}
	return m
}

// func (m *TaskManager) Notification(notificationId primitive.ObjectID, message string) *TaskManager {
// 	m.TaskType = config.TASK_NOTIFICATION
// 	m.Task = &models.Task{
// 		IsUrgent:    true,
// 		Description: message,
// 		TargetId:    notificationId,
// 		Type:        config.TASK_NOTIFICATION,
// 		CreatedAt:   time.Now(),
// 	}
// 	return m
// }
func (m *TaskManager) Post(postId primitive.ObjectID, title string, isReporterBee bool, sellerId primitive.ObjectID) *TaskManager {
	m.TaskType = config.TASK_POST
	m.Task = &models.Task{
		IsUrgent:    isReporterBee,
		Description: title,
		TargetId:    postId,
		Type:        config.TASK_POST,
		CreatedAt:   time.Now(),
		SellerId:    &sellerId,
	}
	return m
}
func (m *TaskManager) Product(productId primitive.ObjectID, heading string, sellerId primitive.ObjectID) *TaskManager {
	m.TaskType = config.TASK_PRODUCT
	m.Task = &models.Task{
		IsUrgent:    false,
		Description: heading,
		TargetId:    productId,
		Type:        config.TASK_PRODUCT,
		CreatedAt:   time.Now(),
		SellerId:    &sellerId,
	}
	return m
}
func (m *TaskManager) SellerProfile(sellerId primitive.ObjectID, bio string) *TaskManager {
	m.TaskType = config.TASK_PROFILE
	m.Task = &models.Task{
		IsUrgent:    false,
		Description: bio,
		TargetId:    sellerId,
		Type:        config.TASK_PROFILE,
		CreatedAt:   time.Now(),
		SellerId:    &sellerId,
	}
	return m
}
func (m *TaskManager) Commit() error {
	fmt.Printf("\ninside task manager commit()...\n")
	if m.TaskType == "" || m.Task == nil {
		return fmt.Errorf("Task not properly created.")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := m.database.Collection(config.TASKS)
	var err error
	for i := 0; i < m.retryCount; i++ {
		_, err = collection.InsertOne(ctx, m.Task)
		if err != nil {
			continue
		}
		break
	}
	m.TaskType = ""
	m.Task = nil
	if err != nil {
		// logger ulansak gowy bolar!
		return errors.New("Task couldn't be created.")
	}
	return nil
}
func a() {
	t := NewTask(config.MI.DB, 3)
	t.Auction(primitive.NewObjectID(), "bu desc")
	t.Commit()
}
