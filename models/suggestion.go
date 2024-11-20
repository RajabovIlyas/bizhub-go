package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BRAND    = "brand"
	CATEGORY = "category"
)

type Suggestion struct {
	Id         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Suggestion string             `json:"suggestion" bson:"suggestion"`
	SellerId   primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	Type       string             `json:"type" bson:"type"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
}
