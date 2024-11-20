package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	ojoTr "github.com/devzatruk/bizhubBackend/transaction_manager"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSellerProfileProducts(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetSellerProfileProducts")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	today := time.Now()
	y, m, d := today.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)

	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
			},
		},
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$project": bson.M{
				"heading": fmt.Sprintf("$heading.%v", culture.Lang),
				"image": bson.M{
					"$first": "$images",
				},
				"price":    1,
				"discount": 1,
				"is_new": bson.M{
					"$gt": bson.A{"$created_at", lastWeek},
				},
				"status": 1,
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	cursor, err := productsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var products = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err = cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		products = append(products, product)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    products,
	})
}
func AddNewProduct(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.AddNewProduct")
	var categoryObjId, sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	now := time.Now()
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(errRes("MultipartForm()", err, config.CANT_DECODE))
	}
	if len(c.FormValue("category_id")) > 0 {
		categoryObjId, err = primitive.ObjectIDFromHex(c.FormValue("category_id"))
		if err != nil {
			return c.JSON(errRes("FormValue(category_id)", err, config.BODY_NOT_PROVIDED))
		}
	} else {
		return c.JSON(errRes("MissingData", errors.New("Category ID not provided."), config.BODY_NOT_PROVIDED))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	categoriesColl := config.MI.DB.Collection(config.CATEGORIES)
	aggregate := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": categoryObjId,
				"parent": bson.M{
					"$ne": nil,
				},
				"image": bson.M{
					"$eq": nil,
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "attributes",
				"localField":   "attributes",
				"foreignField": "_id",
				"as":           "attributes",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":        culture.Stringf("$name.%v"),
							"is_number":   1,
							"units_array": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"attributes": 1,
			},
		},
	}
	cursor, err := categoriesColl.Aggregate(ctx, aggregate)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)

	var category models.CategoryAttributes
	if cursor.Next(ctx) {
		err = cursor.Decode(&category)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if category.Id != categoryObjId {
		return c.JSON(errRes("CategoryNotFound", errors.New("Category not found."), config.NOT_FOUND))
	}
	var productData = models.NewProductData{CategoryId: categoryObjId}

	if len(c.FormValue("heading")) > 0 {
		productData.Heading = c.FormValue("heading")
	} else {
		return c.JSON(errRes("MissingData", errors.New("Heading not provided."), config.BODY_NOT_PROVIDED))
	}
	if len(c.FormValue("more_details")) > 0 {
		productData.MoreDetails = c.FormValue("more_details")
	} else {
		return c.JSON(errRes("MissingData", errors.New("More details not provided."), config.BODY_NOT_PROVIDED))
	}
	if len(c.FormValue("brand_id")) > 0 {
		brandObjId, err := primitive.ObjectIDFromHex(c.FormValue("brand_id"))
		if err != nil {
			return c.JSON(errRes("FormValue(brand_id)", err, config.CANT_DECODE))
		}
		productData.BrandId = brandObjId
	} else {
		return c.JSON(errRes("MissingData", errors.New("Brand ID not provided."), config.BODY_NOT_PROVIDED))
	}
	if len(c.FormValue("price")) > 0 {
		price, err := strconv.ParseFloat(c.FormValue("price"), 64)
		if err != nil {
			return c.JSON(errRes("FormValue(price)", err, config.BODY_NOT_PROVIDED))
		}
		productData.Price = price
	} else {
		return c.JSON(errRes("MissingData", errors.New("Price not provided."), config.BODY_NOT_PROVIDED))
	}
	productData.Attributes = []models.NewProd_Attr{}
	attributes, ok := form.Value["attributes"]

	if ok && len(attributes) > 0 {
		for _, attr := range attributes {
			var at models.NewProd_Attr
			err = json.Unmarshal([]byte(attr), &at)
			if err != nil {
				return c.JSON(errRes("Unmarshal(attribute)", err, config.CANT_DECODE))
			}
			// bolmaly attribute my we index out of range dal dalmi?
			valid_attr := false
			for _, attr := range category.Attributes {
				if at.Id == attr.Id {
					if len(attr.UnitsArray) == 0 {
						valid_attr = true
						break
					}
					if at.UnitIndex < int64(len(attr.UnitsArray)) {
						valid_attr = true
						break
					} else {
						return c.JSON(errRes("IndexOutOfRange", errors.New("Unit index out of range."), config.NOT_ALLOWED))
					}
				}
			}
			if valid_attr {
				productData.Attributes = append(productData.Attributes, at)
			} else {
				return c.JSON(errRes("AttributeNotValid", errors.New("Attribute not valid."), config.NOT_ALLOWED))
			}
		}
	} else { // attribute-y yok category bolsun diyse, su yeri pozay!
		return c.JSON(errRes("MissingData", errors.New("Attributes not provided."), config.BODY_NOT_PROVIDED))
	}
	if images, ok := form.File["images"]; ok && len(images) > 0 {
		for _, imageFile := range images {

			imagePath, err := helpers.SaveFileheader(c, imageFile, config.FOLDER_PRODUCTS)
			if err != nil {
				helpers.DeleteImages(productData.Images)
				return c.JSON(errRes("SaveFileheader(product_image)", err, config.CANT_DECODE))
			} else {
				productData.Images = append(productData.Images, imagePath)
			}
		}
	} else {
		return c.JSON(errRes("MissingData", errors.New("Images not provided."), config.BODY_NOT_PROVIDED))
	}
	heading := models.Translation{}
	more_details := models.Translation{}
	switch culture.Lang {
	case "tm":
		heading.Tm = productData.Heading
		more_details.Tm = productData.MoreDetails
	case "tr":
		heading.Tr = productData.Heading
		more_details.Tr = productData.MoreDetails
	case "ru":
		heading.Ru = productData.Heading
		more_details.Ru = productData.MoreDetails
	case "en":
		heading.En = productData.Heading
		more_details.En = productData.MoreDetails
	}
	prodforDB := models.ProductDetailWithTranslation{
		Heading:      heading,
		CategoryId:   categoryObjId,
		BrandId:      productData.BrandId,
		Images:       productData.Images,
		Price:        productData.Price,
		Discount:     0,
		MoreDetails:  more_details,
		CreatedAt:    now,
		SellerId:     sellerObjId,
		Attributes:   productData.Attributes,
		Viewed:       0,
		Likes:        0,
		Status:       config.STATUS_CHECKING,
		DiscountData: nil,
	}
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_productsColl := transaction_manager.Collection(config.PRODUCTS)
	insert_model := ojoTr.NewModel().SetDocument(prodforDB)
	insertResult, err := tr_productsColl.InsertOne(insert_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		helpers.DeleteImages(prodforDB.Images)
		return c.JSON(errRes("InsertOne(product)", err, config.CANT_INSERT))
	}
	insertedID := insertResult.InsertedID.(primitive.ObjectID)
	prodforDB.Id = insertedID
	tr_sellersColl := transaction_manager.Collection(config.SELLERS)
	update_model := ojoTr.NewModel().
		SetFilter(bson.M{"_id": sellerObjId}).
		SetUpdate(bson.M{
			"$inc":      bson.M{"products_count": 1},           // on -1 eken sebabi name bolup biler?
			"$addToSet": bson.M{"categories": categoryObjId}}). // on bu category yok bolsa gosaly
		SetRollbackUpdateWithOldData(func(i interface{}) bson.M {
			oldData := i.(bson.M)
			oldCategoriesArray := oldData["categories"].(bson.A)
			return bson.M{
				"$inc": bson.M{"products_count": -1},
				"$set": bson.M{"categories": oldCategoriesArray}}
		})
	_, err = tr_sellersColl.FindOneAndUpdate(update_model)
	if err != nil { // indi ErrNoDocument bolsa, dowam edenok!
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(seller_productsCount)", err, config.CANT_DELETE))
	}
	if err = transaction_manager.Err(); err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		helpers.DeleteImages(prodforDB.Images)
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	config.CheckerTaskService.Writer.Product(insertedID, productData.Heading, sellerObjId)
	return c.JSON(models.Response[models.ProductDetailWithTranslation]{
		IsSuccess: true,
		Result:    prodforDB,
	})
}
func EditProduct(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.EditProduct")
	productObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var sellerObjId primitive.ObjectID
	err = helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	productcColl := config.MI.DB.Collection(config.PRODUCTS)
	var oldData models.ProductDetailWithTranslation
	findResult := productcColl.FindOne(ctx, bson.M{"_id": productObjId, "seller_id": sellerObjId})
	if err = findResult.Err(); err != nil {
		return c.JSON(errRes("FindOne(product)", err, config.NOT_FOUND))
	}
	err = findResult.Decode(&oldData)
	if err != nil {
		return c.JSON(errRes("Decode(product)", err, config.CANT_DECODE))
	}
	newData := bson.M{}
	culture := helpers.GetCultureFromQuery(c)
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(errRes("MultipartForm()", err, config.CANT_DECODE))
	}

	if len(c.FormValue("new_heading")) > 0 {
		whichHeading := culture.Stringf("heading.%v")
		newData[whichHeading] = c.FormValue("new_heading")
	}
	if len(c.FormValue("new_more_details")) > 0 {
		whichDetails := culture.Stringf("more_details.%v")
		newData[whichDetails] = c.FormValue("new_more_details")
	}
	if len(c.FormValue("new_price")) > 0 {
		price, err := strconv.ParseFloat(c.FormValue("new_price"), 64)
		if err != nil {
			return c.JSON(errRes("FormValue(price)", err, config.BODY_NOT_PROVIDED))
		}
		newData["price"] = price
	}
	// db
	categoriesColl := config.MI.DB.Collection(config.CATEGORIES)
	aggregate := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": oldData.CategoryId,
				"parent": bson.M{
					"$ne": nil,
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "attributes",
				"localField":   "attributes",
				"foreignField": "_id",
				"as":           "attributes",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":        culture.Stringf("$name.%v"),
							"is_number":   1,
							"units_array": 1,
						},
					},
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"attributes": 1,
			},
		},
	}
	cursor, err := categoriesColl.Aggregate(ctx, aggregate)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)

	var category models.CategoryAttributes
	if cursor.Next(ctx) {
		err = cursor.Decode(&category)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if category.Id != oldData.CategoryId {
		return c.JSON(errRes("NotFound", errors.New("Category not found."), config.NOT_FOUND))
	}
	// end of db
	attrs := []models.NewProd_Attr{}
	attributes, ok := form.Value["new_attributes"]

	if ok && len(attributes) > 0 {
		for _, attr := range attributes {
			var at models.NewProd_Attr
			err = json.Unmarshal([]byte(attr), &at)
			if err != nil {
				return c.JSON(errRes("Unmarshal(attribute)", err, config.CANT_DECODE))
			}
			// bolmaly attribute my we index out of range dal dalmi?
			valid_attr := false
			for _, c_attr := range category.Attributes {
				if at.Id == c_attr.Id {
					if len(c_attr.UnitsArray) == 0 {
						valid_attr = true
						break
					}
					if at.UnitIndex < int64(len(c_attr.UnitsArray)) {
						valid_attr = true
						break
					} else {
						return c.JSON(errRes("IndexOutOfRange", errors.New("Unit index out of range."), config.NOT_ALLOWED))
					}
				}
			}
			if valid_attr {
				attrs = append(attrs, at)
			} else {
				return c.JSON(errRes("NotValid", errors.New("Attribute not valid."), config.NOT_ALLOWED))
			}
		}
	}
	if len(attrs) > 0 { // uytgedilen attribute var ise
		remaining_attrs := []models.NewProd_Attr{}
		remaining_attrs = append(remaining_attrs, oldData.Attributes...)

		for _, attr := range attrs {
			for i := 0; i < len(remaining_attrs); i++ {
				if attr.Id == remaining_attrs[i].Id {
					remaining_attrs[i] = attr
				}
			}
		}
		newData["attrs"] = remaining_attrs
	}
	new_images := []string{}
	if images, ok := form.File["new_images"]; ok && len(images) > 0 {
		for _, imageFile := range images {

			imagePath, err := helpers.SaveFileheader(c, imageFile, config.FOLDER_PRODUCTS)
			if err != nil {
				helpers.DeleteImages(new_images)
				return c.JSON(errRes("SaveFileheader(product_image)", err, config.CANT_DECODE))
			} else {
				new_images = append(new_images, imagePath)
			}
		}
	}
	deleted_images, del_images_ok := form.Value["deleted_images"]
	remaining_images := []string{}
	if del_images_ok && len(deleted_images) > 0 {
		for i := 0; i < len(oldData.Images); i++ {
			if !helpers.SliceContains(deleted_images, oldData.Images[i]) {
				remaining_images = append(remaining_images, oldData.Images[i])
			}
		}
	} else {
		deleted_images = []string{}
		remaining_images = append(remaining_images, oldData.Images...)
	}
	if len(new_images) > 0 {
		remaining_images = append(remaining_images, new_images...)
	}
	if len(deleted_images) > 0 || len(new_images) > 0 {
		newData["images"] = remaining_images
	}
	if len(newData) == 0 {
		return c.JSON(errRes("NoNewData", errors.New("No new data provided."), config.NOT_ALLOWED))
	} else {
		newData["status"] = config.STATUS_CHECKING
	}
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_productsColl := transaction_manager.Collection(config.PRODUCTS)
	update_model := ojoTr.NewModel().SetFilter(bson.M{"_id": productObjId}).
		SetUpdate(bson.M{
			"$set": newData,
		}).
		SetRollbackUpdate(bson.M{
			"heading":      oldData.Heading,
			"more_details": oldData.MoreDetails,
			"price":        oldData.Price,
			"attrs":        oldData.Attributes,
			"images":       oldData.Images,
			"status":       oldData.Status,
		})
	_, err = tr_productsColl.FindOneAndUpdate(update_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		helpers.DeleteImages(new_images)
		return c.JSON(errRes("FindOneAndUpdate(product)", err, config.CANT_UPDATE))
	}

	if err = transaction_manager.Err(); err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		helpers.DeleteImages(new_images)
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	if culture.Lang == "en" && len(c.FormValue("new_heading")) > 0 {
		config.CheckerTaskService.Writer.Product(productObjId, c.FormValue("new_heading"), oldData.SellerId)
	} else {
		config.CheckerTaskService.Writer.Product(productObjId, oldData.Heading.En, oldData.SellerId)
	}
	helpers.DeleteImages(deleted_images)
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.UPDATED,
	})
}

// TODO: yatdan cykan function! yazmaly
func DeletePost(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.DeletePost")
	postObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var sellerObjId primitive.ObjectID
	err = helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_postsColl := transaction_manager.Collection(config.POSTS)
	delete_model := ojoTr.NewModel().SetFilter(bson.M{"_id": postObjId, "seller_id": sellerObjId})
	deleteResult, err := tr_postsColl.FindOneAndDelete(delete_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndDelete(post)", err, config.CANT_DELETE))
	}
	var oldData struct {
		Image string `bson:"image"`
	}
	err = deleteResult.Decode(&oldData)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}

	// del(product) & delFromTasks() & delFromCronJobs() & delFromFavPosts() & updatePostsCountFromSellers()
	tr_tasksColl := transaction_manager.Collection(config.TASKS)
	delete_model = ojoTr.NewModel().SetFilter(bson.M{"target_id": postObjId})
	deleteResult, err = tr_tasksColl.FindOneAndDelete(delete_model)
	if err != nil && err != mongo.ErrNoDocuments {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndDelete(task)", err, config.CANT_DELETE))
	}

	var oldTask = models.Task{}
	if err == nil {
		err = deleteResult.Decode(&oldTask)
	}
	err = config.OjoCronService.RemoveJobsByGroup(postObjId)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("RemoveJobsByGroup(productId)", err, config.CANT_DELETE))
	}
	tr_favPostsColl := transaction_manager.Collection(config.FAV_POSTS)
	delete_model = ojoTr.NewModel().SetFilter(bson.M{"post_id": postObjId})
	_, err = tr_favPostsColl.FindOneAndDelete(delete_model)
	if err != nil && err != mongo.ErrNoDocuments {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndDelete(favorite_post)", err, config.CANT_DELETE))
	}
	tr_sellersColl := transaction_manager.Collection(config.SELLERS)
	update_model := ojoTr.NewModel().SetFilter(bson.M{"_id": sellerObjId}).
		SetUpdate(bson.M{"$inc": bson.M{"posts_count": -1}}).
		SetRollbackUpdate(bson.M{"$inc": bson.M{"posts_count": 1}})

	_, err = tr_sellersColl.FindOneAndUpdate(update_model)
	if err != nil && err != mongo.ErrNoDocuments {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(seller_postsCount)", err, config.CANT_UPDATE))
	}

	if err = transaction_manager.Err(); err != nil && err != mongo.ErrNoDocuments { // it is reachable when any operation in a transaction throws errors but transaction continues till this point, the saved error causes rollback.
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	helpers.DeleteImageFile(oldData.Image)
	if oldTask.Id != primitive.NilObjectID {
		config.CheckerTaskService.RemoveTask(oldTask.Id)
	}

	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.DELETED,
	})
}
func DeleteProduct(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.DeleteProduct")
	productObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var sellerObjId primitive.ObjectID
	err = helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_productsColl := transaction_manager.Collection(config.PRODUCTS)
	delete_model := ojoTr.NewModel().SetFilter(bson.M{"_id": productObjId, "seller_id": sellerObjId})
	deleteResult, err := tr_productsColl.FindOneAndDelete(delete_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndDelete(product)", err, config.CANT_DELETE))
	}
	var oldData models.ProductDetailWithTranslation
	err = deleteResult.Decode(&oldData)
	if err != nil {
		return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
	}
	// del(product) & delFromTasks() & delFromCronJobs() & delFromFavProducts() & updateProductsCountFromSellers()
	tr_tasksColl := transaction_manager.Collection(config.TASKS)
	delete_model = ojoTr.NewModel().SetFilter(bson.M{"target_id": productObjId})
	deleteResult, err = tr_tasksColl.FindOneAndDelete(delete_model)
	if err != nil && err != mongo.ErrNoDocuments {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndDelete(task)", err, config.CANT_DELETE))
	}
	if err == nil {
		var oldTask models.Task
		err = deleteResult.Decode(&oldTask)
		if err == nil {
			config.CheckerTaskService.RemoveTask(oldTask.Id)
		}
	}
	err = config.OjoCronService.RemoveJobsByGroup(productObjId)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("RemoveJobsByGroup(productId)", err, config.CANT_DELETE))
	}
	tr_favProdsColl := transaction_manager.Collection(config.FAV_PRODS)
	delete_model = ojoTr.NewModel().SetFilter(bson.M{"product_id": productObjId})
	_, err = tr_favProdsColl.FindOneAndDelete(delete_model)
	if err != nil && err != mongo.ErrNoDocuments {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndDelete(favorite_product)", err, config.CANT_DELETE))
	}
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	cursor, err := productsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id":   sellerObjId,
				"category_id": oldData.CategoryId,
			},
		}, bson.M{
			"$count": "products_in_category",
		},
	})
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("Aggregate(products_in_category)", err, config.CANT_DECODE))
	}
	defer cursor.Close(ctx)
	var count struct {
		ProductsCountInCategory int64 `bson:"products_in_category"`
	}
	if cursor.Next(ctx) {
		err = cursor.Decode(&count)
		if err != nil {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			return c.JSON(errRes("Decode(products_in_category)", err, config.CANT_DECODE))
		}
	}
	tr_sellersColl := transaction_manager.Collection(config.SELLERS)
	update_model := ojoTr.NewModel().SetFilter(bson.M{"_id": oldData.SellerId}).
		SetUpdate(bson.M{"$inc": bson.M{"products_count": -1}}).
		SetRollbackUpdate(bson.M{"$inc": bson.M{"products_count": 1}})
	if count.ProductsCountInCategory == 0 {
		update_model = ojoTr.NewModel().SetFilter(bson.M{"_id": oldData.SellerId}).
			SetUpdate(bson.M{
				"$inc":  bson.M{"products_count": -1},
				"$pull": bson.M{"categories": oldData.CategoryId}}).
			SetRollbackUpdate(bson.M{
				"$inc":  bson.M{"products_count": 1},
				"$push": bson.M{"categories": oldData.CategoryId}})
	}

	_, err = tr_sellersColl.FindOneAndUpdate(update_model)
	if err != nil && err != mongo.ErrNoDocuments {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("FindOneAndUpdate(seller_productsCount)", err, config.CANT_DELETE))
	}

	if err = transaction_manager.Err(); err != nil && err != mongo.ErrNoDocuments { // it is reachable when any operation in a transaction throws errors but transaction continues till this point, the saved error causes rollback.
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	helpers.DeleteImages(oldData.Images)
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.DELETED,
	})

}
func GetProductForEditing(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetPostForEditing")
	productObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	var sellerObjId primitive.ObjectID
	err = helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	aggregate := bson.A{
		bson.M{
			"$match": bson.M{
				"_id":       productObjId,
				"seller_id": sellerObjId,
				"status": bson.M{
					"$ne": config.STATUS_CHECKING,
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "attributes",
				"localField":   "attrs.attr_id",
				"foreignField": "_id",
				"as":           "attrs_",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name":        fmt.Sprintf("$name.%v", culture.Lang),
							"units_array": 1,
							"is_number":   1,
						},
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"attrs": bson.M{
					"$map": bson.M{
						"input": "$attrs",
						"as":    "attr",
						"in": bson.M{
							"$mergeObjects": bson.A{
								"$$attr",
								bson.M{
									"attr_detail": bson.M{
										"$first": bson.M{
											"$filter": bson.M{
												"input": "$attrs_",
												"cond": bson.M{
													"$eq": bson.A{"$$this._id", "$$attr.attr_id"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "categories",
				"localField":   "category_id",
				"foreignField": "_id",
				"as":           "category",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "categories",
							"localField":   "parent",
							"foreignField": "_id",
							"as":           "parent",
						},
					},
					bson.M{
						"$unwind": bson.M{
							"path": "$parent",
						},
					},
					bson.M{
						"$project": bson.M{
							"name": fmt.Sprintf("$name.%v", culture.Lang),
							"parent": bson.M{
								"_id":  "$parent._id",
								"name": fmt.Sprintf("$parent.name.%v", culture.Lang),
							},
						},
					},
				},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "brands",
				"localField":   "brand_id",
				"foreignField": "_id",
				"as":           "brand",
				"pipeline": bson.A{
					bson.M{
						"$lookup": bson.M{
							"from":         "brands",
							"localField":   "parent",
							"foreignField": "_id",
							"as":           "parent",
						},
					},
					bson.M{
						"$unwind": bson.M{
							"preserveNullAndEmptyArrays": true,
							"path":                       "$parent",
						},
					},
					bson.M{
						"$project": bson.M{
							"name": 1,
							"parent": bson.M{
								"_id":  "$parent._id",
								"name": 1,
							},
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$category",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$brand",
			},
		},
		bson.M{
			"$project": bson.M{
				"heading":      fmt.Sprintf("$heading.%v", culture.Lang),
				"more_details": fmt.Sprintf("$more_details.%v", culture.Lang),
				"price":        1,
				"images":       1,
				"brand":        1,
				"category":     1,
				"attrs":        1,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	cursor, err := productsColl.Aggregate(ctx, aggregate)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var productDetails models.ProductDetailForEditing
	if cursor.Next(ctx) {
		err = cursor.Decode(&productDetails)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if productDetails.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Product not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[models.ProductDetailForEditing]{
		IsSuccess: true,
		Result:    productDetails,
	})
}
func AddNewPost(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.AddNewPost")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	now := time.Now()
	var postData = models.NewPost{
		RelatedProducts: make([]primitive.ObjectID, 0),
	}
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(errRes("MultipartForm()", err, config.CANT_DECODE))
	}
	if len(c.FormValue("title")) > 0 {
		postData.Title = c.FormValue("title")
	} else {
		return c.JSON(errRes("MissingData", errors.New("Title not provided."), config.BODY_NOT_PROVIDED))
	}
	if len(c.FormValue("body")) > 0 {
		postData.Body = c.FormValue("body")
	} else {
		return c.JSON(errRes("MissingData", errors.New("Body not provided."), config.BODY_NOT_PROVIDED))
	}
	relatedProducts, ok := form.Value["related_products"]

	if ok && len(relatedProducts) > 0 {
		for _, relative := range relatedProducts {
			relativeProduct, err := primitive.ObjectIDFromHex(relative)
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(related_products)", err, config.CANT_DECODE))
			}
			postData.RelatedProducts = append(postData.RelatedProducts, relativeProduct)
		}
	}
	if images, ok := form.File["image"]; ok && len(images) > 0 {
		image := images[0]
		imagePath, err := helpers.SaveFileheader(c, image, config.FOLDER_POSTS)
		// imagePath, err := helpers.SaveImageFile(c, "image", config.FOLDER_POSTS)
		if err != nil {
			return c.JSON(errRes("SaveImageFile(image)", err, config.CANT_DECODE))
		} else {
			postData.Image = imagePath
		}
	} else {
		return c.JSON(errRes("MissingData", errors.New("Image not provided."), config.BODY_NOT_PROVIDED))
	}
	title := models.Translation{}
	body := models.Translation{}
	switch culture.Lang {
	case "tm":
		title.Tm = postData.Title
		body.Tm = postData.Body
	case "ru":
		title.Ru = postData.Title
		body.Ru = postData.Body
	case "en":
		title.En = postData.Title
		body.En = postData.Body
	case "tr":
		title.Tr = postData.Title
		body.Tr = postData.Body
	}
	newPost := models.PostUpsert{
		Id:              primitive.NewObjectID(),
		Title:           title,
		Body:            body,
		SellerId:        sellerObjId,
		Viewed:          0,
		Image:           postData.Image,
		Likes:           0,
		Status:          config.STATUS_CHECKING,
		Auto:            false,
		CreatedAt:       now,
		RelatedProducts: postData.RelatedProducts,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_postsColl := transaction_manager.Collection(config.POSTS)
	insert_model := ojoTr.NewModel().SetDocument(newPost)
	_, err = tr_postsColl.InsertOne(insert_model)
	if err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		helpers.DeleteImageFile(newPost.Image)
		return c.JSON(errRes("InsertOne(post)", err, config.CANT_INSERT))
	}
	config.CheckerTaskService.Writer.Post(newPost.Id, false, postData.Title, newPost.SellerId)
	if err = transaction_manager.Err(); err != nil {
		trErr := transaction_manager.Rollback()
		if trErr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
		}
		helpers.DeleteImageFile(newPost.Image)
		return c.JSON(errRes("Rollback()", err, config.TRANSACTION_FAILED))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.CREATED,
	})
}
func GetSellerProfilePosts(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetSellerProfilePosts")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	postsColl := config.MI.DB.Collection(config.POSTS)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
			},
		},
		bson.M{
			"$sort": bson.M{
				"created_at": -1,
			},
		},
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
		bson.M{
			"$addFields": bson.M{
				"title": fmt.Sprintf("$title.%v", culture.Lang),
				"body":  fmt.Sprintf("$body.%v", culture.Lang),
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := postsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var posts = []models.PostWithoutSeller{}
	for cursor.Next(ctx) {
		var post models.PostWithoutSeller
		err = cursor.Decode(&post)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		posts = append(posts, post)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.PostWithoutSeller]{
		IsSuccess: true,
		Result:    posts,
	})
}

// AddNewPost etmek ucin related_products saylamak ucin eken
func GetRelatedProductsForPost(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetRelatedProductsForPost()")
	culture := helpers.GetCultureFromQuery(c)
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	var sellerObjId primitive.ObjectID
	err = helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
				"status":    config.STATUS_PUBLISHED,
			},
		},
		bson.M{"$sort": bson.M{"created_at": -1}},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
		bson.M{
			"$project": bson.M{
				"heading": fmt.Sprintf("$heading.%v", culture.Lang),
				"image": bson.M{
					"$first": "$images",
				},
			},
		},
	}
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	cursor, err := productsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var products = []models.RelatedProductInfo{}
	for cursor.Next(ctx) {
		var prod models.RelatedProductInfo
		err = cursor.Decode(&prod)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		products = append(products, prod)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.RelatedProductInfo]{
		IsSuccess: true,
		Result:    products,
	})
}
func GetSellerProfileCategories(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("GetSellerProfileCategories")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	sellersColl := config.MI.DB.Collection("sellers")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": sellerObjId,
			},
		},
		bson.M{
			"$project": bson.M{
				"categories": 1,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "categories",
				"localField":   "categories",
				"foreignField": "_id",
				"as":           "categories",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$categories",
			},
		},
		bson.M{
			"$replaceRoot": bson.M{
				"newRoot": "$categories",
			},
		},
		bson.M{
			"$project": bson.M{
				"name": fmt.Sprintf("$name.%v", culture.Lang),
			},
		},
	}
	cursor, err := sellersColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var categories = []models.SellerProductCategory{}
	for cursor.Next(ctx) {
		var cat models.SellerProductCategory
		err = cursor.Decode(&cat)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		categories = append(categories, cat)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.SellerProductCategory]{
		IsSuccess: true,
		Result:    categories,
	})
}

// TODO: cron job-a owurmelimi ya seyle galsynmy?
// suny cron job-a donusturmeli! limit, pageindex, current user gerek dal!
func GetSellerProfileCategoriesCron(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("GetSellerProfileCategoriesCron")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerObjId,
			},
		},
		bson.M{
			"$project": bson.M{
				"category_id": 1,
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": "$category_id",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "categories",
				"localField":   "_id",
				"foreignField": "_id",
				"as":           "category",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$category",
			},
		},
		bson.M{
			"$project": bson.M{
				"name": fmt.Sprintf("$category.name.%v", culture.Lang),
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	cursor, err := productsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var categories = []models.SellerProductCategory{}
	for cursor.Next(ctx) {
		var cat models.SellerProductCategory
		err = cursor.Decode(&cat)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		categories = append(categories, cat)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.SellerProductCategory]{
		IsSuccess: true,
		Result:    categories,
	})
}
func UpdateCustomerProfile(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.UpdateCustomerProfile")
	var customerObjId primitive.ObjectID
	err := helpers.GetCurrentCustomer(c, &customerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentCustomer()", err, config.AUTH_REQUIRED))
	}
	var newData struct {
		Name string
		Logo string
	}
	newData.Name = c.FormValue("name")
	newData.Logo, _ = helpers.SaveImageFile(c, "logo", config.FOLDER_USERS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	customersColl := config.MI.DB.Collection(config.CUSTOMERS)
	data := bson.M{}
	if newData.Name != "" {
		data["name"] = newData.Name
	}
	if newData.Logo != "" {
		data["logo"] = newData.Logo
	}
	if len(data) == 0 {
		return c.JSON(errRes("NoData", errors.New("Data not provided."), config.NOT_ALLOWED))
	}
	updateResult := customersColl.FindOneAndUpdate(ctx, bson.M{"_id": customerObjId}, bson.M{"$set": data})
	if err = updateResult.Err(); err != nil {
		if newData.Logo != "" {
			helpers.DeleteImageFile(newData.Logo)
		}
		return c.JSON(errRes("FindOneAndUpdate(customer)", err, config.CANT_UPDATE))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    newData.Logo,
	})
}
func UpdateSellerProfile(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.UpdateSellerProfile")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	var newData struct {
		Name    string
		Logo    string
		CityId  string
		Address string
		Bio     string
	}
	newData.Name = c.FormValue("name")
	newData.CityId = c.FormValue("city_id")
	newData.Address = c.FormValue("address")
	newData.Bio = c.FormValue("bio")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellersColl := config.MI.DB.Collection(config.SELLERS)
	data := bson.M{}
	if newData.Address != "" {
		address := models.Translation{}
		switch culture.Lang {
		case "tm":
			address.Tm = newData.Address
		case "tr":
			address.Tr = newData.Address
		case "en":
			address.En = newData.Address
		case "ru":
			address.Ru = newData.Address
		}
		data["address"] = address
	}
	if newData.Name != "" {
		data["name"] = newData.Name
	}
	if newData.Bio != "" {
		bio := models.Translation{}
		switch culture.Lang {
		case "tm":
			bio.Tm = newData.Bio
		case "tr":
			bio.Tr = newData.Bio
		case "en":
			bio.En = newData.Bio
		case "ru":
			bio.Ru = newData.Bio
		}
		data["bio"] = bio
	}
	if newData.CityId != "" {
		data["city_id"], err = primitive.ObjectIDFromHex(newData.CityId)
		if err != nil {
			return c.JSON(errRes("ObjectIDFromHex(city_id)", err, config.CANT_DECODE))
		}
	}
	newData.Logo, _ = helpers.SaveImageFile(c, "logo", config.FOLDER_SELLERS)
	if newData.Logo != "" {
		data["logo"] = newData.Logo
	}

	if len(data) == 0 {
		return c.JSON(errRes("NoData", errors.New("Data not provided."), config.NOT_ALLOWED))
	}
	data["status"] = config.SELLER_STATUS_CHECKING
	updateResult := sellersColl.FindOneAndUpdate(ctx, bson.M{"_id": sellerObjId}, bson.M{"$set": data})
	if err = updateResult.Err(); err != nil {
		if newData.Logo != "" {
			helpers.DeleteImageFile(newData.Logo)
		}
		return c.JSON(errRes("FindOneAndUpdate(seller)", err, config.CANT_UPDATE))
	}
	var seller struct {
		Id  primitive.ObjectID `bson:"_id"`
		Bio models.Translation `bson:"bio"`
	}
	updateResult.Decode(&seller)
	if culture.Lang == "en" && newData.Bio != "" {
		config.CheckerTaskService.Writer.SellerProfile(sellerObjId, newData.Bio)
	} else {
		config.CheckerTaskService.Writer.SellerProfile(sellerObjId, seller.Bio.En)
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.STATUS_COMPLETED,
	})
}

type SellerCandidate struct {
	Name    string `json:"name" bson:"name"`
	Logo    string `bson:"logo"`
	CityId  string `json:"city_id" bson:"city_id"`
	Address string `json:"address" bson:"address"`
	Bio     string `json:"bio" bson:"bio"`
}

func (s SellerCandidate) HasEmptyFields() bool {
	if len(s.Address) == 0 || len(s.Bio) == 0 || len(s.Name) == 0 || len(s.CityId) == 0 {
		return true
	}
	return false
}
func BecomeSeller(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.BecomeSeller")
	if helpers.IsSeller(c) {
		return c.JSON(errRes("IsSeller()", errors.New("Only a regular customer can become a seller."), config.NOT_ALLOWED))
	}
	var customerObjId primitive.ObjectID
	err := helpers.GetCurrentCustomer(c, &customerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentCustomer()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	var becomeSeller SellerCandidate
	becomeSeller.Name = c.FormValue("name")
	becomeSeller.CityId = c.FormValue("city_id")
	becomeSeller.Address = c.FormValue("address")
	becomeSeller.Bio = c.FormValue("bio")
	if becomeSeller.HasEmptyFields() {
		return c.JSON(errRes("HasEmptyFields()", errors.New("Some data not provided."), config.BODY_NOT_PROVIDED))
	}
	cityObjId, err := primitive.ObjectIDFromHex(becomeSeller.CityId)
	if err != nil {
		return c.JSON(errRes("ObjectIDFromHex(city_id)", err, config.CANT_DECODE))
	}
	imagePath, errImageIsDefault := helpers.SaveImageFile(c, "logo", config.FOLDER_SELLERS)
	if errImageIsDefault != nil {
		imagePath = os.Getenv(config.DEFAULT_SELLER_IMAGE)
	}
	becomeSeller.Logo = imagePath
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	now := time.Now()
	newSellerObjId := primitive.NewObjectID()
	newWallet := models.SellerWallet{
		SellerId:  newSellerObjId,
		Balance:   0,
		CreatedAt: now,
		ClosedAt:  nil,
		Status:    config.STATUS_ACTIVE,
		InAuction: make([]models.InAuctionObject, 0),
	}
	tr_insertModel := ojoTr.NewModel().SetDocument(newWallet)
	tr_walletsColl := transaction_manager.Collection(config.WALLETS)
	_, err = tr_walletsColl.InsertOne(tr_insertModel)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		if errImageIsDefault == nil {
			helpers.DeleteImageFile(becomeSeller.Logo)
		}
		return c.JSON(errRes("InsertOne(wallet)", err, config.CANT_INSERT))
	}
	newPackageHistoryId := primitive.NewObjectID()
	newPackageHistory := models.SellerPackageHistoryFull{
		Id:             newPackageHistoryId,
		SellerId:       newWallet.SellerId,
		From:           now,
		To:             now.AddDate(0, 0, 7),
		AmountPaid:     0,
		Action:         config.PACKAGE_CHANGE,
		CurrentPackage: config.PACKAGE_TYPE_BASIC,
		Text:           "Basic package - 200 TMT.",
		CreatedAt:      now,
	}
	tr_insertModel = ojoTr.NewModel().SetDocument(newPackageHistory)
	tr_packHisColl := transaction_manager.Collection(config.PACKAGEHISTORY)
	_, err = tr_packHisColl.InsertOne(tr_insertModel)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		if errImageIsDefault == nil {
			helpers.DeleteImageFile(becomeSeller.Logo)
		}
		return c.JSON(errRes("InsertOne(package_history)", err, config.CANT_INSERT))
	}

	tr_updateModel := ojoTr.NewModel().
		SetFilter(bson.M{"_id": customerObjId}).
		SetUpdate(bson.M{"$set": bson.M{"seller_id": newSellerObjId}}).
		SetRollbackUpdate(bson.M{"$set": bson.M{"seller_id": nil}})
	tr_customersColl := transaction_manager.Collection(config.CUSTOMERS)
	tr_fupdateResult, err := tr_customersColl.FindOneAndUpdate(tr_updateModel)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		if errImageIsDefault == nil {
			helpers.DeleteImageFile(becomeSeller.Logo)
		}
		return c.JSON(errRes("UpdateOne(customer)", err, config.CANT_INSERT))
	}
	var fupdateResult interface{}
	tr_fupdateResult.Decode(&fupdateResult)
	newSeller := models.NewSeller{
		Id:            newSellerObjId,
		Name:          becomeSeller.Name,
		Address:       culture.ToTranslation(becomeSeller.Address),
		Logo:          becomeSeller.Logo,
		Bio:           culture.ToTranslation(becomeSeller.Bio),
		OwnerId:       customerObjId,
		CityId:        cityObjId,
		Type:          config.SELLER_TYPE_REGULAR,
		Likes:         0,
		Categories:    make([]primitive.ObjectID, 0),
		PostsCount:    0,
		ProductsCount: 0,
		Status:        config.STATUS_CHECKING, // admin checker confirm edyanca
		LastIn:        []time.Time{now},
		Transfers:     []primitive.ObjectID{},
		Package: models.SellerCurrentPackage{
			PackageHistoryId: newPackageHistory.Id,
			To:               newPackageHistory.To,
			Type:             newPackageHistory.CurrentPackage,
		},
	}
	tr_insertModel = ojoTr.NewModel().SetDocument(newSeller)
	tr_sellersColl := transaction_manager.Collection(config.SELLERS)
	_, err = tr_sellersColl.InsertOne(tr_insertModel)
	if err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		if errImageIsDefault == nil {
			helpers.DeleteImageFile(becomeSeller.Logo)
		}
		return c.JSON(errRes("InsertOne(seller)", err, config.CANT_INSERT))
	}
	if err = transaction_manager.Err(); err != nil {
		errTr := transaction_manager.Rollback()
		if errTr != nil {
			err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), errTr.Error())
		}
		if errImageIsDefault == nil {
			helpers.DeleteImageFile(becomeSeller.Logo)
		}
		return c.JSON(errRes("transaction_manager.Err()", err, config.TRANSACTION_FAILED))
	}
	config.CheckerTaskService.Writer.SellerProfile(newSellerObjId, becomeSeller.Bio)

	// TODO: response-a gosmaly zatlar:
	// taze access_token we refresh_token

	currentUser := c.Locals(config.CURRENT_USER)
	currentUserAsMap := currentUser.(map[string]any)
	currentUserAsMap["seller_id"] = newSeller.Id.Hex()

	access_token, err := helpers.CreateACCTForCustomer(currentUserAsMap)
	if err != nil {
		return c.JSON(errRes("CreateACCTForCustomer()", err, ""))
	}
	refresh_token, err := helpers.CreateREFTForCustomer(customerObjId)
	if err != nil {
		return c.JSON(errRes("CreateRFTForCustomer()", err, ""))
	}
	return c.JSON(models.Response[fiber.Map]{
		IsSuccess: true,
		Result: fiber.Map{
			"seller_id":     newSeller.Id,
			"access_token":  access_token,
			"refresh_token": refresh_token,
			"message":       "You are now a seller!",
		},
	})
}
func GetSellerProfile(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetSellerProfile")
	var sellerObjId primitive.ObjectID
	err := helpers.GetCurrentSeller(c, &sellerObjId)
	if err != nil {
		return c.JSON(errRes("GetCurrentSeller()", err, config.AUTH_REQUIRED))
	}
	culture := helpers.GetCultureFromQuery(c)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": sellerObjId,
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
			"$addFields": bson.M{
				"bio":     culture.Stringf("$bio.%v"),
				"address": culture.Stringf("$address.%v"),
			},
		},
	}
	sellers := config.MI.DB.Collection(config.SELLERS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := sellers.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var sellerInfo models.SellerInfo
	if cursor.Next(ctx) {
		err = cursor.Decode(&sellerInfo)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	} else {
		return c.JSON(errRes("cursor.Next()", errors.New("Seller not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[models.SellerInfo]{
		IsSuccess: true,
		Result:    sellerInfo,
	})
}
func GetTopSellers(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetTopSellers")
	// limit, err := strconv.Atoi(c.Query("limit", "5"))
	// if err != nil {
	// 	limit = 5
	// }
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellersColl := config.MI.DB.Collection(config.SELLERS)
	cursor, err := sellersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"type": bson.M{
					"$ne": config.SELLER_TYPE_REPORTERBEE,
				},
			},
		},
		bson.M{
			"$sort": bson.M{
				"likes": -1,
			},
		},
		bson.M{
			"$limit": 5,
		},
		bson.M{
			"$project": bson.M{
				"logo": 1,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	type TopSeller struct {
		Id   primitive.ObjectID `json:"_id" bson:"_id"`
		Logo string             `json:"logo" bson:"logo"`
	}
	var top_sellers = []TopSeller{}
	for cursor.Next(ctx) {
		var seller TopSeller
		err = cursor.Decode(&seller)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		top_sellers = append(top_sellers, seller)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]TopSeller]{
		IsSuccess: true,
		Result:    top_sellers,
	})
}
func GetAllSellers(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetAllSellers")
	culture := helpers.GetCultureFromQuery(c)
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	sellers := config.MI.DB.Collection(config.SELLERS)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"type": bson.M{
					"$ne": config.SELLER_TYPE_REPORTERBEE,
				},
				"status": config.SELLER_STATUS_PUBLISHED,
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
				"path":                       "$city",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$addFields": bson.M{
				"city": bson.M{
					"$ifNull": bson.A{"$city", nil},
				},
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
	}
	cursor, err := sellers.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var sellersResult = []models.Seller{}
	for cursor.Next(ctx) {
		var seller models.Seller
		err = cursor.Decode(&seller)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		sellersResult = append(sellersResult, seller)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Seller]{
		IsSuccess: true,
		Result:    sellersResult,
	})
}

// Get A Seller By Id
// TODO: bu nirede ulanyldy????
func GetProfileOfAnySeller(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetSellerById")
	culture := helpers.GetCultureFromQuery(c)
	objId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellers := config.MI.DB.Collection(config.SELLERS)
	var seller models.SellerInfo
	cursor, err := sellers.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id":    objId,
				"status": config.SELLER_STATUS_PUBLISHED,
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
			"$addFields": bson.M{
				"bio":     culture.Stringf("$bio.%v"),
				"address": culture.Stringf("$address.%v"),
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	if cursor.Next(ctx) {
		err = cursor.Decode(&seller)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	if seller.Id == primitive.NilObjectID {
		return c.JSON(errRes("NilObjectID", errors.New("Seller Profile not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[models.SellerInfo]{
		IsSuccess: true,
		Result:    seller,
	})
}

// Get Products By SellerId if The Seller by id exists
// TODO: eger status != "published" bolsa ,gorkezme!!!
func GetProductsBySellerId(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetProductsBySellerId")
	sellerId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	today := time.Now()
	y, m, d := today.Date()
	yesterday := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	lastWeek := yesterday.AddDate(0, 0, -6)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	productsColl := config.MI.DB.Collection(config.PRODUCTS)
	cursor, err := productsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerId,
				"status":    config.STATUS_PUBLISHED,
			},
		},
		bson.M{"$sort": bson.M{"created_at": -1}},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
		bson.M{
			"$addFields": bson.M{
				"is_new": bson.M{
					"$gt": bson.A{"$created_at", lastWeek},
				},
				"heading": culture.Stringf("$heading.%v"),
				"image": bson.M{
					"$first": "$images",
				},
			},
		},
		bson.M{
			"$project": bson.M{
				"images": 0,
			},
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var products = []models.Product{}
	for cursor.Next(ctx) {
		var product models.Product
		err := cursor.Decode(&product)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		products = append(products, product)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Product]{
		IsSuccess: true,
		Result:    products,
	})
}

// GET - api/v1/sellers/search?q={searchQuery}&page={page}&limit={limit}
func SearchSellers(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.SearchSellers")
	searchQuery := c.Query("q")
	if len(searchQuery) == 0 {
		return c.JSON(errRes("Query(q)", errors.New("Search query not found."), config.QUERY_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellers := config.MI.DB.Collection("sellers")
	aggregationArray := bson.A{
		bson.M{
			"$facet": bson.M{
				"byName": bson.A{
					bson.M{
						"$match": bson.M{
							"name": primitive.Regex{Pattern: fmt.Sprintf("(%v)", searchQuery), Options: "gi"},
						},
					},
				},
			},
		},
		bson.M{
			"$sort": bson.M{
				"type": 1,
				"name": -1,
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
				"from": "cities",
				"pipeline": bson.A{
					bson.M{
						"$match": bson.M{
							culture.Stringf("name.%v"): primitive.Regex{Pattern: fmt.Sprintf("(%v)", searchQuery), Options: "gi"},
						},
					},
					bson.M{
						"$project": bson.M{
							"_id": 1,
						},
					},
					bson.M{
						"$lookup": bson.M{
							"from":         "sellers",
							"foreignField": "city_id",
							"localField":   "_id",
							"as":           "sellers",
						},
					},
				},
				"as": "city_results",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"city_results": bson.M{
					"$map": bson.M{
						"input": "$city_results",
						"as":    "city_result",
						"in":    "$$city_result.sellers",
					},
				},
			},
		},
		bson.M{
			"$addFields": bson.M{
				"result": bson.M{
					"$concatArrays": bson.A{
						bson.M{"$cond": bson.A{
							bson.M{
								"$eq": bson.A{"$city_results", bson.A{}},
							},
							bson.A{}, bson.M{"$first": "$city_results"},
						}},
						"$byName",
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$result",
			},
		},
		bson.M{
			"$group": bson.M{
				"_id": "$result._id",
				"result": bson.M{
					"$first": "$result",
				},
			},
		},
		bson.M{
			"$replaceRoot": bson.M{
				"newRoot": "$result",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "cities",
				"localField":   "city_id",
				"foreignField": "_id",
				"pipeline": bson.A{
					bson.M{
						"$addFields": bson.M{
							"name": culture.Stringf("$name.%v"),
						},
					},
				},
				"as": "city",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$city",
				// "preserveNullAndEmptyArrays": true,
			},
		},
	}

	cursor, err := sellers.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var sellersResult = []models.Seller{}
	for cursor.Next(ctx) {
		var seller models.Seller
		err := cursor.Decode(&seller)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		sellersResult = append(sellersResult, seller)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Seller]{
		IsSuccess: true,
		Result:    sellersResult,
	})

}

func FilterSellers(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.FilterSellers")
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)
	onlyQuery := c.Query("only", "all")
	_citiesQuery := c.Query("city", "all")
	var citiesQuery = []any{}
	if _citiesQuery != "all" {
		err := json.Unmarshal([]byte(_citiesQuery), &citiesQuery)
		if err != nil {
			return c.JSON(errRes("json.Unmarshal(cities)", err, config.CANT_DECODE))
		}
		for i := 0; i < len(citiesQuery); i++ {
			objId, err := primitive.ObjectIDFromHex(citiesQuery[i].(string))
			if err != nil {
				return c.JSON(errRes("ObjectIDFromHex(cityId)", err, config.CANT_DECODE))
			}
			citiesQuery[i] = objId
		}

	}

	match := bson.M{}
	if onlyQuery == config.SELLER_TYPE_MANUFACTURER || onlyQuery == config.SELLER_TYPE_REGULAR {
		match["type"] = onlyQuery
	}
	if _citiesQuery != "all" && len(citiesQuery) == 0 {
		bsonA := bson.A{}
		bsonA = append(bsonA, citiesQuery...)
		match["city_id"] = bson.M{"$in": bsonA}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aggregationArray := bson.A{
		bson.M{
			"$match": match,
		},
		bson.M{
			"$sort": bson.M{
				"type": 1,
				"name": -1,
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
				"from":         "cities",
				"localField":   "city_id",
				"foreignField": "_id",
				"pipeline": bson.A{
					bson.M{
						"$addFields": bson.M{
							"name": fmt.Sprintf("$name.%v", culture.Lang),
						},
					},
				},
				"as": "city",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$city",
				// "preserveNullAndEmptyArrays": true, // TODO: su gerekmi?
			},
		},
	}
	sellers := config.MI.DB.Collection(config.SELLERS)
	cursor, err := sellers.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var sellersResult = []models.Seller{}
	for cursor.Next(ctx) {
		var seller models.Seller
		err := cursor.Decode(&seller)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		sellersResult = append(sellersResult, seller)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.Seller]{
		IsSuccess: true,
		Result:    sellersResult,
	})
}

func GetSellersFilterAggregations(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetSellersFilterAggregations")
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)
	cities := config.MI.DB.Collection(config.CITIES)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	citiesCursor, err := cities.Aggregate(ctx, bson.A{
		bson.M{
			"$addFields": bson.M{
				"name": fmt.Sprintf("$name.%v", culture.Lang),
			},
		},
		bson.M{"$skip": pageIndex * limit},
		bson.M{"$limit": limit},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer citiesCursor.Close(ctx)
	var citiesResult = []models.City{}
	for citiesCursor.Next(ctx) {
		var city models.City
		err = citiesCursor.Decode(&city)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		citiesResult = append(citiesResult, city)
	}
	if err = citiesCursor.Err(); err != nil {
		return c.JSON(errRes("citiesCursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[models.SellersFilterAggregations]{
		IsSuccess: true,
		Result: models.SellersFilterAggregations{
			Cities: citiesResult,
		},
	})
}

/*
func GetPostBySellerId(c *fiber.Ctx) error {
	sellerId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(models.ErrorResponse(err.Error()))
	}
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(models.ErrorResponse(err.Error()))
	}
	fmt.Printf("limit: %v", limit)
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(models.ErrorResponse(err.Error()))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	postsColl := config.MI.DB.Collection("posts")
	cursor, err := postsColl.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"seller_id": sellerId,
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
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$seller",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"title": fmt.Sprintf("$title.%v", culture.Lang),
				"body":  fmt.Sprintf("$body.%v", culture.Lang),
			},
		},
		bson.M{
			"$project": bson.M{
				"seller.bio":      0,
				"seller.address":  0,
				"seller.owner_id": 0,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "cities",
				"localField":   "seller.city_id",
				"foreignField": "_id",
				"as":           "seller.city",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$seller.city",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"seller.city.name": fmt.Sprintf("$seller.city.name.%v", culture.Lang),
			},
		},
	})
	var posts []models.Post

	for cursor.Next(ctx) {
		var post models.Post
		err = cursor.Decode(&post)
		if err != nil {
			return c.JSON(models.ErrorResponse(err.Error()))
		}
		posts = append(posts, post)
	}
	if err != nil {
		return c.JSON(models.ErrorResponse(err.Error()))
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(models.ErrorResponse(err.Error()))
	}
	if posts == nil {
		posts = make([]models.Post, 0)
	}
	return c.JSON(models.Response[[]models.Post]{
		IsSuccess: true,
		Result:    posts,
	})
}
*/
func GetPostBySellerId(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetPostBySellerId")
	culture := helpers.GetCultureFromQuery(c)
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}
	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	postsColl := config.MI.DB.Collection(config.POSTS)
	sellerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	aggregationArray :=
		bson.A{
			bson.M{
				"$match": bson.M{
					"seller_id": sellerObjId,
				},
			},
			bson.M{
				"$skip": pageIndex * limit,
			},
			bson.M{
				"$limit": limit,
			},
			bson.M{
				"$addFields": bson.M{
					"title": fmt.Sprintf("$title.%v", culture.Lang),
					"body":  fmt.Sprintf("$body.%v", culture.Lang),
				},
			},
		}

	cursor, err := postsColl.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var posts = []models.PostWithoutSeller{}
	for cursor.Next(ctx) {
		var post models.PostWithoutSeller
		err = cursor.Decode(&post)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		posts = append(posts, post)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.PostWithoutSeller]{
		IsSuccess: true,
		Result:    posts,
	})
}
func GetCategoriesBySellerId(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Mobile.GetCategoriesBySellerId")
	culture := helpers.GetCultureFromQuery(c)
	// limit, err := strconv.Atoi(c.Query("limit", "1"))
	// if err != nil {
	// 	return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	// }
	// pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	// if err != nil {
	// 	return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	// }
	sellerId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(id)", err, config.PARAM_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellers := config.MI.DB.Collection(config.SELLERS)
	aggregationArray := bson.A{
		bson.M{
			"$match": bson.M{
				"_id": sellerId,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "categories",
				"localField":   "categories",
				"foreignField": "_id",
				"as":           "categories",
				"pipeline": bson.A{
					bson.M{
						"$project": bson.M{
							"name": fmt.Sprintf("$name.%v", culture.Lang),
						},
					},
				},
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path": "$categories",
			},
		},
		bson.M{
			"$replaceRoot": bson.M{
				"newRoot": "$categories",
			},
		},
	}
	cursor, err := sellers.Aggregate(ctx, aggregationArray)
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}
	defer cursor.Close(ctx)
	var categories = []models.SellerProductCategory{}
	for cursor.Next(ctx) {
		var cat models.SellerProductCategory
		err = cursor.Decode(&cat)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		categories = append(categories, cat)
	}
	if err = cursor.Err(); err != nil {
		return c.JSON(errRes("cursor.Err()", err, config.DBQUERY_ERROR))
	}
	return c.JSON(models.Response[[]models.SellerProductCategory]{
		IsSuccess: true,
		Result:    categories,
	})
}
