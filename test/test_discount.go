package test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/devzatruk/bizhubBackend/ojocronservice"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// nezaman discount etmeli
// hour, day, month

type ProductPayload struct {
	ProductId primitive.ObjectID `mapstructure:"product_id"`
	Headings  models.Translation `mapstructure:"headings"`
}

type ProductForDiscount struct {
	Id       primitive.ObjectID `mapstructure:"_id"`
	Discount float64            `mapstructure:"discount"`
	Seller   struct {
		Id   primitive.ObjectID `mapstructure:"_id"`
		Name string             `mapstructure:"name"`
	} `mapstructure:"seller"`
	Name models.Translation `mapstructure:"name"`
}

type DiscountTime struct {
	Time  time.Time
	Value int64
	Type  string
}

func getProductForDiscount(id primitive.ObjectID) (*ProductForDiscount, error) {
	ctx := context.Background()
	cursor, err := config.MI.DB.Collection("products").Aggregate(ctx, bson.A{
		bson.A{
			bson.M{
				"$match": bson.M{
					"_id": id,
				},
			},
			bson.M{
				"$lookup": bson.M{
					"from":         "sellers",
					"localField":   "seller_id",
					"foreignField": "_id",
					"as":           "seller",
					"pipeline": bson.A{
						bson.M{
							"$project": bson.M{
								"name": 1,
							},
						},
					},
				},
			},
			bson.M{
				"$unwind": bson.M{
					"path": "$seller",
				},
			},
			bson.M{
				"$project": bson.M{
					"name":     "$heading",
					"discount": 1,
					"seller":   1,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var product ProductForDiscount

	if cursor.Next(ctx) {
		err := cursor.Decode(&product)
		if err != nil {
			return nil, err
		}
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return &product, nil
}

func Testttt(duration int64, durationType string, productObjId primitive.ObjectID) error {
	now := time.Now()

	var endDate time.Time

	duration_ := time.Duration(duration)
	day_ := time.Hour * 24
	month_ := day_ * 60

	if durationType == "hour" {
		endDate = now.Add(time.Hour * duration_)
	} else if durationType == "day" {
		endDate = now.Add(day_ * duration_)
	} else if durationType == "month" {
		endDate = now.Add(month_ * duration_)
	} else {
		return errors.New("invalid duration type")
	}

	// start

	times := []DiscountTime{}
	sub := endDate.Sub(now)
	daysCount := math.Floor((sub.Hours() / 24) / 2)
	if sub.Hours() > (24*2) && daysCount > 0 {
		for i := 0; i < int(daysCount); i++ {
			tt := endDate.Add(time.Duration(-(time.Hour * 24) * 2 * time.Duration(i+1)))
			if !(tt.Year() == now.Year() && tt.Month() == now.Month() && tt.Day() == now.Day()) {
				times = append(times, DiscountTime{Time: tt, Value: 2, Type: "day"})
			}
		}
	}
	if sub.Hours() > 3 {
		times = append(times, DiscountTime{Time: endDate.Add(-(time.Hour * 3)), Value: 3, Type: "hour"})
	}
	if sub.Hours() > 1 {
		times = append(times, DiscountTime{Time: endDate.Add(-time.Hour), Value: 1, Type: "hour"})
	}
	// end

	if len(times) == 0 {
		return errors.New("gygyrmaly wagt yogey")
	}

	product, err := getProductForDiscount(productObjId)
	if err != nil {
		return err
	}

	for _, discountTime := range times {
		// prepare payload
		payload := map[string]any{
			"product_id": productObjId,
		}
		heading := map[string]any{}
		daysLeftStr := models.Translation{}
		if discountTime.Type == "hour" {
			if discountTime.Value == 1 {
				daysLeftStr.En = "1 hour left."
				daysLeftStr.Ru = "остался 1 час."
				daysLeftStr.Tm = "1 sagat galdy."
				daysLeftStr.Tr = "1 saat kaldı."
			} else if discountTime.Value == 3 {
				daysLeftStr.En = "3 hours left."
				daysLeftStr.Ru = "осталось 3 часов."
				daysLeftStr.Tm = "3 sagat galdy."
				daysLeftStr.Tr = "3 saat kaldı."
			}
		} else if discountTime.Type == "day" {
			daysLeft := int64(endDate.Sub(discountTime.Time).Hours() / 24)
			daysLeftStr.En = fmt.Sprintf("%v days left.", daysLeft)
			if daysLeft > 1 {
				daysLeftStr.Ru = fmt.Sprintf("осталось %v дней.", daysLeft)
			} else {
				daysLeftStr.Ru = fmt.Sprintf("остался %v день.", daysLeft)
			}
			daysLeftStr.Tm = fmt.Sprintf("%v gün galdy.", daysLeft)
			daysLeftStr.Tr = fmt.Sprintf("%v gün kaldı.", daysLeft)
		}

		heading["en"] = fmt.Sprintf(os.Getenv("AnnounceEn"), product.Seller.Name,
			product.Discount, product.Name.En, daysLeftStr.En)
		heading["tm"] = fmt.Sprintf(os.Getenv("AnnounceTm"), product.Seller.Name,
			product.Name.Tm, product.Discount, daysLeftStr.Tm)
		heading["ru"] = fmt.Sprintf(os.Getenv("AnnounceRu"), product.Seller.Name,
			product.Discount, product.Name.Ru, daysLeftStr.Ru)
		heading["tr"] = fmt.Sprintf(os.Getenv("AnnounceTr"), product.Seller.Name,
			product.Name.Tr, product.Discount, daysLeftStr.Tr)
		payload["headings"] = heading

		// add job
		cronModel := ojocronservice.NewOjoCronJobModel()
		cronModel.ListenerName(config.ADD_DISCOUNT)
		cronModel.Group(productObjId)
		cronModel.Payload(payload)
		cronModel.RunAt(discountTime.Time)
		config.OjoCronService.NewJob(cronModel)
	}

	removeJobModel := ojocronservice.NewOjoCronJobModel()
	removeJobModel.Group(productObjId)
	removeJobModel.ListenerName(config.REMOVE_DISCOUNT) // TODO: bu name ucin bos string? // discount_removed bolmaly oydyan???
	removeJobModel.Payload(map[string]interface{}{
		"product_id": productObjId,
		"headings": map[string]any{
			"en": fmt.Sprintf(os.Getenv("AnnounceEnRem"), product.Seller.Name,
				product.Discount, product.Name.En),
			"tm": fmt.Sprintf(os.Getenv("AnnounceTmRem"), product.Seller.Name,
				product.Name.Tm, product.Discount),
			"ru": fmt.Sprintf(os.Getenv("AnnounceRuRem"), product.Seller.Name,
				product.Discount, product.Name.Ru),
			"tr": fmt.Sprintf(os.Getenv("AnnounceTrRem"), product.Seller.Name,
				product.Name.Tr, product.Discount),
		},
	})
	removeJobModel.RunAt(endDate)

	config.OjoCronService.NewJob(removeJobModel)

	return nil

}

func main() {
	now := time.Now()

	fmt.Printf("\nnow => %v\n", now)

	endDate := now.Add((time.Hour * 24) * 3)
	fmt.Printf("\nend => %v\n", endDate)

	// start

	times := []DiscountTime{}
	sub := endDate.Sub(now)
	if sub.Hours() > (24*2) && math.Floor((sub.Hours()/24)/2) > 0 {
		daysCount := math.Floor((sub.Hours() / 24) / 2)
		for i := 0; i < int(daysCount); i++ {
			tt := endDate.Add(time.Duration(-(time.Hour * 24) * 2 * time.Duration(i+1)))
			if !(tt.Year() == now.Year() && tt.Month() == now.Month() && tt.Day() == now.Day()) {
				times = append(times, DiscountTime{Time: tt, Value: 2, Type: "day"})
			}
		}
	}
	if sub.Hours() > 3 {
		times = append(times, DiscountTime{Time: endDate.Add(-(time.Hour * 3)), Value: 3, Type: "hour"})
	}
	if sub.Hours() > 1 {
		times = append(times, DiscountTime{Time: endDate.Add(-time.Hour), Value: 1, Type: "hour"})
	}

	// end

	// log
	fmt.Printf("\nType Value Time\n")
	for _, t := range times {
		fmt.Printf("\n%v %v %v\n", t.Type, t.Value, t.Time)
	}

	fmt.Printf("\njemi gygyrmaly: %v\n", len(times))

}
