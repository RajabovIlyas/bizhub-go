package everydayworkservice

import (
	"context"

	"github.com/devzatruk/bizhubBackend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EverydayWorkServiceReader struct {
	client *EverydayWorkServiceClient
}

func (r *EverydayWorkServiceReader) Works(page int, limit int) ([]EverydayWorkWithoutTasks, error) {
	ctx := context.Background()
	cursor, err := r.client.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"employee_id": r.client.employeeId,
			},
		},
		bson.M{
			"$sort": bson.M{
				"date": -1,
			},
		},
		bson.M{
			"$skip": page * limit,
		},
		bson.M{
			"$limit": limit,
		},
	})
	if err != nil {
		return []EverydayWorkWithoutTasks{}, err
	}

	works := []EverydayWorkWithoutTasks{}

	for cursor.Next(ctx) {
		var work EverydayWorkWithoutTasks
		// et
		err := cursor.Decode(&work)
		if err != nil {
			return []EverydayWorkWithoutTasks{}, err
		}

		works = append(works, work)
	}

	if err := cursor.Err(); err != nil {
		return []EverydayWorkWithoutTasks{}, err
	}

	return works, nil

}
func (r *EverydayWorkServiceReader) Work(workId primitive.ObjectID) *EverydayWorkServiceWorkReader {
	workReader := EverydayWorkServiceWorkReader{
		reader: r,
		workId: workId,
	}
	return &workReader
}

type EverydayWorkServiceWorkReader struct {
	workId primitive.ObjectID
	reader *EverydayWorkServiceReader
}

// reader

func (r *EverydayWorkServiceWorkReader) Products(page int, limit int) ([]models.Product, error) {
	ctx := context.Background()

	products := []models.Product{}

	cursor, err := r.reader.client.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": r.workId,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"task_": bson.M{"$slice": bson.A{"$tasks.products", page * limit, limit}},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "products",
				"foreignField": "_id",
				"localField":   "task_",
				"as":           "products",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"image": bson.M{
								"$first": "$images",
							},
							"heading":  "$heading.tm",
							"price":    1,
							"discount": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"products": 1,
			},
		},
	})

	if err != nil {
		return []models.Product{}, err
	}

	if cursor.Next(ctx) {
		var p struct {
			Products []models.Product `bson:"products"`
		}
		err := cursor.Decode(&p)
		if err != nil {
			return []models.Product{}, err
		}

		products = p.Products
	}

	if err := cursor.Err(); err != nil {
		return []models.Product{}, err
	}

	return products, nil
}

func (r *EverydayWorkServiceWorkReader) Posts(page int, limit int) ([]models.Post, error) {
	ctx := context.Background()

	posts := []models.Post{}

	cursor, err := r.reader.client.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": r.workId,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"task_": bson.M{"$slice": bson.A{"$tasks.posts", page * limit, limit}},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "posts",
				"foreignField": "_id",
				"localField":   "task_",
				"as":           "posts",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "sellers",
							"foreignField": "_id",
							"localField":   "seller_id",
							"as":           "seller",
							"pipeline": bson.A{
								bson.M{
									"$lookup": bson.M{
										"from":         "cities",
										"foreignField": "_id",
										"localField":   "city_id",
										"as":           "city",
										"pipeline": bson.A{
											bson.M{
												"$project": bson.M{
													"name": "$name.en",
												},
											},
										},
									},
								},
								bson.M{
									"$unwind": bson.M{
										"path":                       "$city",
										"preserveNullAndEmptyArrays": true,
									},
								},
								bson.M{
									"$project": bson.M{
										"city": 1,
										"name": 1,
										"logo": 1,
										"type": 1,
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
							"seller":    1,
							"image":     1,
							"seller_id": 1,
							"viewed":    1,
							"likes":     1,
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"posts": 1,
			},
		},
	})

	if err != nil {
		return []models.Post{}, err
	}

	if cursor.Next(ctx) {
		var p struct {
			Posts []models.Post `bson:"posts"`
		}
		err := cursor.Decode(&p)
		if err != nil {
			return []models.Post{}, err
		}

		posts = p.Posts
	}

	if err := cursor.Err(); err != nil {
		return []models.Post{}, err
	}

	return posts, nil
}

func (r *EverydayWorkServiceWorkReader) SellerProfiles(page int, limit int) ([]models.Seller, error) {
	ctx := context.Background()

	sellers := []models.Seller{}

	cursor, err := r.reader.client.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": r.workId,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"task_": bson.M{"$slice": bson.A{"$tasks.seller_profiles", page * limit, limit}},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "sellers",
				"foreignField": "_id",
				"localField":   "task_",
				"as":           "sellers",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "cities",
							"foreignField": "_id",
							"localField":   "city_id",
							"as":           "city",
							"pipeline": bson.A{
								bson.M{
									"$project": bson.M{
										"name": "$name.en",
									},
								},
							},
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path":                       "city",
							"preserveNullAndEmptyArrays": true,
						},
					},
					bson.M{
						"$project": bson.M{
							"city": 1,
							"name": 1,
							"logo": 1,
							"type": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"sellers": 1,
			},
		},
	})

	if err != nil {
		return []models.Seller{}, err
	}

	if cursor.Next(ctx) {
		var p struct {
			Sellers []models.Seller `bson:"sellers"`
		}
		err := cursor.Decode(&p)
		if err != nil {
			return []models.Seller{}, err
		}

		sellers = p.Sellers
	}

	if err := cursor.Err(); err != nil {
		return []models.Seller{}, err
	}

	return sellers, nil
}

func (r *EverydayWorkServiceWorkReader) Notifications(page int, limit int) ([]models.NotificationForEverydayWork, error) {
	ctx := context.Background()

	notifications := []models.NotificationForEverydayWork{}

	cursor, err := r.reader.client.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": r.workId,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"task_": bson.M{"$slice": bson.A{"$tasks.notifications", page * limit, limit}},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "notifications",
				"foreignField": "_id",
				"localField":   "task_",
				"as":           "notifications",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"audience": 1,
							"text":     "$text.en",
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"notifications": 1,
			},
		},
	})

	if err != nil {
		return []models.NotificationForEverydayWork{}, err
	}

	if cursor.Next(ctx) {
		var p struct {
			Notifications []models.NotificationForEverydayWork `bson:"notifications"`
		}
		err := cursor.Decode(&p)
		if err != nil {
			return []models.NotificationForEverydayWork{}, err
		}

		notifications = p.Notifications
	}

	if err := cursor.Err(); err != nil {
		return []models.NotificationForEverydayWork{}, err
	}

	return notifications, nil
}

func (r *EverydayWorkServiceWorkReader) Auctions(page int, limit int) ([]models.Auction, error) {
	ctx := context.Background()

	auctions := []models.Auction{}

	cursor, err := r.reader.client.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": r.workId,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"task_": bson.M{"$slice": bson.A{"$tasks.auctions", page * limit, limit}},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "auctions",
				"foreignField": "_id",
				"localField":   "task_",
				"as":           "auctions",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"image":       1,
							"heading":     "$heading.en",
							"started_at":  1,
							"finished_at": 1,
							"is_finished": 1,
							"text_color":  1,
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"auctions": 1,
			},
		},
	})

	if err != nil {
		return []models.Auction{}, err
	}

	if cursor.Next(ctx) {
		var p struct {
			Auctions []models.Auction `bson:"auctions"`
		}
		err := cursor.Decode(&p)
		if err != nil {
			return []models.Auction{}, err
		}

		auctions = p.Auctions
	}

	if err := cursor.Err(); err != nil {
		return []models.Auction{}, err
	}

	return auctions, nil
}

func (r *EverydayWorkServiceWorkReader) CashierActivities(page int, limit int) ([]models.CompletedCashierWork, error) {
	ctx := context.Background()

	activities := []models.CompletedCashierWork{}

	cursor, err := r.reader.client.service.coll.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": r.workId,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"task_": bson.M{"$slice": bson.A{"$tasks.cashier_activities", page * limit, limit}},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "cashier_works",
				"foreignField": "_id",
				"localField":   "task_",
				"as":           "activities",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "sellers",
							"localField":   "seller_id",
							"foreignField": "_id",
							"as":           "seller",
							"pipeline": bson.A{
								bson.M{
									"$lookup": bson.M{
										"from":         "cities",
										"localField":   "city_id",
										"foreignField": "_id",
										"as":           "city",
										"pipeline": bson.A{
											bson.M{
												"$project": bson.M{
													"name": "$name.en",
												},
											},
										},
									},
								},
								bson.M{
									"$unwind": bson.M{
										"path": "$city",
									},
								},
								bson.M{
									"$project": bson.M{
										"name": 1,
										"logo": 1,
										"type": 1,
										"city": 1,
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
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"activities": 1,
			},
		},
	})

	if err != nil {
		return []models.CompletedCashierWork{}, err
	}

	if cursor.Next(ctx) {
		var p struct {
			Activities []models.CompletedCashierWork `bson:"activities"`
		}
		err := cursor.Decode(&p)
		if err != nil {
			return []models.CompletedCashierWork{}, err
		}

		activities = p.Activities
	}

	if err := cursor.Err(); err != nil {
		return []models.CompletedCashierWork{}, err
	}

	return activities, nil
}
