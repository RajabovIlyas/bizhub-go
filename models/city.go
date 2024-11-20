package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type City struct {
	Id   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
}

type CityUpsert struct {
	Id   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name Translation        `json:"name" bson:"name"`
}
