package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CatAttr struct {
	Id         primitive.ObjectID `json:"_id" bson:"_id"`
	Name       string             `json:"name" bson:"name"`
	IsNumber   bool               `json:"is_number" bson:"is_number"`
	UnitsArray []string           `json:"units_array" bson:"units_array"`
}
type ProductDetailAttribute struct {
	Id    primitive.ObjectID `json:"_id" bson:"attr_id"`
	Value interface{}        `json:"value" bson:"value"`
	// IsVisible       bool               `json:"is_visible" bson:"is_visible"`
	Unit            string    `json:"unit" bson:"unit"`
	AttributeDetail Attribute `json:"attribute_detail" bson:"attr"`
}

type Attribute struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
	// DataType string             `json:"data_type" bson:"data_type"`
}
type AttributeName struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
}
type NewProd_Attr struct {
	Id        primitive.ObjectID `json:"attr_id" bson:"attr_id"`
	Value     interface{}        `json:"value" bson:"value"`
	UnitIndex int64              `json:"unit_index" bson:"unit_index"`
}
type ProductForEditingAttribute struct {
	Id         primitive.ObjectID `json:"attr_id" bson:"attr_id"`
	Value      string             `json:"value" bson:"value"`
	UnitIndex  int64              `json:"unit_index" bson:"unit_index"`
	AttrDetail struct {
		IsNumber   bool     `json:"is_number" bson:"is_number"`
		Name       string   `json:"name" bson:"name"`
		UnitsArray []string `json:"units_array" bson:"units_array"`
	} `json:"attr_detail" bson:"attr_detail"`
}
type NewAttribute struct {
	Id   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name Translation        `json:"name" bson:"name"`
	// Placeholder Translation        `json:"placeholder" bson:"placeholder"`
	Units      string   `json:"units" bson:"-"`
	UnitsArray []string `json:"units_array" bson:"units_array"`
	IsNumber   bool     `json:"is_number" bson:"is_number"`
}

func (a NewAttribute) HasEmptyFields() bool {
	if len(a.Name.En) == 0 || len(a.Name.Ru) == 0 || len(a.Name.Tm) == 0 || len(a.Name.Tr) == 0 {
		// || len(a.UnitsArray) == 0
		return true
	}
	return false
}
