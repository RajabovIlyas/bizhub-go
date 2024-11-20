package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	// Title       string             `json:"title" bson:"title"`
	Description string              `json:"description" bson:"description"`
	TargetId    primitive.ObjectID  `json:"target_id" bson:"target_id"`
	Type        string              `json:"type" bson:"type"`
	IsUrgent    bool                `json:"is_urgent" bson:"is_urgent"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	SellerId    *primitive.ObjectID `json:"seller_id" bson:"seller_id"`
}
type NewTask struct {
	Id          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Description string             `json:"description" bson:"description"`
	TargetId    primitive.ObjectID `json:"target_id" bson:"target_id"`
	Type        string             `json:"type" bson:"type"`
	IsUrgent    bool               `json:"is_urgent" bson:"is_urgent"`
	Seller      *Seller            `json:"seller" bson:"seller"`
}
