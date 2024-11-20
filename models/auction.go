package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Auction struct {
	Id         primitive.ObjectID `json:"_id" bson:"_id"`
	Image      string             `json:"image" bson:"image"`
	Heading    string             `json:"heading" bson:"heading"`
	TextColor  string             `json:"text_color" bson:"text_color"`
	StartedAt  time.Time          `json:"started_at" bson:"started_at"`
	FinishedAt time.Time          `json:"finished_at" bson:"finished_at"`
	IsFinished bool               `json:"is_finished" bson:"is_finished"`
}

type AuctionDetailWinner struct {
	Seller                 Seller `json:"seller" bson:"seller"`
	Index                  int64  `json:"index" bson:"index"`
	AuctionDetailNewWinner `json:",inline" bson:",inline"`
}
type AuctionDetailNewWinner struct {
	SellerId  primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	LastBid   float64            `json:"last_bid" bson:"last_bid"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type AuctionDetail struct {
	Id            primitive.ObjectID    `json:"_id" bson:"_id"`
	Image         string                `json:"image" bson:"image"`
	Heading       string                `json:"heading" bson:"heading"`
	Description   string                `json:"description" bson:"description"`
	Participants  int64                 `json:"participants" bson:"participants"`
	Winners       []AuctionDetailWinner `json:"winners" bson:"winners"`
	InitialMinBid int64                 `json:"initial_min_bid" bson:"initial_min_bid"`
	MinimalBid    int64                 `json:"minimal_bid" bson:"minimal_bid"`
	StartedAt     time.Time             `json:"started_at" bson:"started_at"`
	FinishedAt    time.Time             `json:"finished_at" bson:"finished_at"`
	IsFinished    bool                  `json:"is_finished" bson:"is_finished"`
}
type NewAuction struct {
	Image             string                   `bson:"image"`
	Heading           Translation              `bson:"heading"`
	TextColor         string                   `bson:"text_color"`
	Description       Translation              `bson:"description"`
	StartedAt         time.Time                `bson:"started_at"`
	FinishedAt        time.Time                `bson:"finished_at"`
	Participants      int                      `bson:"participants"`
	InitialMinimalBid int                      `bson:"initial_minimal_bid"`
	MinimalBid        int                      `bson:"minimal_bid"`
	IsFinished        bool                     `bson:"is_finished"`
	Winners           []AuctionDetailNewWinner `bson:"winners"`
	Status            string                   `bson:"status"`
	CreatedAt         time.Time                `bson:"created_at"`
}

func (a NewAuction) HasEmptyFields() bool {
	if len(a.Image) == 0 || len(a.Heading.En) == 0 || len(a.TextColor) == 0 || len(a.Description.En) == 0 {
		return true
	}
	return false
}

type BidAuctionFind struct {
	Id            primitive.ObjectID    `json:"_id" bson:"_id"`
	InitialMinBid float64               `json:"initial_min_bid" bson:"initial_min_bid"`
	MinimalBid    float64               `json:"minimal_bid" bson:"minimal_bid"`
	Winners       []AuctionDetailWinner `json:"winners" bson:"winners"`
	Participants  int64                 `json:"participants" bson:"participants"`
	Heading       Translation           `json:"heading" bson:"heading"`
}
type AuctionForAdminChecker struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id"`
	Heading     Translation        `json:"heading" bson:"heading"`
	Description Translation        `json:"description" bson:"description"`
}
