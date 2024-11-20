package favorite_sellers

import (
	"context"
	"fmt"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SellerCust struct {
	SellerId   primitive.ObjectID `bson:"seller_id"`
	CustomerId primitive.ObjectID `bson:"customer_id"`
}
type FavSellersManager struct {
	List map[SellerCust]bool
}
type FavoriteSeller struct {
	SellerId string `json:"seller_id" bson:"seller_id"`
	Liked    bool   `json:"liked"`
}

func (f *FavSellersManager) GetBackup() map[SellerCust]bool {
	copy := make(map[SellerCust]bool, len(f.List))
	for key := range f.List {
		copy[key] = f.List[key]
	}
	return copy
}
func (f *FavSellersManager) AddFromSlice(customerId primitive.ObjectID, s []FavoriteSeller) error {
	backupList := f.GetBackup()
	for i := 0; i < len(s); i++ {
		item := s[i]
		sellerId, err := primitive.ObjectIDFromHex(item.SellerId)
		if err != nil {
			f.List = backupList
			return err
		}

		key := SellerCust{SellerId: sellerId, CustomerId: customerId}
		if _, ok := f.List[key]; ok {
			if f.List[key] == true && item.Liked == false {
				delete(f.List, key)
			} else if f.List[key] == false && item.Liked == true {
				delete(f.List, key)
			}
		} else {
			f.List[key] = item.Liked
		}
	}
	return nil
}

func (f *FavSellersManager) Init() {
	s := gocron.NewScheduler(time.Local)
	s.Every("5s").Do(f.Save)
	s.StartAsync()
}
func (f *FavSellersManager) Clear() {
	f.List = make(map[SellerCust]bool)
}

// func (f *FavSellersManager) Save() {
// 	fmt.Println("FavSellerManager.Save--------------------------->")
// 	i := 1
// 	for next := range f.List {
// 		fmt.Printf("\n%d - %v - %v\n", i, next, f.List[next])
// 		i++
// 	}
// 	fmt.Println("FavSellerManager.Save--------------------------->")
// 	// f.Clear()
// }
func (f *FavSellersManager) Save() {
	fmt.Println("\nFavSellerManager.Save-------> started ->")
	fav_sellers := config.MI.DB.Collection("favorite_sellers")
	toBeDeleted := []SellerCust{}
	toBeInserted := bson.A{}

	for key, value := range f.List {
		if value == true {
			toBeInserted = append(toBeInserted, bson.M{
				"seller_id":   key.SellerId,
				"customer_id": key.CustomerId,
				"created_at":  time.Now().Local(),
			})
		} else {
			toBeDeleted = append(toBeDeleted, SellerCust{SellerId: key.SellerId, CustomerId: key.CustomerId})
		}

	}
	fmt.Println(toBeInserted)
	if len(toBeInserted) != 0 {
		_, err := fav_sellers.InsertMany(context.Background(), toBeInserted)
		if err != nil {
			fmt.Println("\nFavSellerManager.Save - toBeInserted - error -", err.Error())
		}
	}
	if len(toBeDeleted) != 0 {
		for _, row := range toBeDeleted {
			_, err := fav_sellers.DeleteOne(context.Background(), bson.M{
				"seller_id":   row.SellerId,
				"customer_id": row.CustomerId,
			})
			if err != nil {
				fmt.Println("FavSellerManager.Save - row=", row, "- toBeDeleted - error -", err.Error())
			}
		}
	}

	sellers := config.MI.DB.Collection("sellers")
	for key, value := range f.List {
		like := 1
		if value == false {
			like = -1
		}
		sellers.FindOneAndUpdate(context.Background(), bson.M{
			"_id": key.SellerId,
		}, bson.M{
			"$inc": bson.M{
				"likes": like,
			},
		})
	}

	f.Clear()

	fmt.Println("\nFavSellerManager.Save-------> finished ->")
}

// func InitSellerManager() *FavSellersManager {
// 	manager := FavSellersManager{List: make(map[SellerCust]bool)}
// 	manager.Init()
// 	return &manager
// }

var FavSellerManager = FavSellersManager{List: make(map[SellerCust]bool)}

/**
sellers := [{seller_id-customer_id: true}, {..}]

*/
// her 5 sekuntdan bir gezek favorite sellers listesini consola yazsyn
