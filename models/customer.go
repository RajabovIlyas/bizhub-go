package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerCredentials struct {
	Phone    string `json:"phone" bson:"phone"`
	Password string `json:"password" bson:"password"`
}
type Customer struct {
	Id       primitive.ObjectID  `json:"_id" bson:"_id"`
	Name     string              `json:"name" bson:"name"`
	Logo     string              `json:"logo" bson:"logo"`
	Phone    string              `json:"phone" bson:"phone"`
	SellerId *primitive.ObjectID `json:"seller_id" bson:"seller_id"`
}
type CustomerWithPassword struct {
	Password string              `json:"password" bson:"password"`
	Id       primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string              `json:"name" bson:"name"`
	Logo     string              `json:"logo" bson:"logo"`
	Phone    string              `json:"phone" bson:"phone"`
	SellerId *primitive.ObjectID `json:"seller_id" bson:"seller_id"`
}
type CustomerForDb struct {
	CustomerWithPassword `json:",inline" bson:",inline"`
	CreatedAt            time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt            *time.Time           `json:"updated_at" bson:"updated_at"`
	Status               string               `json:"status" bson:"status"`
	DeletedProfiles      []primitive.ObjectID `json:"deleted_profiles" bson:"deleted_profiles"`
	Rooms                []primitive.ObjectID `json:"rooms" bson:"rooms"`
}

func (c CustomerForDb) String() string {
	return fmt.Sprintf("Customer ID: %v - Phone: %v - Name: %v - Logo: %v - SellerID: %v - PWD: %v - Status: %v",
		c.Id, c.Phone, c.Name, c.Logo, c.SellerId, c.Password, c.Status)
}

func (c *CustomerWithPassword) WithoutPassword() Customer {
	return Customer{
		Id:       c.Id,
		Name:     c.Name,
		Logo:     c.Logo,
		Phone:    c.Phone,
		SellerId: c.SellerId,
	}
}

type CustomerWithToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Customer
}
