package notificationmanager

import (
	"context"
	"fmt"
	"math"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NotificationToken struct {
	Token      string              `bson:"token"`
	Id         primitive.ObjectID  `bson:"_id,omitempty"`
	ClientId   *primitive.ObjectID `bson:"client_id"`
	ClientType *string             `bson:"client_type"`
	OS         string              `bson:"os"`
}

type NotificationEventClientType struct {
	All       bool
	Customers bool
	Sellers   bool
}

type NotificationEvent struct {
	Title       string
	Description string
	ClientIds   []primitive.ObjectID
	ClientType  NotificationEventClientType
	RetryCount  int64
}

type NotificationManager struct {
	database *mongo.Database
	queue    chan *NotificationEvent
	failed   []*NotificationEvent
	client   *messaging.Client
}

func NewNotificationManager() *NotificationManager {
	manager := &NotificationManager{
		queue:  make(chan *NotificationEvent, 1000),
		failed: []*NotificationEvent{},
	}

	go manager.pump()

	return manager
}

func (m *NotificationManager) SetFirebase(firebaseApp *firebase.App) {
	client, err := firebaseApp.Messaging(context.Background())

	if err != nil {
		panic(fmt.Sprintf("notification manager err: %v", err))
	}
	m.client = client
}

func (m *NotificationManager) SetDatabase(database *mongo.Database) {
	m.database = database
}

func (m *NotificationManager) getTokensCount(event *NotificationEvent) (int64, error) {
	notificationsColl := m.database.Collection("notification_tokens")
	match := m.prepareTokensMatchObjectForDb(event)
	return notificationsColl.CountDocuments(context.Background(), match)
}

func (m *NotificationManager) prepareTokensMatchObjectForDb(event *NotificationEvent) bson.M {
	match := bson.M{}

	if len(event.ClientIds) > 0 {
		ids := bson.A{}

		for _, v := range event.ClientIds {
			ids = append(ids, v)
		}

		if len(event.ClientIds) != 0 {
			match["client_id"] = bson.M{
				"$in": ids,
			}
		}
	}

	if event.ClientType.All == false {
		if event.ClientType.Customers {
			match["client_type"] = "customer"
		} else {
			match["client_type"] = "seller"
		}
	}
	return match
}

func (m *NotificationManager) getTokens(event *NotificationEvent, limit int, skip int) ([]string, error) {
	notificationsColl := m.database.Collection("notification_tokens")
	match := m.prepareTokensMatchObjectForDb(event)

	ctx := context.Background()
	tokens := []string{}

	cursor, err := notificationsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": match,
		},
		bson.M{
			"$skip": skip,
		},
		bson.M{
			"$limit": limit,
		},
	})
	if err != nil {
		return []string{}, err
	}

	for cursor.Next(ctx) {
		var token struct {
			Token string `bson:"token"`
		}
		err := cursor.Decode(&token)
		if err != nil {
			return []string{}, err
		}

		tokens = append(tokens, token.Token)
	}

	if err := cursor.Err(); err != nil {
		return []string{}, err
	}

	return tokens, nil
}

func (m *NotificationManager) sendNotification(tokens []string, title string, description string) error {
	// notification-i ugratmaly

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  description,
		},
	}

	_, err := m.client.SendMulticast(context.Background(), message)
	if err != nil {
		fmt.Printf("\n[notification] - send notification error - %v\n", err)
		return err
	}

	return nil
}

func (m *NotificationManager) retryNotification(event *NotificationEvent) {
	event.RetryCount++
	if event.RetryCount >= 3 {
		m.failed = append(m.failed, event)
		return
	}
	m.queue <- event
}

func (m *NotificationManager) pump() {
	for {
		select {
		case event, ok := <-m.queue:
			if !ok {
				continue
			}

			fmt.Printf("\n[notification] - event - %v\n", event)
			tokensCount, err := m.getTokensCount(event)
			if err != nil {
				m.retryNotification(event)
				fmt.Printf("\n[notification] - tokens count error - %v\n", err)
				continue
			}

			limit := 50

			loopCount := int(math.Ceil(float64(tokensCount) / float64(limit))) // 110 / 2

			for i := 0; i < loopCount; i++ {
				tokens, err := m.getTokens(event, limit, i*int(limit))
				if err != nil {
					fmt.Printf("\n[notification] - tokens error - %v\n", err)
					continue
				}

				m.sendNotification(tokens, event.Title, event.Description)
			}

		}
	}
}

func (m *NotificationManager) SaveNotificationToken(token NotificationToken) error {
	notifications := m.database.Collection("notification_tokens")

	ctx := context.Background()
	_, err := notifications.InsertOne(ctx, token)
	if err != nil {
		return err
	}

	return nil
}

func (m *NotificationManager) AddNotificationEvent(event *NotificationEvent) {
	m.queue <- event
}
