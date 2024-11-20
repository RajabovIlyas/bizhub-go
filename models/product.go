package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	Id       primitive.ObjectID `json:"_id" bson:"_id"`
	Heading  string             `json:"heading" bson:"heading"`
	Price    float64            `json:"price" bson:"price"`
	Discount float64            `json:"discount" bson:"discount"`
	IsNew    bool               `json:"is_new" bson:"is_new"`
	Image    string             `json:"image" bson:"image"`
	Status   string             `json:"status" bson:"status"`
}
type NewProductData struct {
	Id          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Heading     string             `json:"heading" bson:"heading"`
	MoreDetails string             `json:"more_details" bson:"more_details"`
	CategoryId  primitive.ObjectID `json:"category_id" bson:"category_id"`
	BrandId     primitive.ObjectID `json:"brand_id" bson:"brand_id"`
	Price       float64            `json:"price" bson:"price"`
	Attributes  []NewProd_Attr     `json:"attributes" bson:"attributes"`
	Images      []string           `json:"images" bson:"images"`
}

type ProductDetail struct {
	ProductDetailWithoutSeller `json:",inline" bson:",inline"`
	Seller                     Seller `json:"seller" bson:"seller"`
}
type ProductDetailWithoutSeller struct {
	Id           primitive.ObjectID       `json:"_id" bson:"_id"`
	Heading      string                   `json:"heading" bson:"heading"`
	Price        float64                  `json:"price" bson:"price"`
	Discount     float64                  `json:"discount" bson:"discount"`
	DiscountData *DiscountData            `json:"discount_data" bson:"discount_data"`
	Images       []string                 `json:"images" bson:"images"`
	CategoryId   primitive.ObjectID       `json:"category_id" bson:"category_id"`
	BrandId      primitive.ObjectID       `json:"brand_id" bson:"brand_id"`
	MoreDetails  string                   `json:"more_details" bson:"more_details"`
	SellerId     primitive.ObjectID       `json:"seller_id" bson:"seller_id"`
	Attributes   []ProductDetailAttribute `json:"attributes" bson:"attrs"`
	Category     ProductDetailCategory    `json:"category" bson:"category"`
	Brand        ProductDetailBrand       `json:"brand" bson:"brand"`
	Viewed       int64                    `json:"viewed" bson:"viewed"`
	Likes        int64                    `json:"likes" bson:"likes"`
	Status       string                   `json:"status" bson:"status"`
}
type ProductDetailWithTranslationWithoutSeller struct {
	Id           primitive.ObjectID       `json:"_id" bson:"_id"`
	Heading      Translation              `json:"heading" bson:"heading"`
	Price        float64                  `json:"price" bson:"price"`
	Discount     float64                  `json:"discount" bson:"discount"`
	DiscountData *DiscountData            `json:"discount_data" bson:"discount_data"`
	Images       []string                 `json:"images" bson:"images"`
	CategoryId   primitive.ObjectID       `json:"category_id" bson:"category_id"`
	BrandId      primitive.ObjectID       `json:"brand_id" bson:"brand_id"`
	MoreDetails  Translation              `json:"more_details" bson:"more_details"`
	SellerId     primitive.ObjectID       `json:"seller_id" bson:"seller_id"`
	Attributes   []ProductDetailAttribute `json:"attributes" bson:"attrs"`
	Category     ProductDetailCategory    `json:"category" bson:"category"`
	Brand        ProductDetailBrand       `json:"brand" bson:"brand"`
	Viewed       int64                    `json:"viewed" bson:"viewed"`
	Likes        int64                    `json:"likes" bson:"likes"`
	Status       string                   `json:"status" bson:"status"`
}
type ProductDetailForEditing struct {
	Id          primitive.ObjectID           `json:"_id" bson:"_id"`
	Heading     string                       `json:"heading" bson:"heading"`
	MoreDetails string                       `json:"more_details" bson:"more_details"`
	Price       float64                      `json:"price" bson:"price"`
	Images      []string                     `json:"images" bson:"images"`
	Brand       ProductDetailBrand           `json:"brand" bson:"brand"`
	Category    ProductDetailCategory        `json:"category" bson:"category"`
	Attrs       []ProductForEditingAttribute `json:"attrs" bson:"attrs"`
}
type ProductDetailWithTranslation struct {
	Id           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Heading      Translation        `json:"heading" bson:"heading"`
	CategoryId   primitive.ObjectID `json:"category_id" bson:"category_id"`
	BrandId      primitive.ObjectID `json:"brand_id" bson:"brand_id"`
	Images       []string           `json:"images" bson:"images"`
	Price        float64            `json:"price" bson:"price"`
	Discount     float64            `json:"discount" bson:"discount"`
	MoreDetails  Translation        `json:"more_details" bson:"more_details"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	SellerId     primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	Attributes   []NewProd_Attr     `json:"attrs" bson:"attrs"`
	Viewed       int64              `json:"viewed" bson:"viewed"`
	Likes        int64              `json:"likes" bson:"likes"`
	Status       string             `json:"status" bson:"status"`
	DiscountData *DiscountData      `json:"discount_data" bson:"discount_data"`
	// Category     ProductDetailCategory    `json:"category" bson:"category"`
	// Brand        ProductDetailBrand       `json:"brand" bson:"brand"`
}
type ProductDetailForAdminChecker struct {
	Id          primitive.ObjectID           `json:"_id" bson:"_id"`
	Heading     Translation                  `json:"heading" bson:"heading"`
	Images      []string                     `json:"images" bson:"images"`
	Price       float64                      `json:"price" bson:"price"`
	Discount    float64                      `json:"discount" bson:"discount"`
	MoreDetails Translation                  `json:"more_details" bson:"more_details"`
	Attrs       []ProductAttrForAdminChecker `json:"attrs" bson:"attrs"`
	Category    ProductCatForAdminChecker    `json:"category" bson:"category"`
	Brand       ProductBrandForAdminChecker  `json:"brand" bson:"brand"`
	Seller      SellerForAdminChecker        `json:"seller" bson:"seller"`
}
type ProductAttrForAdminChecker struct {
	Id    primitive.ObjectID `json:"attr_id" bson:"attr_id"`
	Value interface{}        `json:"value" bson:"value"`
	// IsVisible  bool               `json:"is_visible" bson:"is_visible"`
	UnitIndex  int64 `json:"unit_index" bson:"unit_index"`
	AttrDetail struct {
		Id          primitive.ObjectID `json:"_id" bson:"_id"`
		IsNumber    bool               `json:"is_number" bson:"is_number"`
		Name        string             `json:"name" bson:"name"`
		Placeholder string             `json:"placeholder" bson:"placeholder"`
		Units       []string           `json:"units_array" bson:"units_array"`
	} `json:"attr_detail" bson:"attr_detail"`
}
type ProductCatForAdminChecker struct {
	Id     primitive.ObjectID `json:"_id" bson:"_id"`
	Name   string             `json:"name" bson:"name"`
	Parent struct {
		Id   primitive.ObjectID `json:"_id" bson:"_id"`
		Name string             `json:"name" bson:"name"`
	} `json:"parent" bson:"parent"`
}
type ProductBrandForAdminChecker struct {
	Id     primitive.ObjectID `json:"_id" bson:"_id"`
	Name   string             `json:"name" bson:"name"`
	Parent struct {
		Id   primitive.ObjectID `json:"_id" bson:"_id"`
		Name string             `json:"name" bson:"name"`
	} `json:"parent" bson:"parent"`
}
type SellerForAdminChecker struct {
	Id   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name" bson:"name"`
	Logo string             `json:"logo" bson:"logo"`
	Type string             `json:"type" bson:"type"`
	City struct {
		Id   primitive.ObjectID `json:"_id" bson:"_id"`
		Name string             `json:"name" bson:"name"`
	} `json:"city" bson:"city"`
}

type DiscountData struct {
	Percent      float64 `json:"percent" bson:"percent"`
	Price        float64 `json:"price" bson:"price"`
	Type         string  `json:"type" bson:"type"`
	Duration     int64   `json:"duration" bson:"duration"`
	DurationType string  `json:"duration_type" bson:"duration_type"`
}
type NewProductForAutoPost struct {
	Id     primitive.ObjectID `bson:"_id"`
	Image  string             `bson:"image"`
	Seller struct {
		Id   primitive.ObjectID `bson:"_id"`
		Name string             `bson:"name"`
		Logo string             `bson:"logo"`
		City CityUpsert         `bson:"city"`
	} `bson:"seller"`
}
type RelatedProductInfo struct {
	Id      primitive.ObjectID `json:"_id" bson:"_id"`
	Heading string             `json:"heading" bson:"heading"`
	Image   string             `json:"image" bson:"image"`
}
