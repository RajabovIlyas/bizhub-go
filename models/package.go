package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Package struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Type        string             `json:"type" bson:"type"`
	Price       float64            `json:"price" bson:"price"`
	MaxProducts int64              `json:"max_products" bson:"max_products"`
	Color       string             `json:"color" bson:"color"`
	TextColor   string             `json:"text_color" bson:"text_color"`
}
type PackageForPayment struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Type        string             `json:"type" bson:"type"`
	Price       float64            `json:"price" bson:"price"`
	MaxProducts int64              `json:"max_products" bson:"max_products"`
}
type PackageWithoutName struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Type        string             `json:"type" bson:"type"`
	Price       float64            `json:"price" bson:"price"`
	MaxProducts int64              `json:"max_products" bson:"max_products"`
}
