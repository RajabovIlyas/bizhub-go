package favorite_products

import (
	"context"
	"fmt"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductCust struct {
	ProductId  primitive.ObjectID `bson:"product_id"`
	CustomerId primitive.ObjectID `bson:"customer_id"`
}
type FavProductsManager struct {
	List map[ProductCust]bool
}
type FavoriteProduct struct {
	ProductId string `json:"product_id" bson:"product_id"`
	Liked     bool   `json:"liked"`
}

func (f *FavProductsManager) GetBackup() map[ProductCust]bool {
	copy := make(map[ProductCust]bool, len(f.List))
	for key := range f.List {
		copy[key] = f.List[key]
	}
	return copy
}
func (f *FavProductsManager) AddFromSlice(customerId primitive.ObjectID, s []FavoriteProduct) error {
	backupList := f.GetBackup()
	for i := 0; i < len(s); i++ {
		item := s[i]
		productId, err := primitive.ObjectIDFromHex(item.ProductId)
		if err != nil {
			f.List = backupList
			return err
		}

		key := ProductCust{ProductId: productId, CustomerId: customerId}
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

func (f *FavProductsManager) Init() {
	s := gocron.NewScheduler(time.Local)
	s.Every("5s").Do(f.Save)
	s.StartAsync()
}
func (f *FavProductsManager) Clear() {
	f.List = make(map[ProductCust]bool)
}

// func (f *FavProductsManager) Save() {
// 	fmt.Println("FavProductManager.Save--------------------------->")
// 	i := 1
// 	for next := range f.List {
// 		fmt.Printf("\n%d - %v - %v\n", i, next, f.List[next])
// 		i++
// 	}
// 	fmt.Println("FavProductManager.Save--------------------------->")
// 	// f.Clear()
// }
func (f *FavProductsManager) Save() {
	fmt.Println("FavProductManager.Save-------> started ->")
	fav_products := config.MI.DB.Collection("favorite_products")
	toBeDeleted := []ProductCust{}
	toBeInserted := bson.A{}

	for key, value := range f.List {
		if value == true {
			toBeInserted = append(toBeInserted, bson.M{
				"product_id":  key.ProductId,
				"customer_id": key.CustomerId,
				"created_at":  time.Now().Local(),
			})
		} else {
			toBeDeleted = append(toBeDeleted, ProductCust{ProductId: key.ProductId, CustomerId: key.CustomerId})
		}

	}
	fmt.Println(toBeInserted)
	if len(toBeInserted) != 0 {
		_, err := fav_products.InsertMany(context.Background(), toBeInserted)
		if err != nil {
			fmt.Println("FavProductManager.Save - toBeInserted - error -", err.Error())
		}
	}
	if len(toBeDeleted) != 0 {
		for _, row := range toBeDeleted {
			_, err := fav_products.DeleteOne(context.Background(), bson.M{
				"product_id":  row.ProductId,
				"customer_id": row.CustomerId,
			})
			if err != nil {
				fmt.Println("FavProductManager.Save - row=", row, "- toBeDeleted - error -", err.Error())
			}
		}
	}

	products := config.MI.DB.Collection("products")
	for key, value := range f.List {
		like := 1
		if value == false {
			like = -1
		}
		products.FindOneAndUpdate(context.Background(), bson.M{
			"_id": key.ProductId,
		}, bson.M{
			"$inc": bson.M{
				"likes": like,
			},
		})
	}

	f.Clear()

	fmt.Println("FavProductManager.Save-------> finished ->")
}

// func InitProductManager() *FavProductsManager {
// 	manager := FavProductsManager{List: make(map[ProductCust]bool)}
// 	manager.Init()
// 	return &manager
// }

var FavProductManager = FavProductsManager{List: make(map[ProductCust]bool)}

/**
products := [{product_id-customer_id: true}, {..}]

*/
// her 5 sekuntdan bir gezek favorite products listesini consola yazsyn
