package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationForEverydayWork struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Audience Audience           `json:"audience" bson:"audience"`
	Text     string             `json:"text" bson:"text"`
}

type Notification struct {
	Id       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Audience Audience           `json:"audience" bson:"audience"`
	// Text        Translation         `json:"text" bson:"text"`
	Text        string              `json:"text" bson:"text"`
	CreatedBy   primitive.ObjectID  `json:"created_by" bson:"created_by"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	IsConfirmed bool                `json:"is_confirmed" bson:"is_confirmed"`
	CheckedBy   *primitive.ObjectID `json:"checked_by" bson:"checked_by"`
}
type Audience struct {
	All     bool `json:"all" bson:"all"`
	Users   bool `json:"users" bson:"users"`
	Sellers bool `json:"sellers" bson:"sellers"`
}
type NotificationForAdminChecker struct {
	Id primitive.ObjectID `json:"_id" bson:"_id"`
	// Text Translation        `json:"text" bson:"text"`
	Text string `json:"text" bson:"text"`
}
