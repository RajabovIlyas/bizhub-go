package v1

import (
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
	"golang.org/x/net/context"
)

func EditBanners(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.EditBanners")
	fieldNames := c.FormValue("fields")      // "bannerId.tm,bannerId.ru,...."
	fields := strings.Split(fieldNames, ",") // ["bannerId.tm", "bannerId.ru",...]
	if len(fields) == 0 {
		return c.JSON(errRes("NoBannerImages", errors.New("No banner images provided."), config.BODY_NOT_PROVIDED))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	transaction_manager := ojoTr.NewTransaction(&ctx, config.MI.DB, 3)
	tr_bannersColl := transaction_manager.Collection(config.BANNERS)
	imageFiles := []string{}
	for i := 0; i < len(fields); i++ {
		field := strings.TrimSpace(fields[i])
		imagePath, err := helpers.SaveImageFile(c, field, config.FOLDER_BANNERS)
		if err == nil {
			imageFiles = append(imageFiles, imagePath)
			bannerAndLang := strings.Split(field, ".")
			bannerObjId, err := primitive.ObjectIDFromHex(bannerAndLang[0])
			if err != nil {
				trErr := transaction_manager.Rollback()
				if trErr != nil {
					err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
				}
				helpers.DeleteImages(imageFiles)
				return c.JSON(errRes("ObjectIDFromHex(bannerID)", err, config.CANT_DECODE))
			}
			imageLang := fmt.Sprintf("image.%v", bannerAndLang[1])
			update_model := ojoTr.NewModel().SetFilter(bson.M{"_id": bannerObjId}).
				SetUpdate(bson.M{
					"$set": bson.M{
						imageLang: imagePath,
					},
				}).
				SetRollbackUpdate(bson.M{
					"$set": bson.M{
						imageLang: os.Getenv(config.DEFAULT_BANNER_IMAGE),
					},
				})
			_, err = tr_bannersColl.FindOneAndUpdate(update_model)
			if err != nil {
				trErr := transaction_manager.Rollback()
				if trErr != nil {
					err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
				}
				helpers.DeleteImages(imageFiles)
				return c.JSON(errRes("FindOneAndUpdate(banner)", err, config.CANT_UPDATE))
			}
		} else {
			trErr := transaction_manager.Rollback()
			if trErr != nil {
				err = fmt.Errorf("Source: %v - Rollback: %v", err.Error(), trErr.Error())
			}
			helpers.DeleteImages(imageFiles)
			return c.JSON(errRes("SaveImageFile(bannerImage)", err, config.CANT_DECODE))
		}
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.STATUS_COMPLETED,
	})
}
func EditBanner(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.EditBanner")
	var bannerObjId primitive.ObjectID
	bannerObjId, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.JSON(errRes("Params(bannerID)", err, config.PARAM_NOT_PROVIDED))
	}
	imagePath, err := helpers.SaveImageFile(c, "image", config.FOLDER_BANNERS)
	if err != nil {
		return c.JSON(errRes("SaveImageFile()", err, config.BODY_NOT_PROVIDED))
	}
	culture := helpers.GetCultureFromQuery(c)
	imageField := fmt.Sprintf("image.%v", culture.Lang)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	bannersColl := config.MI.DB.Collection(config.BANNERS)
	updateResult, err := bannersColl.UpdateOne(ctx, bson.M{"_id": bannerObjId}, bson.M{
		"$set": bson.M{
			imageField: imagePath,
		},
	})
	if err != nil {
		helpers.DeleteImageFile(imagePath)
		return c.JSON(errRes("UpdateOne(banner)", err, config.CANT_UPDATE))
	}
	if updateResult.MatchedCount == 0 {
		helpers.DeleteImageFile(imagePath)
		return c.JSON(errRes("ZeroMatchedCount", errors.New("Banner not found."), config.NOT_FOUND))
	}
	return c.JSON(models.Response[string]{
		IsSuccess: true,
		Result:    config.STATUS_COMPLETED,
	})
}
func GetBanners(c *fiber.Ctx) error {
	errRes := helpers.ErrorResponse("Admin.GetBanners")

	pageIndex, err := strconv.Atoi(c.Query("page", "0"))
	if err != nil {
		return c.JSON(errRes("Query(page)", err, config.QUERY_NOT_PROVIDED))
	}
	limit, err := strconv.Atoi(c.Query("limit", "1"))
	if err != nil {
		return c.JSON(errRes("Query(limit)", err, config.QUERY_NOT_PROVIDED))
	}

	bannersColl := config.MI.DB.Collection("banners")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var bannersResult []models.BannerForAdmin

	bannersCursor, err := bannersColl.Aggregate(ctx, bson.A{
		bson.M{
			"$skip": pageIndex * limit,
		},
		bson.M{
			"$limit": limit,
		},
	})
	if err != nil {
		return c.JSON(errRes("Aggregate()", err, config.DBQUERY_ERROR))
	}

	for bannersCursor.Next(ctx) {
		var banner models.BannerForAdmin
		err := bannersCursor.Decode(&banner)
		if err != nil {
			return c.JSON(errRes("Decode()", err, config.CANT_DECODE))
		}
		bannersResult = append(bannersResult, banner)
	}

	if bannersResult == nil {
		bannersResult = make([]models.BannerForAdmin, 0)
	}

	return c.JSON(models.Response[[]models.BannerForAdmin]{
		IsSuccess: true,
		Result:    bannersResult,
	})
}
