package statisticsservice

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StatisticMoney struct {
	Total     float64 `json:"total" bson:"total"`
	Deposited float64 `json:"deposited" bson:"deposited"`
	Withdrew  float64 `json:"withdrew" bson:"withdrew"`
}
type StatisticExpense struct {
	Id     primitive.ObjectID `json:"_id" bson:"_id" mapstructure:"_id"`
	Date   time.Time          `json:"date" bson:"date" mapstructure:"date"`
	Amount float64            `json:"amount" bson:"amount" mapstructure:"amount"`
	Note   string             `json:"note" bson:"note" mapstructure:"note"`
}

type StatisticDifference struct {
	Up   int64 `json:"up" bson:"up"`
	Down int64 `json:"down" bson:"down"`
}

type StatisticUsers struct {
	All     int64 `json:"all" bson:"all"`
	Active  int64 `json:"active" bson:"active"`
	Deleted int64 `json:"deleted" bson:"deleted"`
}
type StatisticSellers struct {
	All     int64 `json:"all" bson:"all"`
	Active  int64 `json:"active" bson:"active"`
	Deleted int64 `json:"deleted" bson:"deleted"`
}

type StatisticDetailByHour[T any] struct {
	Detail T   `json:",inline" bson:",inline"`
	Hour   int `json:"hour" bson:"hour"`
}

type StatisticActiveDifferenceWith[T any] struct {
	Detail           T                   `json:",inline" bson:",inline"`
	ActiveDifference StatisticDifference `json:"active_difference" bson:"active_difference"`
}

type Statistic struct {
	Id                primitive.ObjectID                              `json:"_id" bson:"_id,omitempty"`
	Date              time.Time                                       `json:"date" bson:"date"`
	Money             StatisticMoney                                  `json:"money" bson:"money"`
	PublishedPosts    int64                                           `json:"published_posts" bson:"published_posts"`
	PublishedProducts int64                                           `json:"published_products" bson:"published_products"`
	Expenses          []StatisticExpense                              `json:"expenses" bson:"expenses"`
	Users             StatisticActiveDifferenceWith[StatisticUsers]   `json:"users" bson:"users"`
	UsersDetail       map[int]StatisticDetailByHour[StatisticUsers]   `json:"users_detail" bson:"users_detail"`
	Sellers           StatisticActiveDifferenceWith[StatisticSellers] `json:"sellers" bson:"sellers"`
	SellersDetail     map[int]StatisticDetailByHour[StatisticSellers] `json:"sellers_detail" bson:"sellers_detail"`
}
