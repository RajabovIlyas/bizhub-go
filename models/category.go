package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CategoryAttributes struct {
	Id         primitive.ObjectID `json:"_id" bson:"_id"`
	Attributes []CatAttr          `json:"attributes" bson:"attributes"`
}
type ProductDetailCategory struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
	// Parent ProductDetailCategoryParent `json:"parent,omitempty" bson:"parent,omitempty"`
	Parent *ProductDetailCategoryParent `json:"parent" bson:"parent"`
}
type ProductDetailCategoryParent struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}
type SellerProductCategory struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}
type CategoryParent struct {
	Id    primitive.ObjectID `json:"_id" bson:"_id"`
	Name  string             `json:"name" bson:"name"`
	Image string             `json:"image" bson:"image"`
}
type CategoryChild struct {
	Id    primitive.ObjectID `json:"_id" bson:"_id"`
	Name  string             `json:"name" bson:"name"`
	Image string             `json:"image" bson:"image"`
}
type CategoryWithSubCats struct {
	Id            primitive.ObjectID `json:"_id" bson:"_id"`
	Name          string             `json:"name" bson:"name"`
	Image         string             `json:"image" bson:"image"`
	SubCategories []CategoryName     `json:"sub_categories" bson:"sub_categories"`
}
type CategoryName struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}
type NewCategory struct {
	Id         primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       Translation          `json:"name" bson:"name"`
	Parent     *primitive.ObjectID  `json:"parent" bson:"parent"`
	Image      *string              `json:"image" bson:"image"`
	Attributes []primitive.ObjectID `json:"attributes" bson:"attributes"`
	Order      *int64               `json:"order" bson:"order"`
}
type EditCategory struct {
	Id         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       Translation        `json:"name" bson:"name"`
	Parent     *CategoryParent    `json:"parent" bson:"parent"`
	Image      *string            `json:"image" bson:"image"`
	Attributes []AttributeName    `json:"attributes" bson:"attributes"`
	Order      *int64             `json:"order" bson:"order"`
}

func (c NewCategory) IsParent() bool {
	return c.Parent == nil && len(c.Attributes) == 0 && len(*c.Image) > 0
}
func (c NewCategory) IsChild() bool {
	return c.Parent != nil && len(*c.Image) == 0 && len(c.Attributes) > 0
}
func (c NewCategory) String() string {
	out := "Id: %v - Name: %v - Parent: %v - Image: %v - Attributes: %v"
	return fmt.Sprintf(out, c.Id, c.Name.En, c.Parent, *c.Image, c.Attributes)
}
