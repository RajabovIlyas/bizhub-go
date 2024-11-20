package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Collection struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
	Path string             `json:"path" bson:"path"`
}
type CollectionUpsert struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name Translation        `json:"name" bson:"name"`
	Path string             `json:"path" bson:"path"`
}
