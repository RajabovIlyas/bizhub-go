package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EverydayWork struct {
	Id                primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	EmployeeId        primitive.ObjectID   `json:"employee_id" bson:"employee_id"`
	Date              time.Time            `json:"date" bson:"date"`
	WorkTime          WorkTime             `json:"work_time" bson:"work_time"`
	ChecksCount       int64                `json:"checks_count" bson:"checks_count"`
	Products          []primitive.ObjectID `json:"products" bson:"products"`
	Posts             []primitive.ObjectID `json:"posts" bson:"posts"`
	SellerProfiles    []primitive.ObjectID `json:"seller_profiles" bson:"seller_profiles"`
	Notifications     []primitive.ObjectID `json:"notifications" bson:"notifications"`
	Auctions          []primitive.ObjectID `json:"auctions" bson:"auctions"`
	CashierActivities []primitive.ObjectID `json:"cashier_activities" bson:"cashier_activities"`
	Text              string               `json:"text" bson:"text"`
}
