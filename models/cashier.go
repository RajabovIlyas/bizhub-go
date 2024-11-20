package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CompletedCashierWork struct {
	CashierWork `json:",inline" bson:",inline"`
	Seller      Seller `json:"seller" bson:"seller"`
}

type CashierWork struct {
	Id         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	EmployeeId primitive.ObjectID `json:"employee_id" bson:"employee_id"`
	SellerId   primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	Intent     string             `json:"intent" bson:"intent"`
	Amount     float64            `json:"amount" bson:"amount"`
	Code       *string            `json:"code" bson:"code"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
}
