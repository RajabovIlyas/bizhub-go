package v1

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/devzatruk/bizhubBackend/config"

	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	ojoTr "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: transaction ulanmalymy? optimize etmelimi?

func DeleteFavorite(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.DeleteFavorite")
	var customerObjId primitive.ObjectID
	err := helpers.GetCurrentCustomer(c, &customerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentCustomer()", err, config.AUTH_REQUIRED))
	}
	var body struct {
		Id   string `json:"_id"`
		Type string `json:"type"`
	}
	err = c.BodyParser(&body)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	bodyId, err := primitive.ObjectIDFromHex(body.Id)
	if err != nil {
		return c.JSON(errRes("ObjectIDFromHex()", err, config.CANT_DECODE))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	switch body.Type {
	case "seller":
		sellersColl := config.MI.DB.Collection(config.SELLERS)
		fav_sellers := config.MI.DB.Collection(config.FAV_SELLERS)
		deleteResult, err := fav_sellers.DeleteOne(ctx, bson.M{"seller_id": bodyId, "customer_id": customerObjId})
		if err != nil {
			return c.JSON(errRes("DeleteOne()", err, config.CANT_DELETE))
		}
		if deleteResult.DeletedCount > 0 {
			_ = sellersColl.FindOneAndUpdate(ctx, bson.M{"_id": bodyId}, bson.M{"$inc": bson.M{"likes": -1}})
		}
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.REMOVED,
		})
	case "post":
		postsColl := config.MI.DB.Collection(config.POSTS)
		fav_posts := config.MI.DB.Collection(config.FAV_POSTS)
		deleteResult, err := fav_posts.DeleteOne(ctx, bson.M{"post_id": bodyId, "customer_id": customerObjId})
		if err != nil {
			return c.JSON(errRes("DeleteOne()", err, config.CANT_DELETE))
		}
		if deleteResult.DeletedCount > 0 {
			_ = postsColl.FindOneAndUpdate(ctx, bson.M{"_id": bodyId}, bson.M{"$inc": bson.M{"likes": -1}})
		}
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.REMOVED,
		})
	case "product":
		prodsColl := config.MI.DB.Collection(config.PRODUCTS)
		fav_prods := config.MI.DB.Collection(config.FAV_PRODS)
		deleteResult, err := fav_prods.DeleteOne(ctx, bson.M{"product_id": bodyId, "customer_id": customerObjId})
		if err != nil {
			return c.JSON(errRes("DeleteOne()", err, config.CANT_DELETE))
		}
		if deleteResult.DeletedCount > 0 {
			_ = prodsColl.FindOneAndUpdate(ctx, bson.M{"_id": bodyId}, bson.M{"$inc": bson.M{"likes": -1}})
		}
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.REMOVED,
		})
	}
	return c.JSON(errRes("Favorite.Type", err, config.TYPE_INVALID))
}

func GetCustomerFavorites(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCustomerFavorites")
	culture := helpers.GetCultureFromQuery(c)
	var customerObjId primitive.ObjectID
	err := helpers.GetCurrentCustomer(c, &customerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentCustomer()", err, config.AUTH_REQUIRED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.CANT_DECODE))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.CANT_DECODE))
	}
	favoriteType := c.Query("type")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	y, m, d := now.Date()
	// today, yesterday, lastWeek, lastMonth gerek bolsa, time.UTC ulansan dogry bolar!!
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)

	if favoriteType == "products" {
		favoriteProductsColl := config.MI.DB.Collection(config.FAV_PRODS)
		favoriteProductsCursor, err := favoriteProductsColl.Aggregate(ctx, bson.A{
			bson.M{
				"$match": bson.M{
					"customer_id": customerObjId,
				},
			},
			bson.M{"$skip": pageIndex * limit},
			bson.M{"$limit": limit},
			bson.M{
				"$lookup": bson.M{
					"from":         "products",
					"localField":   "product_id",
					"foreignField": "_id",
					"as":           "product",
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"status": "published",
							},
						},
						bson.M{
							"$project": bson.M{
								"is_new": bson.M{
									"$gt": bson.A{"$created_at", lastWeek},
								},
								"heading":  culture.Stringf("$heading.%v"),
								"image":    bson.M{"$first": "$images"},
								"price":    1,
								"discount": 1,
								"status":   1,
							},
						},
					},
				},
			},
			bson.M{
				"$unwind": bson.M{"path": "$product"},
			},
			bson.M{
				"$replaceRoot": bson.M{"newRoot": "$product"},
			},
		})
		if err != nil {
			return c.JSON(errRes("Aggregate(fav_products)", err, config.DBQUERY_ERROR))
		}

		var productsResult = []models.Product{}

		for favoriteProductsCursor.Next(ctx) {
			var row models.Product
			err := favoriteProductsCursor.Decode(&row)
			if err != nil {
				return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
			}
			productsResult = append(productsResult, row)
		}

		if err := favoriteProductsCursor.Err(); err != nil {
			return c.JSON(errRes("favoriteProductsCursor.Err()", err, config.DBQUERY_ERROR))
		}
		return c.JSON(models.Response[[]models.Product]{
			IsSuccess: true,
			Result:    productsResult,
		})
	} else if favoriteType == "sellers" {
		favoriteSellersColl := config.MI.DB.Collection(config.FAV_SELLERS)
		favoriteSellersCursor, err := favoriteSellersColl.Aggregate(ctx, bson.A{
			bson.M{
				"$match": bson.M{
					"customer_id": customerObjId,
				},
			},
			bson.M{
				"$skip": pageIndex * limit,
			},
			bson.M{
				"$limit": limit,
			},
			bson.M{
				"$lookup": bson.M{
					"from":         "sellers",
					"localField":   "seller_id",
					"foreignField": "_id",
					"as":           "seller",
					"pipeline": bson.A{
						bson.M{
							"$match": bson.M{
								"status": "published",
							},
						},
						bson.M{
							"$lookup": bson.M{
								"from":         "cities",
								"localField":   "city_id",
								"foreignField": "_id",
								"as":           "city",
								"pipeline": bson.A{
									bson.M{
										"$project": bson.M{
											"name": culture.Stringf("$name.%v"),
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
			bson.M{
				"$replaceRoot": bson.M{
					"newRoot": "$seller",
				},
			},
		})
		if err != nil {
			return c.JSON(errRes("Aggregate(fav_sellers)", err, config.DBQUERY_ERROR))
		}
		var sellersResult = []models.Seller{}

		for favoriteSellersCursor.Next(ctx) {
			var row models.Seller
			err := favoriteSellersCursor.Decode(&row)
			if err != nil {
				return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
			}

			sellersResult = append(sellersResult, row)
		}

		if err := favoriteSellersCursor.Err(); err != nil {
			return c.JSON(errRes("favoriteSellersCursor.Err()", err, config.DBQUERY_ERROR))
		}
		return c.JSON(models.Response[[]models.Seller]{
			IsSuccess: true,
			Result:    sellersResult,
		})

	} else {
		return c.JSON(errRes("NotValid", errors.New("Favorite type invalid."), config.TYPE_INVALID))
	}
}

//TODO: transaction etsek gowy bolar!
func AddFavorite(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("AddFavorite")
	var customerObjId primitive.ObjectID
	err := helpers.GetCurrentCustomer(c, &customerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentCustomer()", err, config.AUTH_REQUIRED))
	}
	var body struct {
		Id   string `json:"_id" bson:"_id"`
		Type string `json:"type" bson:"type"`
	}
	err = c.BodyParser(&body)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	bodyId, err := primitive.ObjectIDFromHex(body.Id)
	if err != nil {
		return c.JSON(errRes("ObjectIDFromHex(id)", err, config.CANT_DECODE))
	}
	hazir := time.Now()
	/* ya db-e oldugu gibi yazmaly we display edende local time donusturmeli,
	yada db-e yazanda time.Add edip yazmaly!
	fmt.Printf("\nlocation: %v and time: %v\n", hazir.Location(), hazir)
	tm_time, err := time.LoadLocation("Asia/Ashgabat")
	fmt.Printf("\nlocation: %v and time: %v\n", tm_time, hazir.In(tm_time))
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	switch body.Type {
	case "seller":
		sellersColl := config.MI.DB.Collection(config.SELLERS)
		fav_sellers := config.MI.DB.Collection(config.FAV_SELLERS)
		_, err := fav_sellers.InsertOne(ctx, bson.M{
			"seller_id": bodyId, "customer_id": customerObjId,
			"created_at": hazir,
		})
		if err != nil {
			return c.JSON(errRes("InsertOne(fav_sellers)", err, config.CANT_INSERT))
		}
		_ = sellersColl.FindOneAndUpdate(ctx, bson.M{"_id": bodyId}, bson.M{"$inc": bson.M{"likes": 1}})
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.ADDED,
		})
	case "post":
		postColl := config.MI.DB.Collection(config.POSTS)
		fav_posts := config.MI.DB.Collection(config.FAV_POSTS)
		_, err := fav_posts.InsertOne(ctx, bson.M{"post_id": bodyId, "customer_id": customerObjId,
			"created_at": hazir,
		})
		if err != nil {
			return c.JSON(errRes("InsertOne(fav_posts)", err, config.CANT_INSERT))
		}
		_ = postColl.FindOneAndUpdate(ctx, bson.M{"_id": bodyId}, bson.M{"$inc": bson.M{"likes": 1}})
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.ADDED,
		})
	case "product":
		productsColl := config.MI.DB.Collection(config.PRODUCTS)
		fav_prods := config.MI.DB.Collection(config.FAV_PRODS)
		_, err := fav_prods.InsertOne(ctx, bson.M{"product_id": bodyId, "customer_id": customerObjId,
			"created_at": hazir,
		})
		if err != nil {
			return c.JSON(errRes("InsertOne(fav_prods)", err, config.CANT_INSERT))
		}
		_ = productsColl.FindOneAndUpdate(ctx, bson.M{"_id": bodyId}, bson.M{"$inc": bson.M{"likes": 1}})
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    config.ADDED,
		})
	}
	return c.JSON(errRes("NotValid", errors.New("Favorite type not valid."), config.TYPE_INVALID))
}

// TODO: su SaveFavorite() gerek dal oydyan!
// func SaveFavorites(c *fiber.Ctx) error {
// 	errRes := helpers.ErrorResponse("SaveFavorites")
// 	var body struct {
// 		// Favorites []favoriteposts.FavoritePost `json:"favorites"`
// 		Favorites struct {
// 			Sellers  []favorite_sellers.FavoriteSeller   `json:"sellers"`
// 			Products []favorite_products.FavoriteProduct `json:"products"`
// 			Posts    []favoritePosts.FavoritePost        `json:"posts"`
// 		}
// 	}
// 	err := c.BodyParser(&body)
// 	if err != nil {
// 		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
// 	}
// 	var customerId primitive.ObjectID
// 	err = helpers.GetCurrentCustomer(c, &customerId)
// 	if err != nil {
// 		return c.JSON(errRes("GetCurrentCustomer()", err, config.CANT_DECODE))
// 	}
// 	err = favoritePosts.FavPostManager.AddFromSlice(customerId, body.Favorites.Posts)
// 	if err != nil {
// 		return c.JSON(errRes("AddFromSlice(FavPostManager)", err, ""))
// 	}
// 	err = favorite_products.FavProductManager.AddFromSlice(customerId, body.Favorites.Products)
// 	if err != nil {
// 		return c.JSON(errRes("AddFromSlice(FavProductManager)", err, ""))
// 	}
// 	err = favorite_sellers.FavSellerManager.AddFromSlice(customerId, body.Favorites.Sellers)
// 	if err != nil {
// 		return c.JSON(errRes("AddFromSlice(FavSellerManager)", err, ""))
// 	}
// 	return c.JSON(models.Response[string]{
// 		IsSuccess: true,
// 		Result:    config.UPDATED,
// 	})
// }

func AuthCustomerLogin(c *fiber.Ctx) error {

	/*
		login ucin
		phone + password
		gelyar
	*/
	errRes := helpers.ErrorResponse("AuthCustomerLogin")
	var credentials models.CustomerCredentials
	err := c.BodyParser(&credentials)
	if err != nil {
		return c.Status(401).JSON(errRes("BodyParser()", err, config.CREDENTIALS_ERROR))
	}
	if !validPhone(credentials.Phone) {
		return c.JSON(errRes("InvalidPhone", fmt.Errorf("Phone number must be 8 digits: %v not valid.", credentials.Phone), config.NOT_ALLOWED))
	}
	is_secure, entropy, err := helpers.IsPasswordSecure(credentials.Password, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("InsecurePassword(%v)", entropy), err, config.INSERCURE_PWD))
	}

	customers := config.MI.DB.Collection(config.CUSTOMERS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	customerCursor := customers.FindOne(ctx, bson.M{
		"phone": credentials.Phone,
	})
	if err = customerCursor.Err(); err != nil {
		return c.Status(401).JSON(errRes("FindOne(customer)", err, config.NOT_FOUND))
	}
	// password valid mi?
	var customer models.CustomerWithPassword
	err = customerCursor.Decode(&customer)
	if err != nil {
		return c.Status(401).JSON(errRes("Decode(customer)", err, config.CANT_DECODE))
	}
	if err = helpers.ComparePassword(customer.Password, credentials.Password); err != nil {
		return c.Status(401).JSON(errRes("ComparePassword()", err, config.CREDENTIALS_ERROR)) // error message uytgetmeli !!!
	}
	// valid user, then create token
	ttl, err := time.ParseDuration(os.Getenv(config.ACCT_EXPIREDIN))
	if err != nil {
		return c.Status(401).JSON(errRes("ParseDuration(acctexpiredin)", err, config.ACCT_TTL_NOT_VALID))
	}
	access_token, err := helpers.CreateToken(ttl, customer.WithoutPassword(), os.Getenv(config.ACCT_PRIVATE_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("CreateToken(access_token)", err, config.ACCT_GENERATION_ERROR))
	}
	ttl, err = time.ParseDuration(os.Getenv(config.REFT_EXPIREDIN))
	if err != nil {
		return c.JSON(errRes("ParseDuration(reftexpiredin)", err, config.REFT_TTL_NOT_VALID))
	}
	refresh_token, err := helpers.CreateToken(ttl, customer.Id, os.Getenv(config.REFT_PRIVATE_KEY))
	if err != nil {
		return c.JSON(errRes("CreateToken(refresh_token)", err, config.REFT_GENERATION_ERROR))
	}
	config.StatisticsService.Writer.NewActiveUser()
	if customer.SellerId != nil {
		config.StatisticsService.Writer.NewActiveSeller()
	}
	favorites := GetFavoritesByCustomer(ctx, customer.Id)
	return c.JSON(models.Response[fiber.Map]{
		IsSuccess: true,
		Result: fiber.Map{
			"access_token":  access_token,
			"refresh_token": refresh_token,
			"user":          customer.WithoutPassword(),
			"favorites":     favorites,
		},
	})
}

func DeleteProfile(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.DeleteProfile")
	var customerObjId, sellerObjId primitive.ObjectID
	err := helpers.GetCurrentCustomer(c, &customerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentCustomer()", err, config.AUTH_REQUIRED))
	}
	is_seller := helpers.IsSeller(c)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	now := time.Now()
	if !is_seller {

		transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
		tr_customersColl := transaction_manager.Collection(config.CUSTOMERS)
		update_model := ojoTr.NewModel().SetFilter(bson.M{"_id": customerObjId}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"status":     config.STATUS_DELETED,
					"updated_at": now,
				}}).
			SetRollbackUpdate(bson.M{
				"$set": bson.M{
					"status":     config.STATUS_PUBLISHED,
					"updated_at": now,
				}})
		_, err = tr_customersColl.FindOneAndUpdate(update_model)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(customer)", err, config.CANT_UPDATE))
		}
		tr_notifTokensColl := transaction_manager.Collection(config.NOTIFICATION_TOKENS)
		update_model = ojoTr.NewModel().
			SetFilter(bson.M{"client_id": customerObjId, "client_type": "customer"}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"client_id":   nil,
					"client_type": nil,
				},
			}).SetRollbackUpdate(bson.M{
			"$set": bson.M{
				"client_id":   customerObjId,
				"client_type": "customer",
			},
		})
		_, err = tr_notifTokensColl.FindOneAndUpdate(update_model)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(notification_token)", err, config.CANT_UPDATE))
		}

		config.StatisticsService.Writer.NewDeletedUser()

		favProdColl := config.MI.DB.Collection(config.FAV_PRODS)
		favProdColl.DeleteMany(ctx, bson.M{"customer_id": customerObjId})
		favPostColl := config.MI.DB.Collection(config.FAV_POSTS)
		favPostColl.DeleteMany(ctx, bson.M{"customer_id": customerObjId})
		favSellersColl := config.MI.DB.Collection(config.FAV_SELLERS)
		favSellersColl.DeleteMany(ctx, bson.M{"customer_id": customerObjId})

		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    "Deleted customer data successfully.",
		})
	}
	if is_seller {
		// sellerObjId, _ = primitive.ObjectIDFromHex("62ce628c8cae982f654a3578")
		err = helpers.GetCurrentSeller(c, &sellerObjId)
		if err != nil {
			return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
		}
		walletsColl := config.MI.DB.Collection(config.WALLETS)
		var wallet models.SellerWallet
		findResult := walletsColl.FindOne(ctx, bson.M{"seller_id": sellerObjId})
		if err = findResult.Err(); err != nil {
			return c.JSON(errRes("FindOne(seller_wallet)", err, config.NOT_FOUND))
		}
		err = findResult.Decode(&wallet)
		if err != nil {
			return c.JSON(errRes("Decode(seller_wallet)", err, config.CANT_DECODE))
		}
		if wallet.Balance > 0 || len(wallet.InAuction) > 0 {
			return c.JSON(errRes("WalletNotEmpty", errors.New("Wallet not empty or some transactions not completed yet."), config.STATUS_ABORTED))
		}
		walletHistColl := config.MI.DB.Collection(config.WALLETHISTORY)
		waitingWalletHistory, err := walletHistColl.CountDocuments(ctx, bson.M{
			"seller_id": sellerObjId,
			"wallet_id": wallet.Id,
			"intent":    config.INTENT_WITHDRAW,
			"status":    config.STATUS_WAITING,
		})
		if err != nil {
			return c.JSON(errRes("CountDocuments(wallet_history)", err, config.CANT_DECODE))
		}
		if waitingWalletHistory > 0 {
			return c.JSON(errRes("WaitingTransactions", errors.New("Withdraw waiting transaction."), config.STATUS_ABORTED))
		}
		// indi seller profile pozup bilersin
		ctx = context.Background() // gyssanman pozaysyn
		has_finished := make(chan bool)

		transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
		tr_customersColl := transaction_manager.Collection(config.CUSTOMERS)
		update_model := ojoTr.NewModel().SetFilter(bson.M{"_id": customerObjId}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"status":     config.SELLER_STATUS_DELETED,
					"updated_at": now,
				},
				"$push": bson.M{
					"deleted_profiles": bson.M{
						"$each": bson.A{sellerObjId},
					},
				},
			}).
			SetRollbackUpdate(bson.M{
				"$set": bson.M{
					"status":     config.SELLER_STATUS_PUBLISHED,
					"updated_at": now,
				},
				"$pull": bson.M{"deleted_profiles": sellerObjId},
			})
		_, err = tr_customersColl.FindOneAndUpdate(update_model)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(customer)", err, config.CANT_UPDATE))
		}

		tr_sellersColl := transaction_manager.Collection(config.SELLERS)
		_, err = tr_sellersColl.FindOneAndDelete(ojoTr.NewModel().SetFilter(bson.M{"_id": sellerObjId}))
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndDelete(seller)", err, config.CANT_DELETE))
		}
		tr_notifTokensColl := transaction_manager.Collection(config.NOTIFICATION_TOKENS)
		update_model = ojoTr.NewModel().
			SetFilter(bson.M{"client_id": sellerObjId, "client_type": "seller"}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"client_id":   nil,
					"client_type": nil,
				},
			}).SetRollbackUpdate(bson.M{
			"$set": bson.M{
				"client_id":   sellerObjId,
				"client_type": "seller",
			},
		})
		_, err = tr_notifTokensColl.FindOneAndUpdate(update_model)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("FindOneAndUpdate(notification_token)", err, config.CANT_UPDATE))
		}

		postsColl := config.MI.DB.Collection(config.POSTS)
		postImagesCursor, err := postsColl.Aggregate(ctx, bson.A{
			bson.M{
				"$match": bson.M{"seller_id": sellerObjId},
			},
			bson.M{
				"$group": bson.M{
					"_id":    bson.M{"seller_id": "$seller_id"},
					"images": bson.M{"$addToSet": "$image"},
				},
			},
			bson.M{
				"$project": bson.M{"_id": 0},
			},
		})
		type images struct {
			Images []string `bson:"images"`
		}
		postImages := images{Images: make([]string, 0)}
		productImages := images{Images: make([]string, 0)}
		if postImagesCursor.Next(ctx) {
			_ = postImagesCursor.Decode(&postImages)
			// if err != nil {
			// 	fmt.Printf("\npost images decode() failed...\n")
			// }
		}
		// if len(postImages.Images) > 0 {
		// 	fmt.Printf("\nfetched post images array %v : %v\n", len(postImages.Images), postImages.Images)
		// }
		productsColl := config.MI.DB.Collection(config.PRODUCTS)
		productImagesCursor, err := productsColl.Aggregate(ctx, bson.A{
			bson.M{
				"$match": bson.M{"seller_id": sellerObjId},
			},
			bson.M{
				"$group": bson.M{
					"_id":    0,
					"images": bson.M{"$push": "$images"},
				},
			},
			bson.M{
				"$project": bson.M{
					"_id": 0,
					"images": bson.M{
						"$reduce": bson.M{
							"input":        "$images",
							"initialValue": bson.A{},
							"in": bson.M{
								"$concatArrays": bson.A{"$$value", "$$this"},
							},
						},
					},
				},
			},
			bson.M{
				"$project": bson.M{
					"images": bson.M{"$setUnion": "$images"},
				},
			},
		})
		if productImagesCursor.Next(ctx) {
			_ = productImagesCursor.Decode(&productImages)
			// if err != nil {
			// 	fmt.Printf("\nproduct images decode() failed...\n")
			// }
		}
		// if len(productImages.Images) > 0 {
		// 	fmt.Printf("\nfetched product images array %v : %v\n", len(productImages.Images), productImages.Images)
		// }
		// start deleting all data...
		go func() {
			feedbackColl := config.MI.DB.Collection(config.FEEDBACKS)
			feedbackColl.DeleteMany(ctx, bson.M{"sent_by": sellerObjId})
			suggestionsColl := config.MI.DB.Collection(config.SUGGESTIONS)
			suggestionsColl.DeleteMany(ctx, bson.M{"seller_id": sellerObjId})
			tasksColl := config.MI.DB.Collection(config.TASKS)
			tasksColl.DeleteMany(ctx, bson.M{"seller_id": sellerObjId})
			packageHistColl := config.MI.DB.Collection(config.PACKAGEHISTORY)
			packageHistColl.DeleteMany(ctx, bson.M{"seller_id": sellerObjId})
			walletHistColl.DeleteMany(ctx, bson.M{"seller_id": sellerObjId})
			walletsColl.DeleteMany(ctx, bson.M{"seller_id": sellerObjId})
			postsColl.DeleteMany(ctx, bson.M{"seller_id": sellerObjId})
			productsColl.DeleteMany(ctx, bson.M{"seller_id": sellerObjId})
			has_finished <- true
		}()
		helpers.DeleteImages(postImages.Images)
		helpers.DeleteImages(productImages.Images)
		// goroutine_finished := <-has_finished
		<-has_finished
		config.StatisticsService.Writer.NewDeletedUser() // NewInactiveUser() logout edende
		config.StatisticsService.Writer.NewDeletedSeller()
		// if goroutine_finished {
		// 	fmt.Printf("\ngoroutine HAS finished deleting data...\n")
		// } else {
		// 	fmt.Printf("\ngoroutine HAS NOT finished deleting data...\n")
		// }
		return c.JSON(models.Response[string]{
			IsSuccess: true,
			Result:    "Deleted seller data successfully.",
		})
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    "profile may be deleted or not...",
	})
}
func Logout(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.Logout")
	var customerObjId, sellerObjId primitive.ObjectID
	err := helpers.GetCurrentCustomer(c, &customerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentCustomer()", err, config.AUTH_REQUIRED))
	}
	is_seller := helpers.IsSeller(c)
	if is_seller {
		err = helpers.GetCurrentSeller(c, &sellerObjId)
		if err != nil {
			return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tokensColl := config.MI.DB.Collection(config.NOTIFICATION_TOKENS)
	filterMap := bson.M{}
	if is_seller {
		filterMap["client_id"] = sellerObjId
		filterMap["client_type"] = "seller"
	} else {
		filterMap["client_id"] = customerObjId
		filterMap["client_type"] = "customer"
	}
	updateMap := bson.M{
		"$set": bson.M{
			"client_id":   nil,
			"client_type": nil,
		},
	}
	updateResult, err := tokensColl.UpdateOne(ctx, filterMap, updateMap)
	if err != nil {
		return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
	}
	if updateResult.ModifiedCount == 0 {
		return c.JSON(errRes("UpdateOne()", errors.New("Couldn't update tokens."), config.CANT_UPDATE))
	}
	config.StatisticsService.Writer.NewInactiveUser()
	if is_seller {
		config.StatisticsService.Writer.NewInactiveSeller()
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    "Logged out",
	})
}
func validPhone(phone string) bool {
	if len(phone) != 8 {
		return false
	}
	digits := strings.Split("0123456789", "")
	for _, char := range phone {
		if !helpers.SliceContains(digits, string(char)) {
			return false
		}
	}
	ikiSifr := string(phone[:2])
	if helpers.SliceContains([]string{"61", "62", "63", "64", "65"}, ikiSifr) {
		return true
	}
	return false
}

// TODO: test Signup()
func Signup(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.Signup()")
	// ilki user-i db ekle, sonra login et, we response iber
	var userInfo models.CustomerWithPassword
	userInfo.Phone = c.FormValue("phone") // if a valid phone
	if !validPhone(userInfo.Phone) {
		return c.JSON(errRes("InvalidPhone", fmt.Errorf("Phone number must be 8 digits: %v not valid.", userInfo.Phone), config.NOT_ALLOWED))
	}
	now := time.Now()
	userInfo.Name = c.FormValue("name")
	userInfo.Password = c.FormValue("password")
	if len(userInfo.Name) == 0 || len(userInfo.Password) == 0 {
		return c.JSON(errRes("HasEmptyFields()", fmt.Errorf("Some data not provided."), config.NOT_ALLOWED))
	}
	is_secure, entropy, err := helpers.IsPasswordSecure(userInfo.Password, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("InsecurePassword(%v)", entropy), err, config.INSERCURE_PWD))
	}

	imagePath, errImageIsDefault := helpers.SaveImageFile(c, "logo", config.FOLDER_USERS)
	if errImageIsDefault != nil {
		imagePath = os.Getenv(config.DEFAULT_USER_IMAGE)
	}
	userInfo.Logo = imagePath
	userInfo.Password = helpers.HashPassword(userInfo.Password)
	userInfo.SellerId = nil

	// if returning customer then update his data else create a new customer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	customersColl := config.MI.DB.Collection(config.CUSTOMERS)
	findResult := customersColl.FindOne(ctx, bson.M{"phone": userInfo.Phone})
	is_new_customer := false
	if err := findResult.Err(); err != nil {
		is_new_customer = true
	}
	if is_new_customer == false {
		var returning_customer models.CustomerForDb
		err := findResult.Decode(&returning_customer)
		if err != nil {
			helpers.DeleteImageFile(userInfo.Logo)
			return c.JSON(errRes("Decode(returning_customer)", err, config.CANT_DECODE))
		}
		if returning_customer.Status != config.STATUS_DELETED {
			helpers.DeleteImageFile(userInfo.Logo)
			return c.JSON(errRes("ExistingCustomer", fmt.Errorf("Phone: %v is already registered.", userInfo.Phone), config.NOT_ALLOWED))
		}
		returning_customer.CustomerWithPassword = userInfo
		returning_customer.UpdatedAt = &now
		returning_customer.Status = config.STATUS_ACTIVE

		updateResult, err := customersColl.UpdateOne(ctx, bson.M{"_id": returning_customer.Id}, bson.M{
			"name":       returning_customer.Name,
			"logo":       returning_customer.Logo,
			"password":   returning_customer.Password,
			"seller_id":  nil,
			"updated_at": returning_customer.UpdatedAt,
			"status":     returning_customer.Status,
			"rooms":      bson.A{},
		})
		if err != nil {
			helpers.DeleteImageFile(userInfo.Logo)
			return c.JSON(errRes("UpdateOne(existing_customer)", err, config.CANT_UPDATE))
		}
		if updateResult.MatchedCount == 0 {
			helpers.DeleteImageFile(userInfo.Logo)
			return c.JSON(errRes("ZeroMatchedCount", fmt.Errorf("Customer not found."), config.NOT_FOUND))
		}
		config.StatisticsService.Writer.NewActiveUser()
	} else {
		new_customer := models.CustomerForDb{
			CustomerWithPassword: userInfo,
			DeletedProfiles:      make([]primitive.ObjectID, 0),
			CreatedAt:            now,
			UpdatedAt:            nil,
			Status:               config.STATUS_ACTIVE,
			Rooms:                []primitive.ObjectID{},
		}
		insertResult, err := customersColl.InsertOne(ctx, new_customer)
		if err != nil {
			helpers.DeleteImageFile(userInfo.Logo)
			return c.JSON(errRes("InsertOne(customer)", err, config.CANT_INSERT))
		}
		userInfo.Id = insertResult.InsertedID.(primitive.ObjectID)
		config.StatisticsService.Writer.NewUser() // NewActiveUser() mi ya?
	}
	newCustomer := userInfo.WithoutPassword()
	access_token, err := helpers.CreateACCTForCustomer(newCustomer)
	if err != nil {
		return c.JSON(errRes("CreateACCTForCustomer()", err, ""))
	}
	refresh_token, err := helpers.CreateREFTForCustomer(newCustomer.Id)
	if err != nil {
		return c.JSON(errRes("CreateREFTForCustomer()", err, ""))
	}

	return c.JSON(models.Response[fiber.Map]{
		IsSuccess: true,
		Result: fiber.Map{
			"access_token":  access_token,
			"refresh_token": refresh_token,
			"user":          newCustomer,
		},
	})
}
func GetFavoritesByCustomer(ctx context.Context, customerId primitive.ObjectID) fiber.Map {
	// errRes := helpers.ErrorResponse("GetFavoritesByCustomer")
	m := fiber.Map{}
	fav_posts := config.MI.DB.Collection(config.FAV_POSTS)
	fav_posts_res, err := fav_posts.Find(ctx, bson.M{
		"customer_id": customerId,
	})
	type postID struct {
		PostId primitive.ObjectID `json:"post_id" bson:"post_id"`
	}
	if err == nil {
		m["posts"] = []primitive.ObjectID{}
		for fav_posts_res.Next(ctx) {
			var favPostRow postID
			err := fav_posts_res.Decode(&favPostRow)
			if err != nil {
				continue
			}
			m["posts"] = append(m["posts"].([]primitive.ObjectID), favPostRow.PostId)
		}
		if err := fav_posts_res.Err(); err != nil {
			m["posts"] = []primitive.ObjectID{}
		}
	} else {
		m["posts"] = []primitive.ObjectID{}
	}

	fav_products := config.MI.DB.Collection(config.FAV_PRODS)
	fav_products_res, err := fav_products.Find(ctx, bson.M{
		"customer_id": customerId,
	})
	type productID struct {
		ProductId primitive.ObjectID `json:"product_id" bson:"product_id"`
	}
	if err == nil {
		m["products"] = []primitive.ObjectID{}
		for fav_products_res.Next(ctx) {
			var favProductRow productID
			err := fav_products_res.Decode(&favProductRow)
			if err != nil {
				continue
			}
			m["products"] = append(m["products"].([]primitive.ObjectID), favProductRow.ProductId)
		}
		if err := fav_products_res.Err(); err != nil {
			m["products"] = []primitive.ObjectID{}
		}
	} else {
		m["products"] = []primitive.ObjectID{}
	}

	fav_sellers := config.MI.DB.Collection(config.FAV_SELLERS)
	fav_sellers_res, err := fav_sellers.Find(ctx, bson.M{
		"customer_id": customerId,
	})
	type sellerID struct {
		SellerId primitive.ObjectID `json:"seller_id" bson:"seller_id"`
	}
	if err == nil {
		m["sellers"] = []primitive.ObjectID{}
		for fav_sellers_res.Next(ctx) {
			var favSellerRow sellerID
			err := fav_sellers_res.Decode(&favSellerRow)
			if err != nil {
				continue
			}
			m["sellers"] = append(m["sellers"].([]primitive.ObjectID), favSellerRow.SellerId)
		}
		if err := fav_sellers_res.Err(); err != nil {
			m["sellers"] = []primitive.ObjectID{}
		}
	} else {
		m["sellers"] = []primitive.ObjectID{}
	}

	return m
}
func GetMe(c *fiber.Ctx) error {
	return c.JSON(models.Response[any]{
		IsSuccess: true,
		Result:    c.Locals("currentUser"),
	})
}

func GetMyIp(c *fiber.Ctx) error {
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    c.IP(),
	})
}
func RefreshAccessToken(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("RefreshAccessToken")
	token, err := helpers.GetTokenFromHeader(c)
	if err != nil {
		return c.Status(401).JSON(errRes("GetTokenFromHeader()", err, config.REFT_NOT_FOUND))
	}
	sub, err := helpers.ValidateToken(token, os.Getenv(config.REFT_PUBLIC_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("ValidateToken(refresh_token)", err, config.REFT_EXPIRED))
	}
	customerId, err := primitive.ObjectIDFromHex(sub.(string))
	if err != nil {
		return c.Status(401).JSON(errRes("ObjectIDFromHex()", err, config.CANT_DECODE))
	}
	customers := config.MI.DB.Collection(config.CUSTOMERS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user := customers.FindOne(ctx, bson.M{
		"_id": customerId,
	})
	if err = user.Err(); err != nil {
		return c.Status(401).JSON(errRes("FindOne()", err, config.NOT_FOUND))
	}
	var customer models.Customer
	err = user.Decode(&customer)
	if err != nil {
		return c.Status(401).JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	ttl, err := time.ParseDuration(os.Getenv(config.ACCT_EXPIREDIN))
	if err != nil {
		return c.Status(401).JSON(errRes("ParseDuration(acctexpiredin)", err, ""))
	}
	access_token, err := helpers.CreateToken(ttl, customer, os.Getenv(config.ACCT_PRIVATE_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("CreateToken(access_token)", err, ""))
	}
	ttl, err = time.ParseDuration(os.Getenv(config.REFT_EXPIREDIN))
	if err != nil {
		return c.Status(401).JSON(errRes("ParseDuration(reftexpiredin)", err, ""))
	}
	refresh_token, err := helpers.CreateToken(ttl, customer.Id, os.Getenv(config.REFT_PRIVATE_KEY))
	if err != nil {
		return c.Status(401).JSON(errRes("CreateToken(refresh_token)", err, ""))
	}

	return c.JSON(models.Response[fiber.Map]{
		IsSuccess: true,
		Result: fiber.Map{
			"access_token":  access_token,
			"refresh_token": refresh_token,
			// "user":          customer,
		},
	})
}
func ValidatePassword(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.ValidatePassword")
	var customerObjId primitive.ObjectID
	err := helpers.GetCurrentCustomer(c, &customerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentCustomer()", err, config.AUTH_REQUIRED))
	}
	var payload struct {
		Password string `json:"password"`
	}
	err = c.BodyParser(&payload)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}

	is_secure, entropy, err := helpers.IsPasswordSecure(payload.Password, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("Password(%v)", entropy), err, config.INSERCURE_PWD))
	}
	customersColl := config.MI.DB.Collection(config.CUSTOMERS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	customerCursor, err := customersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id":    customerObjId,
				"status": config.STATUS_ACTIVE,
			},
		},
		bson.M{
			"$project": bson.M{
				"password": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}

	var customerResult struct {
		Id       primitive.ObjectID `bson:"_id"`
		Password string             `bson:"password"`
	}

	if customerCursor.Next(ctx) {
		err = customerCursor.Decode(&customerResult)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = customerCursor.Err(); err != nil {
		return c.JSON(errRes("customerCursor.Err()", err, config.DBQUERY_ERROR))
	}
	if customerResult.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("User not found."), config.NOT_FOUND))
	}
	err = helpers.ComparePassword(customerResult.Password, payload.Password)
	if err != nil {
		return c.JSON(errRes("ComparePassword()", err, config.CREDENTIALS_ERROR))
	}
	return c.JSON(models.Response[bool]{
		IsSuccess: true,
		Result:    true,
	})
}
func RecoverPassword(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.RecoverPassword")
	var payload struct {
		Phone    string `json:"phone" bson:"phone"`
		Password string `json:"password" bson:"password"`
	}
	var fromDb struct {
		Id       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
		Phone    string             `json:"phone" bson:"phone"`
		Password string             `json:"password" bson:"password"`
	}
	err := c.BodyParser(&payload)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	is_phone_valid := validPhone(payload.Phone)
	if is_phone_valid == false {
		return c.JSON(errRes("InvalidPhone", errors.New("Phone number not valid."), config.STATUS_ABORTED))
	}
	is_secure, entropy, err := helpers.IsPasswordSecure(payload.Password, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("InsecurePassword(%v)", entropy), err, config.INSERCURE_PWD))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	customersColl := config.MI.DB.Collection(config.CUSTOMERS)
	cursor, err := customersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"phone":  payload.Phone,
				"status": config.STATUS_ACTIVE,
			},
		},
		bson.M{
			"$project": bson.M{
				"phone":    1,
				"password": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	if cursor.Next(ctx) {
		err = cursor.Decode(&fromDb)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if fromDb.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Phone not found."), config.NOT_FOUND))
	}
	new_hashed_password := helpers.HashPassword(payload.Password)
	_, err = customersColl.UpdateOne(ctx, bson.M{"_id": fromDb.Id}, bson.M{
		"$set": bson.M{
			"password":   new_hashed_password,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
	}
	return c.JSON(models.Response[bool]{
		IsSuccess: true,
		Result:    true,
	})
}
func HasPhone(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.HasPhone")
	var payload struct {
		Phone string `json:"phone" bson:"phone"`
	}
	var fromDb struct {
		Id    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
		Phone string             `json:"phone" bson:"phone"`
	}
	err := c.BodyParser(&payload)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}
	is_phone_valid := validPhone(payload.Phone)
	if is_phone_valid == false {
		return c.JSON(errRes("InvalidPhone", errors.New("Phone number not valid."), config.STATUS_ABORTED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	customersColl := config.MI.DB.Collection(config.CUSTOMERS)
	cursor, err := customersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"phone":  payload.Phone,
				"status": config.STATUS_ACTIVE,
			},
		},
		bson.M{
			"$project": bson.M{
				"phone": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	if cursor.Next(ctx) {
		err = cursor.Decode(&fromDb)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if fromDb.Id == primitive.NilObjectID || fromDb.Phone != payload.Phone {
		return c.JSON(errRes("NilObjectID", errors.New("Phone not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[bool]{
		IsSuccess: true,
		Result:    true,
	})
}
func ChangePassword(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.ChangePassword")
	var customerObjId primitive.ObjectID
	err := helpers.GetCurrentCustomer(c, &customerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentCustomer()", err, config.AUTH_REQUIRED))
	}
	var passwords struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	err = c.BodyParser(&passwords)
	if err != nil {
		return c.JSON(errRes("BodyParser()", err, config.CANT_DECODE))
	}

	is_secure, entropy, err := helpers.IsPasswordSecure(passwords.OldPassword, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("OldPassword(%v)", entropy), err, config.INSERCURE_PWD))
	}
	is_secure, entropy, err = helpers.IsPasswordSecure(passwords.NewPassword, config.MIN_PWD_ENTROPY)
	if is_secure == false {
		return c.JSON(errRes(fmt.Sprintf("NewPassword(%v)", entropy), err, config.INSERCURE_PWD))
	}

	customersColl := config.MI.DB.Collection(config.CUSTOMERS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	customerCursor, err := customersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": customerObjId,
			},
		},
		bson.M{
			"$project": bson.M{
				"password": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}

	var customerResult struct {
		Id       primitive.ObjectID `bson:"_id"`
		Password string             `bson:"password"`
	}

	if customerCursor.Next(ctx) {
		err = customerCursor.Decode(&customerResult)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = customerCursor.Err(); err != nil {
		return c.JSON(errRes("customerCursor.Err()", err, config.DBQUERY_ERROR))
	}
	if customerResult.Id == primitive.NilObjectID || customerResult.Password == "" {
		return c.JSON(errRes("NilObjectID", errors.New("User not found."), config.NOT_FOUND))
	}
	hashed_new := helpers.HashPassword(passwords.NewPassword)
	err = helpers.ComparePassword(customerResult.Password, passwords.OldPassword)
	if err != nil {
		return c.JSON(errRes("ComparePassword()", err, config.CREDENTIALS_ERROR))
	}

	result, err := customersColl.UpdateOne(ctx, bson.M{
		"_id": customerObjId,
	}, bson.M{
		"$set": bson.M{
			"password": hashed_new,
		},
	})
	if err != nil {
		return c.JSON(errRes("UpdateOne()", err, config.CANT_UPDATE))
	}
	if result.ModifiedCount != 1 {
		return c.JSON(errRes("ZeroMatchedCount", errors.New("No documents matched."), config.NOT_FOUND))
	}

	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.UPDATED,
	})
}
