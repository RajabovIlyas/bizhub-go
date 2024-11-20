package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Seller struct {
	Id   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
	Logo string             `json:"logo" bson:"logo"`
	Type string             `json:"type" bson:"type"`
	City *City              `json:"city" bson:"city"`
}
type SellerWithStatus struct {
	Seller `json:",inline" bson:",inline"`
	Status string `json:"status" bson:"status"`
}
type SellersWithCount struct {
	SellersCount int64              `json:"sellers_count" bson:"sellers_count"`
	Sellers      []SellerWithStatus `json:"sellers" bson:"sellers"`
}
type SellersFilterAggregations struct {
	Cities []City `json:"cities"`
}

type SellerInfo struct {
	Id            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name          string             `json:"name" bson:"name"`
	Logo          string             `json:"logo" bson:"logo"`
	Type          string             `json:"type" bson:"type"`
	CityId        primitive.ObjectID `json:"city_id" bson:"city_id"`
	City          City               `json:"city" bson:"city"`
	Address       string             `json:"address" bson:"address"`
	Bio           string             `json:"bio" bson:"bio"`
	Likes         int64              `json:"likes" bson:"likes"`
	ProductsCount int64              `json:"products_count" bson:"products_count"`
	PostsCount    int64              `json:"posts_count" bson:"posts_count"`
	Status        string             `json:"status" bson:"status"`
}

type SellerProfilePackage struct {
	Id   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Type string             `json:"type" bson:"type"`
	To   time.Time          `json:"to" bson:"to"`
	Name string             `json:"name" bson:"name"`
}
type SellerPackageHistoryFull struct {
	Id             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SellerId       primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	From           time.Time          `json:"from" bson:"from"`
	To             time.Time          `json:"to" bson:"to"`
	AmountPaid     float64            `json:"amount_paid" bson:"amount_paid"`
	Action         string             `json:"action" bson:"action"`
	OldPackage     *string            `json:"old_package" bson:"old_package"`
	CurrentPackage string             `json:"current_package" bson:"current_package"`
	Text           string             `json:"text" bson:"text"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
}
type SellerPackageHistory struct {
	Id        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	Text      string             `json:"text" bson:"text"`
}
type SellerProfileTransferBy struct {
	Id       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FullName string             `json:"full_name" bson:"full_name"`
}

type SellerProfileTransfer struct {
	Id          primitive.ObjectID       `json:"_id,omitempty" bson:"_id,omitempty"`
	CreatedAt   time.Time                `json:"created_at" bson:"created_at"`
	Type        string                   `json:"type" bson:"type"`
	Code        *string                  `json:"code" bson:"code"`
	Amount      float64                  `json:"amount" bson:"amount"`
	By          *SellerProfileTransferBy `json:"by" bson:"by"`
	Description *string                  `json:"description" bson:"description"`
}

type SellerProfile struct {
	Id         primitive.ObjectID      `json:"_id,omitempty" bson:"_id,omitempty"`
	Name       string                  `json:"name" bson:"name"`
	City       City                    `json:"city" bson:"city"`
	Bio        string                  `json:"bio" bson:"bio"`
	Logo       string                  `json:"logo" bson:"logo"`
	Address    string                  `json:"address" bson:"address"`
	Categories []string                `json:"categories" bson:"categories"`
	Owner      SellerProfileOwner      `json:"owner" bson:"owner"`
	Status     string                  `json:"status" bson:"status"`
	Package    SellerProfilePackage    `json:"package" bson:"package"`
	Transfers  []SellerProfileTransfer `json:"transfers" bson:"transfers"`
	LastIn     []time.Time             `json:"last_in" bson:"last_in"`
	Type       string                  `json:"type" bson:"type"`
}
type NewSeller struct {
	Id            primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name          string               `json:"name" bson:"name"`
	Address       Translation          `json:"address" bson:"address"`
	Logo          string               `json:"logo" bson:"logo"`
	Bio           Translation          `json:"bio" bson:"bio"`
	OwnerId       primitive.ObjectID   `json:"owner_id" bson:"owner_id"`
	CityId        primitive.ObjectID   `json:"city_id" bson:"city_id"`
	Type          string               `json:"type" bson:"type"`
	Likes         int64                `json:"likes" bson:"likes"`
	Categories    []primitive.ObjectID `json:"categories" bson:"categories"`
	PostsCount    int64                `json:"posts_counts" bson:"posts_counts"`
	ProductsCount int64                `json:"products_counts" bson:"products_counts"`
	Status        string               `json:"status" bson:"status"`
	LastIn        []time.Time          `json:"last_in" bson:"last_in"`
	Transfers     []primitive.ObjectID `json:"transfers" bson:"transfers"`
	Package       SellerCurrentPackage `json:"package" bson:"package"`
}
type SellerInfoFull struct {
	Id            primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Name          string               `json:"name" bson:"name"`
	Address       string               `json:"address" bson:"address"`
	Logo          string               `json:"logo" bson:"logo"`
	Bio           string               `json:"bio" bson:"bio"`
	OwnerId       primitive.ObjectID   `json:"owner_id" bson:"owner_id"`
	CityId        primitive.ObjectID   `json:"city_id" bson:"city_id"`
	Type          string               `json:"type" bson:"type"`
	Likes         int64                `json:"likes" bson:"likes"`
	Categories    []primitive.ObjectID `json:"categories" bson:"categories"`
	PostsCount    int64                `json:"posts_counts" bson:"posts_counts"`
	ProductsCount int64                `json:"products_counts" bson:"products_counts"`
	Status        string               `json:"status" bson:"status"`
	LastIn        []time.Time          `json:"last_in" bson:"last_in"`
	Transfers     []primitive.ObjectID `json:"transfers" bson:"transfers"`
	Package       SellerCurrentPackage `json:"package" bson:"package"`
}

type SellerCurrentPackage struct {
	PackageHistoryId primitive.ObjectID `json:"package_history_id" bson:"package_history_id"`
	To               time.Time          `json:"to" bson:"to"`
	Type             string             `json:"type" bson:"type"`
}
type SellerProfileOwner struct {
	Id    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Phone string             `json:"phone" bson:"phone"`
	Name  string             `json:"name" bson:"name"`
}
type SellerProfileForAdminChecker struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Logo string             `json:"logo" bson:"logo"`
	Name string             `json:"name" bson:"name"`
	// City    CityUpsert         `json:"city" bson:"city"`
	Address Translation `json:"address" bson:"address"`
	Bio     Translation `json:"bio" bson:"bio"`
}
