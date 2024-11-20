package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Feedback struct {
	Seller                `json:"seller" bson:"seller"`
	FeedbackWithoutSeller `json:",inline" bson:",inline"`
}
type FeedbackWithoutSeller struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	IsRead    bool               `json:"is_read" bson:"is_read"`
	Text      string             `json:"text" bson:"text"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
