package ojocronlisteners

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
	"github.com/devzatruk/bizhubBackend/ojologger"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	)
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

func ScheduleAutoPostAddDiscountTimes(duration int64, durationType string, productObjId primitive.ObjectID) error {
	// fmt.Printf("\n*******discount-schedule*******\n")
	now := time.Now()

	var endDate time.Time

	duration_ := time.Duration(duration)
	day_ := time.Hour * 24
	month_ := day_ * 30

	if durationType == "hour" {
		endDate = now.Add(time.Hour * duration_)
	} else if durationType == "day" {
		endDate = now.Add(day_ * duration_)
	} else if durationType == "month" {
		endDate = now.Add(month_ * duration_)
	} else {
		return errors.New("invalid duration type")
	}
	max_last_day := now.Add(time.Hour * 24 * 60) // 2 aydan kop discount edip bilenoklar!!!
	if endDate.After(max_last_day) {
		endDate = max_last_day
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

	// fmt.Printf("\ntimes: %v\n", times)
	// fmt.Printf("\ntimes length: %v\n", len(times))

	if len(times) == 0 {
		return errors.New("gygyrmaly wagt yogey")
	}

	product, err := getProductForDiscount(productObjId)
	// fmt.Printf("\nproduct: %v\n", product)
	if err != nil {
		return err
	}

	for _, discountTime := range times {
		// fmt.Printf("\n times for loopyna girdim %v\n", i)
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
		// fmt.Printf("\ndiscount time-i schedule edildi\n")
	}

	removeJobModel := ojocronservice.NewOjoCronJobModel()
	removeJobModel.Group(productObjId)
	removeJobModel.ListenerName(config.REMOVE_DISCOUNT)
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

	// fmt.Printf("\n*******discount-schedule*******\n")
	return nil

}

func aScheduleAutoPostAddDiscountTimes(duration int64, durationType string, productObjId primitive.ObjectID) error {
	fmt.Printf("\ninside ScheduleAutoPostTimes...\n")
	if duration < 1 {
		return fmt.Errorf("Duration invalid.")
	}
	now := time.Now()
	y, m, _ := now.Date()
	// currentMonth := int64(m) - 1
	thisMonthDuration := time.Date(y, m+1, 0, 0, 0, 0, 0, time.Local)
	_, _, lastDay := thisMonthDuration.Date()
	// fmt.Printf("\nayyn sonky guni nacesi? %v\n", thisMonthDuration.Format(time.RFC3339))
	if durationType == config.DURATION_MONTH {
		if duration > 2 {
			duration = 2
		}
		if duration == 2 {
			nextMonthDuration := time.Date(y, m+2, 0, 0, 0, 0, 0, time.Local)
			_, _, numOfDays := nextMonthDuration.Date()
			if lastDay+numOfDays > 60 {
				duration = 60 // max 60 days
			} else {
				duration = int64(lastDay) + int64(numOfDays) // 31.jan + 28.feb = 59 days
			}
		} else { // duration 1-den kici bolanok!
			duration = int64(lastDay) // 1 month ~ 30 days
		}
		duration = duration * 24 // cast to hours
	} else if durationType == config.DURATION_DAY {
		if duration > 60 {
			duration = 60
		}
		duration = duration * 24 // cast to hours
	} else if durationType == config.DURATION_HOUR {
		// 1000 sagat beren bolsa name etmeli?
		max_hours := int64(60 * 24) // 60 days * 24 hours
		if duration > max_hours {
			duration = max_hours
		}
	}
	// indi db-e ekle
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": productObjId,
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
	}
	cursor, err := productsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return fmt.Errorf("Aggregate(products) - db error.")
	}
	defer cursor.Close(ctx)
	var product ProductForDiscount
	for cursor.Next(ctx) {
		err = cursor.Decode(&product)
		if err != nil {
			return fmt.Errorf("Decode(product) - %v", err)
		}
		break
	}
	if err = cursor.Err(); err != nil {
		return fmt.Errorf("cursor.Err(product) - %v", err)
	}
	fmt.Printf("\nproduct: %v\n", product)
	var batch []*ojocronservice.OjoCronJobModel
	twoDays := int64(2 * 24) // hours, test-den son ulan
	// twoDays := int64(4)                                       // test maksatly: minute edyas
	// duration = int64(10)                                      // test maksatly: 10 minutda test tamamlansyn
	endTime := now.Add(time.Hour * time.Duration(duration)) // test maksatly: time.Minute edildi, son Hour etmeli
	days := duration / twoDays                              // bitin san beryar, her 2gunden bir auto post ucin
	threeHour := endTime.Add(time.Hour * -3)                // sonky -3h sagatda
	// threeHour := endTime.Add(time.Minute * -3) // test maksatly: sonky -3minute
	oneHour := endTime.Add(time.Hour * -1) // sonky -1h sagatda auto post ucin
	// oneHour := endTime.Add(time.Minute * -1) // test maksatly: sonky -1minute auto post ucin
	remainingTime := duration - days*twoDays // sonky 24 sagatlyk wagty alyas
	// fmt.Printf("\nendTime: %v\n", endTime)
	// fmt.Printf("\nnumber of days: %v\n", days)
	// fmt.Printf("\n-three hours: %v\n", threeHour)
	// fmt.Printf("\n-one hour: %v\n", oneHour)
	// fmt.Printf("\nremaining time: %v\n", remainingTime)
	var daysLeftStr models.Translation
	// var autoPost *ojocronservice.OjoCronJob
	payload := ProductPayload{ProductId: productObjId}

	daysLeft := int64(0)
	if days > 0 {
		numberOfDays := days
		if remainingTime == 0 {
			numberOfDays = days - 1 // eger goni 2d beren bolsa, dine -3h we -1h auto post etmeli
		}
		for i := 1; i <= int(numberOfDays); i++ {
			model := ojocronservice.NewOjoCronJobModel()
			model.Group(productObjId)
			model.ListenerName(config.ADD_DISCOUNT)
			nextTime := now.Add(time.Hour * time.Duration(twoDays*int64(i)))
			// fmt.Printf("\nnext time: %v\n", nextTime.Format(time.RFC3339))
			model.RunAt(nextTime)
			daysLeft = int64(endTime.Sub(nextTime).Hours() / 24)
			daysLeftStr.En = fmt.Sprintf("%v days left.", daysLeft)
			if daysLeft > 1 {
				daysLeftStr.Ru = fmt.Sprintf("осталось %v дней.", daysLeft)
			} else {
				daysLeftStr.Ru = fmt.Sprintf("остался %v день.", daysLeft)
			}
			daysLeftStr.Tm = fmt.Sprintf("%v gün galdy.", daysLeft)
			daysLeftStr.Tr = fmt.Sprintf("%v gün kaldı.", daysLeft)

			payload.Headings.En = fmt.Sprintf(os.Getenv("AnnounceEn"), product.Seller.Name,
				product.Discount, product.Name.En, daysLeftStr.En)
			payload.Headings.Tm = fmt.Sprintf(os.Getenv("AnnounceTm"), product.Seller.Name,
				product.Name.Tm, product.Discount, daysLeftStr.Tm)
			payload.Headings.Ru = fmt.Sprintf(os.Getenv("AnnounceRu"), product.Seller.Name,
				product.Discount, product.Name.Ru, daysLeftStr.Ru)
			payload.Headings.Tr = fmt.Sprintf(os.Getenv("AnnounceTr"), product.Seller.Name,
				product.Name.Tr, product.Discount, daysLeftStr.Tr)

			model.Payload(map[string]interface{}{
				"product_id": payload.ProductId,
				"headings": map[string]interface{}{
					"tm": payload.Headings.Tm,
					"ru": payload.Headings.Ru,
					"en": payload.Headings.En,
					"tr": payload.Headings.Tr,
				},
			})
			batch = append(batch, model)
		}
	}
	if remainingTime >= 3 || remainingTime == 0 {
		model_ := ojocronservice.NewOjoCronJobModel()
		model_.Group(productObjId)
		model_.ListenerName(config.ADD_DISCOUNT)
		daysLeftStr.En = "3 hours left."
		daysLeftStr.Ru = "осталось 3 часов."
		daysLeftStr.Tm = "3 sagat galdy."
		daysLeftStr.Tr = "3 saat kaldı."
		model_.RunAt(threeHour)
		payload.Headings.En = fmt.Sprintf(os.Getenv("AnnounceEn"), product.Seller.Name,
			product.Discount, product.Name.En, daysLeftStr.En)
		payload.Headings.Tm = fmt.Sprintf(os.Getenv("AnnounceTm"), product.Seller.Name,
			product.Name.Tm, product.Discount, daysLeftStr.Tm)
		payload.Headings.Ru = fmt.Sprintf(os.Getenv("AnnounceRu"), product.Seller.Name,
			product.Discount, product.Name.Ru, daysLeftStr.Ru)
		payload.Headings.Tr = fmt.Sprintf(os.Getenv("AnnounceTr"), product.Seller.Name,
			product.Name.Tr, product.Discount, daysLeftStr.Tr)
		model_.Payload(map[string]interface{}{
			"product_id": payload.ProductId,
			"headings": map[string]interface{}{
				"tm": payload.Headings.Tm,
				"ru": payload.Headings.Ru,
				"en": payload.Headings.En,
				"tr": payload.Headings.Tr,
			},
		})
		batch = append(batch, model_)
	}

	model := ojocronservice.NewOjoCronJobModel()
	model.ListenerName(config.ADD_DISCOUNT)
	model.Group(productObjId)

	daysLeftStr.En = "1 hour left."
	daysLeftStr.Ru = "остался 1 час."
	daysLeftStr.Tm = "1 sagat galdy."
	daysLeftStr.Tr = "1 saat kaldı."
	model.RunAt(oneHour)
	payload.Headings.En = fmt.Sprintf(os.Getenv("AnnounceEn"), product.Seller.Name,
		product.Discount, product.Name.En, daysLeftStr.En)
	payload.Headings.Tm = fmt.Sprintf(os.Getenv("AnnounceTm"), product.Seller.Name,
		product.Name.Tm, product.Discount, daysLeftStr.Tm)
	payload.Headings.Ru = fmt.Sprintf(os.Getenv("AnnounceRu"), product.Seller.Name,
		product.Discount, product.Name.Ru, daysLeftStr.Ru)
	payload.Headings.Tr = fmt.Sprintf(os.Getenv("AnnounceTr"), product.Seller.Name,
		product.Name.Tr, product.Discount, daysLeftStr.Tr)
	model.Payload(map[string]interface{}{
		"product_id": payload.ProductId,
		"headings": map[string]interface{}{
			"tm": payload.Headings.Tm,
			"ru": payload.Headings.Ru,
			"en": payload.Headings.En,
			"tr": payload.Headings.Tr,
		},
	})
	batch = append(batch, model)
	// discount removed
	modelR := ojocronservice.NewOjoCronJobModel()
	modelR.Group(productObjId)
	modelR.ListenerName(config.REMOVE_DISCOUNT)
	modelR.RunAt(endTime)
	payload.Headings.En = fmt.Sprintf(os.Getenv("AnnounceEnRem"), product.Seller.Name,
		product.Discount, product.Name.En)
	payload.Headings.Tm = fmt.Sprintf(os.Getenv("AnnounceTmRem"), product.Seller.Name,
		product.Name.Tm, product.Discount)
	payload.Headings.Ru = fmt.Sprintf(os.Getenv("AnnounceRuRem"), product.Seller.Name,
		product.Discount, product.Name.Ru)
	payload.Headings.Tr = fmt.Sprintf(os.Getenv("AnnounceTrRem"), product.Seller.Name,
		product.Name.Tr, product.Discount)
	modelR.Payload(map[string]interface{}{
		"product_id": payload.ProductId,
		"headings": map[string]interface{}{
			"tm": payload.Headings.Tm,
			"ru": payload.Headings.Ru,
			"en": payload.Headings.En,
			"tr": payload.Headings.Tr,
		},
	})
	batch = append(batch, model)

	for _, model := range batch {
		err := config.OjoCronService.NewJob(model)
		if err != nil {
			return fmt.Errorf("InserMany(cron_jobs) - %v", err)
		}
	}
	return nil
}

// func AutoPostAddDiscount(job *Ojo) { // bu dine discount doredilende
// bir yerde error berse, rollback() etmeli,
// hic error yok bolsa, onda TASK-y hem pozmaly!!
func AutoPostAddDiscount(job *ojocronservice.OjoCronJob) {
	logger := ojologger.LoggerService.Logger("AddOjoCronListeners()")
	log := logger.Group("autoPostAddDiscount()")

	var payload ProductPayload
	// productID := job.Payload["product_id"].(primitive.ObjectID)
	err := mapstructure.Decode(job.Payload, &payload)
	if err != nil {
		log.Error(err)
		job.Failed()
		return
	}
	ctx := context.Background()
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	pResult, err := productsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": payload.ProductId,
			},
		},
		bson.M{
			"$project": bson.M{
				"image": bson.M{
					"$first": "$images",
				},
				"seller_id": 1,
			},
		},
	})

	if err != nil {
		log.Error(fmt.Errorf("AutoPost aggregate error: %v", err))
		job.Failed()
		return
	}

	var product struct {
		SellerId primitive.ObjectID `bson:"seller_id"`
		Image    string             `bson:"image"`
	}
	if pResult.Next(ctx) {
		err = pResult.Decode(&product)
		if err != nil {
			log.Error(fmt.Errorf("AutoPost decode error: %v", err))
			job.Failed()
			return
		}
	}

	if err = pResult.Err(); err != nil {
		log.Error(fmt.Errorf("AutoPost result error: %v", err))
		job.Failed()
		return
	}

	post := models.PostUpsert{
		Image:           product.Image,
		SellerId:        product.SellerId,
		Title:           payload.Headings,
		Body:            models.Translation{Tm: "", Ru: "", En: "", Tr: ""},
		RelatedProducts: []primitive.ObjectID{payload.ProductId},
		Viewed:          0,
		Likes:           0,
		Auto:            true,
		Status:          config.STATUS_PUBLISHED,
	}
	postsColl := config.MI.DB.Collection("posts")
	_, err = postsColl.InsertOne(ctx, post)
	if err != nil {
		log.Error(fmt.Errorf("AutoPost create post error: %v", err))
		job.Failed()
		return
	}
	job.Finish()

}
