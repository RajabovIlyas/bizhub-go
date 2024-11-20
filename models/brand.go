package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type ProductDetailBrand struct {
	Id     primitive.ObjectID       `json:"_id" bson:"_id"`
	Name   string                   `json:"name" bson:"name"`
	Parent ProductDetailBrandParent `json:"parent" bson:"parent"`
}
type ProductDetailBrandParent struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}
type BrandParent struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
	Logo string             `json:"logo" bson:"logo"`
}
type BrandChild struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
	// Logo string             `json:"logo" bson:"logo"`
}
type BrandWithSubBrands struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	Logo      string             `json:"logo" bson:"logo"`
	SubBrands []BrandChild       `json:"sub_brands" bson:"sub_brands"`
}
type NewBrand struct {
	Id         primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string               `json:"name" bson:"name"`
	Logo       *string              `json:"logo" bson:"logo"`
	Parent     *primitive.ObjectID  `json:"parent" bson:"parent"`
	Categories []primitive.ObjectID `json:"categories" bson:"categories"`
	Order      *int64               `json:"order" bson:"order"`
}
type EditBrand struct {
	Id         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string             `json:"name" bson:"name"`
	Parent     *BrandParent       `json:"parent" bson:"parent"`
	Logo       *string            `json:"logo" bson:"logo"`
	Categories []CategoryName     `json:"categories" bson:"categories"`
	Order      *int64             `json:"order" bson:"order"`
}
