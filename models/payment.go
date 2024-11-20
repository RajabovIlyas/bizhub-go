package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PayPayload struct {
	PackageType string `json:"package_type"`
	Action      string `json:"action"`
}
type WalletTransfer struct {
	Id          primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	SellerId    primitive.ObjectID  `json:"seller_id" bson:"seller_id"`
	WalletId    primitive.ObjectID  `json:"wallet_id" bson:"wallet_id"`
	OldBalance  float64             `json:"old_balance" bson:"old_balance"`
	Amount      float64             `json:"amount" bson:"amount"`
	Intent      string              `json:"intent" bson:"intent"`
	Note        *string             `json:"note" bson:"note"`
	Code        *string             `json:"code" bson:"code"`
	Status      string              `json:"status" bson:"status"`
	EmployeeId  *primitive.ObjectID `json:"employee_id" bson:"employee_id"`
	CompletedAt *time.Time          `json:"completed_at" bson:"completed_at"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
}
