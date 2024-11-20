package statisticsservice

import (
	"context"
	"errors"
	"time"

	"github.com/devzatruk/bizhubBackend/models"
	"github.com/devzatruk/bizhubBackend/ojologger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StatisticsServiceReaderManager struct {
	service *StatisticsService
	logger  *ojologger.OjoLogGroup
}

var (
	InValidType = errors.New("invalid type")
)

type StatisticsServiceReader struct {
	logger       *ojologger.OjoLogGroup
	service      *StatisticsService
	DurationType string
}

func (r *StatisticsServiceReaderManager) run() {
	log := r.logger.Group("run()")

	log.Logf("StatisticsServiceReader started at: %v", time.Now())
}

func (r *StatisticsServiceReaderManager) isValidType(t string) bool {
	types := []string{"today", "yesterday", "month"}
	for _, v := range types {
		if v == t {
			return true
		}
	}
	return false
}

func (r *StatisticsServiceReaderManager) New(t string) (*StatisticsServiceReader, error) {
	log := r.logger.Group("New()")
	if !r.isValidType(t) {
		log.Error(InValidType)
		return nil, InValidType
	}

	reader := &StatisticsServiceReader{
		logger:       r.logger.Group("StatisticsServiceReader"),
		DurationType: t,
		service:      r.service,
	}

	return reader, nil
}

// reader

type StatisticReportUsersActivity struct {
	AllUsers         int64                          `json:"all_users" bson:"all_users"`
	ActiveUsers      int64                          `json:"active_users" bson:"active_users"`
	DeletedUsers     int64                          `json:"deleted_users" bson:"deleted_users"`
	ActiveDifference StatisticDifference            `json:"active_difference" bson:"active_difference"`
	GraphTable       []StatisticReportActivityGraph `json:"graph_table" bson:"graph_table"`
}

type StatisticReportSellersActivity struct {
	AllSellers       int64                          `json:"all_sellers" bson:"all_sellers"`
	ActiveSellers    int64                          `json:"active_sellers" bson:"active_sellers"`
	DeletedSellers   int64                          `json:"deleted_sellers" bson:"deleted_sellers"`
	ActiveDifference StatisticDifference            `json:"active_difference" bson:"active_difference"`
	GraphTable       []StatisticReportActivityGraph `json:"graph_table" bson:"graph_table"`
}

type StatisticReportActivityGraph struct {
	Value  int   `json:"value" bson:"value"`
	All    int64 `json:"all" bson:"all"`
	Active int64 `json:"active" bson:"active"`
}

type StatisticReportPublishedProductsAndPosts struct {
	Products int64 `json:"products" bson:"products"`
	Posts    int64 `json:"posts" bson:"posts"`
}

type StatisticReportMoneyActivity struct {
	Deposited float64 `json:"deposited" bson:"deposited"`
	Withdrew  float64 `json:"withdrew"  bson:"withdrew"`
	Total     float64 `json:"total" bson:"total"`
}

type StatisticReportExpensesActivity struct {
	Total float64            `json:"total" bson:"total"`
	List  []StatisticExpense `json:"list" bson:"list"`
}

type StatisticReportUsersAndSellersActivityDetail struct {
	Date             time.Time           `json:"date" bson:"date"`
	ActiveDifference StatisticDifference `json:"active_difference" bson:"active_difference"`
	Active           int64               `json:"active" bson:"active"`
}

type StatisticReportMoneyActivityDetail struct {
	Date      time.Time `json:"date" bson:"date"`
	Deposited float64   `json:"deposited" bson:"deposited"`
	Withdrew  float64   `json:"withdrew" bson:"withdrew"`
	Total     float64   `json:"total" bson:"total"`
}

type StatisticReportPublishedProductsAndPostsDetail struct {
	Date     time.Time `json:"date" bson:"date"`
	Products int64     `json:"products" bson:"products"`
	Posts    int64     `json:"posts" bson:"posts"`
}

type StatisticReportEmployeeActivity struct {
	Id                  primitive.ObjectID `json:"_id" bson:"_id"`
	FullName            string             `json:"full_name" bson:"full_name"`
	Avatar              string             `json:"avatar" bson:"avatar"`
	Job                 models.Job         `json:"job" bson:"job"`
	CompletedTasksCount int64              `json:"completed_tasks_count" bson:"completed_tasks_count"`
	Note                string             `json:"note" bson:"note"`
}

type DateTimeRange struct {
	Start time.Time
	End   time.Time
}

func (r *StatisticsServiceReader) getDateTimes() (time.Time, time.Time, DateTimeRange) {
	now := time.Now()
	y, m, d := now.Date()

	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)

	monthStart := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC) // GMT+5
	monthEnd := time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC)

	return today, yesterday, DateTimeRange{
		Start: monthStart,
		End:   monthEnd,
	}
}

func (r *StatisticsServiceReader) prepareBsonMatchByDurationType() (bson.M, time.Time, time.Time, DateTimeRange) {
	today, yesterday, month := r.getDateTimes()

	match := bson.M{
		"date": today,
	}

	if r.DurationType == "yesterday" {
		match = bson.M{
			"date": yesterday,
		}
	}

	if r.DurationType == "month" {
		match = bson.M{
			"date": bson.M{
				"$gt":  month.Start,
				"$lte": month.End,
			},
		}
	}

	return match, today, yesterday, month
}

func (r *StatisticsServiceReader) UsersActivity() (*StatisticReportUsersActivity, error) {
	ctx := context.Background()

	now := time.Now()

	match, _, _, month := r.prepareBsonMatchByDurationType()

	var usersActivity StatisticReportUsersActivity
	if r.DurationType != "month" {
		aggregate := bson.A{
			bson.M{
				"$match": match,
			},
			bson.M{
				"$limit": 1,
			},
			bson.M{
				"$project": bson.M{
					"all_users":         "$users.all",
					"active_users":      "$users.active",
					"deleted_users":     "$users.deleted",
					"users_detail":      1,
					"active_difference": "$users.active_difference",
				},
			},
			bson.M{
				"$addFields": bson.M{
					"graph_table": bson.M{
						"$function": bson.M{
							"body": `function(arr) {
							arr.sort((a, b) => a.value - b.value);
							return arr;
						}`,
							"args": bson.A{
								bson.M{"$map": bson.M{
									"input": bson.M{"$objectToArray": "$users_detail"},
									"as":    "row",
									"in": bson.M{ // k v
										"$mergeObjects": bson.A{"$$row.v", bson.M{
											"value": "$$row.v.hour",
										}},
									},
								},
								},
							},
							"lang": "js",
						},
					},
				},
			},
		}

		cursor, err := r.service.coll.Aggregate(ctx, aggregate)
		if err != nil {
			return nil, err
		}

		if cursor.Next(ctx) {
			err := cursor.Decode(&usersActivity)
			if err != nil {
				return nil, err
			}
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}
	} else {
		table := []StatisticReportActivityGraph{}
		y, m, d := now.Date()
		totalDaysOfMonth := time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC).Day()

		// default
		usersActivity.ActiveUsers = 0
		usersActivity.AllUsers = 0
		usersActivity.DeletedUsers = 0
		usersActivity.GraphTable = []StatisticReportActivityGraph{}

		for i := 0; i < totalDaysOfMonth; i++ {
			table = append(table, StatisticReportActivityGraph{
				Value:  i + 1,
				All:    0,
				Active: 0,
			})
		}

		graphCursor, err := r.service.coll.Aggregate(ctx, bson.A{
			bson.M{
				"$match": bson.M{
					"date": bson.M{
						"$gt":  month.Start,
						"$lte": month.End,
					},
				},
			},
			bson.M{
				"$addFields": bson.M{
					"date": bson.M{
						"$dateToParts": bson.M{
							"date": "$date",
						},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"all":               "$users.all",
					"active":            "$users.active",
					"value":             "$date.day",
					"deleted":           "$users.deleted",
					"active_difference": "$users.active_difference",
				},
			},
		})
		if err != nil {
			return nil, err
		}

		for graphCursor.Next(ctx) {
			var g struct {
				StatisticReportActivityGraph `bson:",inline"`
				ActiveDifference             StatisticDifference `bson:"active_difference"`
				Deleted                      int64               `bson:"deleted"`
			}
			err := graphCursor.Decode(&g)
			if err != nil {
				return nil, err
			}
			if d == g.Value {
				usersActivity.AllUsers = g.All
				usersActivity.ActiveUsers = g.Active
				usersActivity.ActiveDifference = g.ActiveDifference
			}
			usersActivity.DeletedUsers += g.Deleted
			table[g.Value-1] = g.StatisticReportActivityGraph
		}

		if err := graphCursor.Err(); err != nil {
			return nil, err
		}

		usersActivity.GraphTable = table
	}

	if usersActivity.GraphTable == nil {
		usersActivity.GraphTable = []StatisticReportActivityGraph{}
	}

	return &usersActivity, nil
}

func (r *StatisticsServiceReader) EmployeesActivity(page int, limit int) (*[]StatisticReportEmployeeActivity, error) {
	ctx := context.Background()

	match, _, _, _ := r.prepareBsonMatchByDurationType()

	employeesActivity := []StatisticReportEmployeeActivity{}

	aggregate := bson.A{
		bson.M{
			"$match": bson.M{
				"exited_on": nil,
				"job.name": bson.M{
					"$in": bson.A{"cashier", "admin_checker"},
				},
			},
		},
		bson.M{
			"$skip": page * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"full_name": bson.M{"$concat": bson.A{"$name", " ", "$surname"}},
				"avatar":    1,
				"job":       1,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "everyday_work",
				"localField":   "_id",
				"foreignField": "employee_id",
				"as":           "works",
				"pipeline": bson.A{
					bson.M{
						"$match": match,
					},
					bson.M{
						"$project": bson.M{
							"completed_tasks_count": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"completed_tasks_count": bson.M{
					"$reduce": bson.M{
						"input":        "$works",
						"initialValue": 0,
						"in": bson.M{
							"$add": bson.A{"$$value", "$$this.completed_tasks_count"},
						},
					},
				},
			},
		},
	}

	cursor, err := r.service.employeesColl.Aggregate(ctx, aggregate)
	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		var a StatisticReportEmployeeActivity
		err := cursor.Decode(&a)
		if err != nil {
			return nil, err
		}

		employeesActivity = append(employeesActivity, a)
	}

	return &employeesActivity, nil
}

func (r *StatisticsServiceReader) SellersActivity() (*StatisticReportSellersActivity, error) {
	ctx := context.Background()

	now := time.Now()

	match, _, _, month := r.prepareBsonMatchByDurationType()

	var sellersActivity StatisticReportSellersActivity
	if r.DurationType != "month" {
		aggregate := bson.A{
			bson.M{
				"$match": match,
			},
			bson.M{
				"$limit": 1,
			},
			bson.M{
				"$project": bson.M{
					"all_sellers":       "$sellers.all",
					"active_sellers":    "$sellers.active",
					"deleted_sellers":   "$sellers.deleted",
					"sellers_detail":    1,
					"active_difference": "$sellers.active_difference",
				},
			},
			bson.M{
				"$addFields": bson.M{
					"graph_table": bson.M{
						"$function": bson.M{
							"body": `function(arr) {
							arr.sort((a, b) => a.value - b.value);
							return arr;
						}`,
							"args": bson.A{
								bson.M{"$map": bson.M{
									"input": bson.M{"$objectToArray": "$sellers_detail"},
									"as":    "row",
									"in": bson.M{ // k v
										"$mergeObjects": bson.A{"$$row.v", bson.M{
											"value": "$$row.v.hour",
										}},
									},
								},
								},
							},
							"lang": "js",
						},
					},
				},
			},
		}

		cursor, err := r.service.coll.Aggregate(ctx, aggregate)
		if err != nil {
			return nil, err
		}

		if cursor.Next(ctx) {
			err := cursor.Decode(&sellersActivity)
			if err != nil {
				return nil, err
			}
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}
	} else {
		table := []StatisticReportActivityGraph{}
		y, m, d := now.Date()
		totalDaysOfMonth := time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC).Day()

		// default
		sellersActivity.ActiveSellers = 0
		sellersActivity.AllSellers = 0
		sellersActivity.DeletedSellers = 0
		sellersActivity.GraphTable = []StatisticReportActivityGraph{}

		for i := 0; i < totalDaysOfMonth; i++ {
			table = append(table, StatisticReportActivityGraph{
				Value:  i + 1,
				All:    0,
				Active: 0,
			})
		}

		graphCursor, err := r.service.coll.Aggregate(ctx, bson.A{
			bson.M{
				"$match": bson.M{
					"date": bson.M{
						"$gt":  month.Start,
						"$lte": month.End,
					},
				},
			},
			bson.M{
				"$addFields": bson.M{
					"date": bson.M{
						"$dateToParts": bson.M{
							"date": "$date",
						},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"all":               "$sellers.all",
					"active":            "$sellers.active",
					"value":             "$date.day",
					"deleted":           "$sellers.deleted",
					"active_difference": "$sellers.active_difference",
				},
			},
		})
		if err != nil {
			return nil, err
		}

		for graphCursor.Next(ctx) {
			var g struct {
				StatisticReportActivityGraph `bson:",inline"`
				ActiveDifference             StatisticDifference `bson:"active_difference"`
				Deleted                      int64               `bson:"deleted"`
			}
			err := graphCursor.Decode(&g)
			if err != nil {
				return nil, err
			}
			if d == g.Value {
				sellersActivity.AllSellers = g.All
				sellersActivity.ActiveSellers = g.Active
				sellersActivity.ActiveDifference = g.ActiveDifference
			}
			sellersActivity.DeletedSellers += g.Deleted
			table[g.Value-1] = g.StatisticReportActivityGraph
		}

		if err := graphCursor.Err(); err != nil {
			return nil, err
		}

		sellersActivity.GraphTable = table
	}

	if sellersActivity.GraphTable == nil {
		sellersActivity.GraphTable = []StatisticReportActivityGraph{}
	}

	return &sellersActivity, nil
}

func (r *StatisticsServiceReader) PublishedProductsAndPosts() (*StatisticReportPublishedProductsAndPosts, error) {
	// log := r.logger.Group("PublishedProductsAndPosts()")

	ctx := context.Background()
	match, _, _, _ := r.prepareBsonMatchByDurationType()
	var report StatisticReportPublishedProductsAndPosts

	if r.DurationType != "month" {
		cursor, err := r.service.coll.Aggregate(ctx, bson.A{
			bson.M{
				"$match": match,
			},
			bson.M{
				"$limit": 1,
			},
			bson.M{
				"$project": bson.M{
					"products": "$published_products",
					"posts":    "$published_posts",
				},
			},
		})
		if err != nil {
			return nil, err
		}

		if cursor.Next(ctx) {
			err := cursor.Decode(&report)
			if err != nil {
				return nil, err
			}
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}
	} else {
		cursor, err := r.service.coll.Aggregate(ctx, bson.A{
			bson.M{
				"$match": match,
			},
			bson.M{
				"$project": bson.M{
					"products": "$published_products",
					"posts":    "$published_posts",
				},
			},
		})
		if err != nil {
			return nil, err
		}

		for cursor.Next(ctx) {
			var dailyReport StatisticReportPublishedProductsAndPosts
			err := cursor.Decode(&dailyReport)
			if err != nil {
				return nil, err
			}

			report.Posts += dailyReport.Posts
			report.Products += dailyReport.Products
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}
	}

	return &report, nil
}

func (r *StatisticsServiceReader) MoneyActivity() (*StatisticReportMoneyActivity, error) {
	// log := r.logger.Group("PublishedProductsAndPosts()")

	ctx := context.Background()
	match, _, _, _ := r.prepareBsonMatchByDurationType()
	var moneyActivity StatisticReportMoneyActivity

	if r.DurationType != "month" {
		cursor, err := r.service.coll.Aggregate(ctx, bson.A{
			bson.M{
				"$match": match,
			},
			bson.M{
				"$limit": 1,
			},
			bson.M{
				"$project": bson.M{
					"deposited": "$money.deposited",
					"withdrew":  "$money.withdrew",
					"total":     "$money.total",
				},
			},
		})
		if err != nil {
			return nil, err
		}

		if cursor.Next(ctx) {
			err := cursor.Decode(&moneyActivity)
			if err != nil {
				return nil, err
			}
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}
	} else {
		cursor, err := r.service.coll.Aggregate(ctx, bson.A{
			bson.M{
				"$match": match,
			},
			bson.M{
				"$project": bson.M{
					"deposited": "$money.deposited",
					"withdrew":  "$money.withdrew",
					"total":     "$money.total",
					"value_": bson.M{
						"$dateToParts": bson.M{
							"date": "$date",
						},
					},
				},
			},
			bson.M{
				"$addFields": bson.M{
					"value": "$value_.day",
				},
			},
		})
		if err != nil {
			return nil, err
		}

		_, _, d := time.Now().Date()

		for cursor.Next(ctx) {
			var activity struct {
				StatisticReportMoneyActivity `bson:",inline"`
				Value                        int `bson:"value"`
			}

			err := cursor.Decode(&activity)
			if err != nil {
				return nil, err
			}

			if d == activity.Value {
				moneyActivity.Total = activity.Total
			}

			moneyActivity.Deposited += activity.Deposited
			moneyActivity.Withdrew += activity.Withdrew
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}
	}

	return &moneyActivity, nil
}

func (r *StatisticsServiceReader) ExpensesActivity() (*StatisticReportExpensesActivity, error) {
	// log := r.logger.Group("ExpensesActivity()")

	match, _, _, _ := r.prepareBsonMatchByDurationType()

	ctx := context.Background()
	var activity StatisticReportExpensesActivity

	cursor, err := r.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": match,
		},
		bson.M{
			"$project": bson.M{
				"expenses": 1,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	for cursor.Next(ctx) {
		var row struct {
			Expenses []StatisticExpense `bson:"expenses"`
		}

		err := cursor.Decode(&row)
		if err != nil {
			return nil, err
		}

		t := 0.0
		for _, v := range row.Expenses {
			t += v.Amount
		}

		activity.Total += t
		activity.List = append(activity.List, row.Expenses...)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if activity.List == nil {
		activity.List = []StatisticExpense{}
	}

	return &activity, nil
}

// detail

func (r *StatisticsServiceReader) UsersActivityDetail(page int, limit int) ([]StatisticReportUsersAndSellersActivityDetail, error) {

	var data []StatisticReportUsersAndSellersActivityDetail
	ctx := context.Background()

	cursor, err := r.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$skip": page * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"date":              1,
				"active_difference": "$users.active_difference",
				"active":            "$users.active",
			},
		},
	})
	if err != nil {
		return []StatisticReportUsersAndSellersActivityDetail{}, err
	}

	for cursor.Next(ctx) {
		var row StatisticReportUsersAndSellersActivityDetail
		err := cursor.Decode(&row)
		if err != nil {
			return []StatisticReportUsersAndSellersActivityDetail{}, err
		}
		data = append(data, row)
	}

	if err := cursor.Err(); err != nil {
		return []StatisticReportUsersAndSellersActivityDetail{}, err
	}

	if data == nil {
		data = []StatisticReportUsersAndSellersActivityDetail{}
	}

	return data, nil
}

func (r *StatisticsServiceReader) SellersActivityDetail(page int, limit int) ([]StatisticReportUsersAndSellersActivityDetail, error) {

	var data []StatisticReportUsersAndSellersActivityDetail
	ctx := context.Background()

	cursor, err := r.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$skip": page * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"date":              1,
				"active_difference": "$sellers.active_difference",
				"active":            "$sellers.active",
			},
		},
	})
	if err != nil {
		return []StatisticReportUsersAndSellersActivityDetail{}, err
	}

	for cursor.Next(ctx) {
		var row StatisticReportUsersAndSellersActivityDetail
		err := cursor.Decode(&row)
		if err != nil {
			return []StatisticReportUsersAndSellersActivityDetail{}, err
		}
		data = append(data, row)
	}

	if err := cursor.Err(); err != nil {
		return []StatisticReportUsersAndSellersActivityDetail{}, err
	}

	if data == nil {
		data = []StatisticReportUsersAndSellersActivityDetail{}
	}

	return data, nil
}

func (r *StatisticsServiceReader) PublishedProductsAndPostsDetail(page int, limit int) ([]StatisticReportPublishedProductsAndPostsDetail, error) {
	var data []StatisticReportPublishedProductsAndPostsDetail
	ctx := context.Background()

	cursor, err := r.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$skip": page * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"date":     1,
				"products": "$published_products",
				"posts":    "$published_posts",
			},
		},
	})
	if err != nil {
		return []StatisticReportPublishedProductsAndPostsDetail{}, err
	}

	for cursor.Next(ctx) {
		var row StatisticReportPublishedProductsAndPostsDetail
		err := cursor.Decode(&row)
		if err != nil {
			return []StatisticReportPublishedProductsAndPostsDetail{}, err
		}
		data = append(data, row)
	}

	if err := cursor.Err(); err != nil {
		return []StatisticReportPublishedProductsAndPostsDetail{}, err
	}

	if data == nil {
		data = []StatisticReportPublishedProductsAndPostsDetail{}
	}

	return data, nil
}

func (r *StatisticsServiceReader) MoneyActivityDetail(page int, limit int) ([]StatisticReportMoneyActivityDetail, error) {

	var data []StatisticReportMoneyActivityDetail
	ctx := context.Background()

	cursor, err := r.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$skip": page * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"date":      1,
				"deposited": "$money.deposited",
				"withdrew":  "$money.withdrew",
				"total":     "$money.total",
			},
		},
	})
	if err != nil {
		return []StatisticReportMoneyActivityDetail{}, err
	}

	for cursor.Next(ctx) {
		var row StatisticReportMoneyActivityDetail
		err := cursor.Decode(&row)
		if err != nil {
			return []StatisticReportMoneyActivityDetail{}, err
		}
		data = append(data, row)
	}

	if err := cursor.Err(); err != nil {
		return []StatisticReportMoneyActivityDetail{}, err
	}

	if data == nil {
		data = []StatisticReportMoneyActivityDetail{}
	}

	return data, nil
}
