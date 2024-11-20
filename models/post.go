package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	Id    primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Image string             `json:"image" bson:"image"`
	Title string             `json:"title" bson:"title"`
	// Body      string             `json:"body" bson:"body"`
	SellerId primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	Seller   Seller             `json:"seller" bson:"seller,omitempty"`
	Viewed   int64              `json:"viewed" bson:"viewed"`
	Likes    int64              `json:"likes" bson:"likes"`
	// CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type ReporterBeePost struct {
	Id          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Image       string             `json:"image" bson:"image"`
	Title       string             `json:"title" bson:"title"`
	Body        string             `json:"body" bson:"body"`
	SellerId    primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	ReporterBee ReporterBee        `json:"reporter_bee" bson:"reporter_bee,omitempty"`
	Viewed      int64              `json:"viewed" bson:"viewed"`
	Likes       int64              `json:"likes" bson:"likes"`
}

type ReporterBeePostDetail struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Image    string             `json:"image" bson:"image"`
	Title    string             `json:"title" bson:"title"`
	Body     string             `json:"body" bson:"body"`
	SellerId primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	Viewed   int64              `json:"viewed" bson:"viewed"`
	Likes    int64              `json:"likes" bson:"likes"`
}

type PostWithoutSeller struct {
	Id    primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Image string             `json:"image" bson:"image"`
	Title string             `json:"title" bson:"title"`
	// Body     string             `json:"body" bson:"body"`
	SellerId primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	Viewed   int64              `json:"viewed" bson:"viewed"`
	Likes    int64              `json:"likes" bson:"likes"`
	Status   string             `json:"status" bson:"status"`
}
type PostUpsert struct {
	Id              primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Title           Translation          `json:"title" bson:"title"`
	Body            Translation          `json:"body" bson:"body"`
	RelatedProducts []primitive.ObjectID `json:"related_products" bson:"related_products"`
	SellerId        primitive.ObjectID   `json:"seller_id" bson:"seller_id"`
	Viewed          int64                `json:"viewed" bson:"viewed"`
	Image           string               `json:"image" bson:"image"`
	CreatedAt       time.Time            `json:"created_at" bson:"created_at"`
	Likes           int64                `json:"likes" bson:"likes"`
	Status          string               `json:"status" bson:"status"`
	Auto            bool                 `json:"auto" bson:"auto"`
}
type NewPost struct {
	Id              primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	Title           string               `json:"title" bson:"title"`
	Body            string               `json:"body" bson:"body"`
	Image           string               `json:"image" bson:"image"`
	RelatedProducts []primitive.ObjectID `json:"related_products" bson:"related_products"`
}
type RelatedProduct struct {
	Id    primitive.ObjectID `json:"_id" bson:"_id"`
	Image string             `json:"image" bson:"image"`
}
type RelatedProductUpsert struct {
	Id primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
}

type PostDetail struct {
	PostDetailWithoutSeller `json:",inline" bson:",inline"`
	Seller                  Seller `json:"seller" bson:"seller,omitempty"`
}
type PostDetailWithoutSeller struct {
	Id              primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Image           string             `json:"image" bson:"image"`
	Title           string             `json:"title" bson:"title"`
	Body            string             `json:"body" bson:"body"`
	SellerId        primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	RelatedProducts []RelatedProduct   `json:"related_products" bson:"related_products"`
	Viewed          int64              `json:"viewed" bson:"viewed"`
	Likes           int64              `json:"likes" bson:"likes"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	// Auto            bool               `json:"auto" bson:"auto"`
	Status string `json:"status" bson:"status"`
}
type PostDetailWithTranslationWithoutSeller struct {
	Id              primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Image           string             `json:"image" bson:"image"`
	Title           Translation        `json:"title" bson:"title"`
	Body            Translation        `json:"body" bson:"body"`
	SellerId        primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	RelatedProducts []RelatedProduct   `json:"related_products" bson:"related_products"`
	Viewed          int64              `json:"viewed" bson:"viewed"`
	Likes           int64              `json:"likes" bson:"likes"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	// Auto            bool               `json:"auto" bson:"auto"`
	Status string `json:"status" bson:"status"`
}

type PostForAdminChecker struct {
	Id              primitive.ObjectID `json:"_id" bson:"_id"`
	Image           string             `json:"image" bson:"image"`
	Heading         Translation        `json:"title" bson:"title"`
	Description     Translation        `json:"body" bson:"body"`
	RelatedProducts []RelatedProduct   `json:"related_products" bson:"related_products"`
	Seller          Seller             `json:"seller" bson:"seller"`
}
