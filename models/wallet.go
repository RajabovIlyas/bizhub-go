package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SellerWallet struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SellerId  primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	Balance   float64            `json:"balance" bson:"balance"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	ClosedAt  *time.Time         `json:"closed_at" bson:"closed_at"`
	Status    string             `json:"status" bson:"status"`
	InAuction []InAuctionObject  `json:"in_auction" bson:"in_auction"`
}
type MyWallet struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SellerId  primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	Balance   float64            `json:"balance" bson:"balance"`
	Package   MyWalletPackage    `json:"package" bson:"package"`
	InAuction []InAuction        `json:"in_auction" bson:"in_auction"`
}

type MyWalletPackage struct {
	Type        string    `json:"type" bson:"type"`
	Name        string    `json:"name" bson:"name"`
	Price       float64   `json:"price" bson:"price"`
	MaxProducts int64     `json:"max_products" bson:"max_products"`
	ExpiresAt   time.Time `json:"expires_at" bson:"expires_at"`
}
type InAuction struct {
	AuctionId primitive.ObjectID `json:"auction_id,omitempty" bson:"auction_id,omitempty"`
	Amount    int64              `json:"amount" bson:"amount"`
	Name      string             `json:"name" bson:"name"`
}
type InAuctionObject struct {
	AuctionId primitive.ObjectID `json:"auction_id,omitempty" bson:"auction_id,omitempty"`
	Amount    int64              `json:"amount" bson:"amount"`
	Name      Translation        `json:"name" bson:"name"`
}
type SellerWalletHistory struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Amount    float64            `json:"amount" bson:"amount"`
	Intent    string             `json:"intent" bson:"intent"`
	Note      *string            `json:"note" bson:"note"`
	Code      *string            `json:"code" bson:"code"`
	Status    string             `json:"status" bson:"status"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
type MyWalletHistory struct {
	Id          primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	SellerId    primitive.ObjectID  `json:"seller_id" bson:"seller_id"`
	WalletId    primitive.ObjectID  `json:"wallet_id" bson:"wallet_id"`
	OldBalance  float64             `json:"old_balance" bson:"old_balance"`
	Amount      float64             `json:"amount" bson:"amount"`
	Intent      string              `json:"intent" bson:"intent"`
	Note        *Translation        `json:"note" bson:"note"`
	Code        *string             `json:"code" bson:"code"`
	Status      string              `json:"status" bson:"status"`
	CreatedAt   time.Time           `json:"created_at" bson:"created_at"`
	CompletedAt *time.Time          `json:"completed_at" bson:"completed_at"`
	EmployeeId  *primitive.ObjectID `json:"employee_id" bson:"employee_id"`
}
