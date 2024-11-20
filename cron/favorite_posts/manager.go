package favoritePosts

import (
	"context"
	"fmt"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostCust struct {
	PostId     primitive.ObjectID `bson:"post_id"`
	CustomerId primitive.ObjectID `bson:"customer_id"`
}
type FavPostsManager struct {
	List map[PostCust]bool
}
type FavoritePost struct {
	PostId string `json:"post_id" bson:"post_id"`
	Liked  bool   `json:"liked"`
}

func (f *FavPostsManager) GetBackup() map[PostCust]bool {
	copy := make(map[PostCust]bool, len(f.List))
	for key := range f.List {
		copy[key] = f.List[key]
	}
	return copy
}
func (f *FavPostsManager) AddFromSlice(customerId primitive.ObjectID, s []FavoritePost) error {
	backupList := f.GetBackup()
	for i := 0; i < len(s); i++ {
		item := s[i]
		postId, err := primitive.ObjectIDFromHex(item.PostId)
		if err != nil {
			f.List = backupList
			return err // errors.New("Error: post_id to objectId>>>" + err.Error())
		}
		key := PostCust{PostId: postId, CustomerId: customerId}
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

func (f *FavPostsManager) Init() {
	s := gocron.NewScheduler(time.Local)
	s.Every("5s").Do(f.Save)
	s.StartAsync()
}
func (f *FavPostsManager) Clear() {
	f.List = make(map[PostCust]bool)
}

// func (f *FavPostsManager) Save() {
// 	fmt.Println("FavPostManager.Save--------------------------->")
// 	i := 1
// 	for next := range f.List {
// 		fmt.Printf("\n%d - %v - %v\n", i, next, f.List[next])
// 		i++
// 	}
// 	fmt.Println("FavPostManager.Save--------------------------->")
// 	// f.Clear()
// }
func (f *FavPostsManager) Save() {
	fmt.Println("FavPostManager.Save-------> started ->")
	fav_posts := config.MI.DB.Collection("favorite_posts")
	toBeDeleted := []PostCust{}
	toBeInserted := bson.A{}

	for key, value := range f.List {
		if value == true {
			toBeInserted = append(toBeInserted, bson.M{
				"post_id":     key.PostId,
				"customer_id": key.CustomerId,
				"created_at":  time.Now(),
			})
		} else {
			toBeDeleted = append(toBeDeleted, PostCust{PostId: key.PostId, CustomerId: key.CustomerId})
		}

	}
	fmt.Println(toBeInserted)
	if len(toBeInserted) != 0 {
		_, err := fav_posts.InsertMany(context.Background(), toBeInserted)
		if err != nil {
			fmt.Println("FavPostManager.Save - toBeInserted - error -", err.Error())
		}
	}
	if len(toBeDeleted) != 0 {
		for _, row := range toBeDeleted {
			_, err := fav_posts.DeleteOne(context.Background(), bson.M{
				"post_id":     row.PostId,
				"customer_id": row.CustomerId,
			})
			if err != nil {
				fmt.Println("FavPostManager.Save - row=", row, "- toBeDeleted - error -", err.Error())
			}
		}
	}

	posts := config.MI.DB.Collection("posts")
	for key, value := range f.List {
		like := 1
		if value == false {
			like = -1
		}
		posts.FindOneAndUpdate(context.Background(), bson.M{
			"_id": key.PostId,
		}, bson.M{
			"$inc": bson.M{
				"likes": like,
			},
		})
	}

	f.Clear()

	fmt.Println("FavPostManager.Save-------> finished ->")
}

// func InitPostManager() *FavPostsManager {
// 	manager := FavPostsManager{List: make(map[PostCust]bool)}
// 	manager.Init()
// 	return &manager
// }

var FavPostManager = FavPostsManager{List: make(map[PostCust]bool)}

/**
posts := [{post_id-customer_id: true}, {..}]

*/
// her 5 sekuntdan bir gezek favorite posts listesini consola yazsyn
