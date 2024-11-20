package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Banner struct {
	Id    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Image string             `json:"image" bson:"image"`
}

type BannerForAdmin struct {
	Id    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Image Translation        `json:"image" bson:"image"`
}
