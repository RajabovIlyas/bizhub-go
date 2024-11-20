package seeders

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
_id
images
title
body
seller_id
related_products
*/
var prodImages = []string{
	"images/products/prod1.webp",
	"images/products/prod2.webp",
	"images/products/prod3.webp",
	"images/products/prod4.webp",
	"images/products/prod1.jpg",
	"images/products/prod2.jpg",
	"images/products/prod3.jpg",
	"images/products/prod4.jpg",
	"images/products/prod5.jpg",
	"images/products/prod6.jpg",
	"images/products/prod7.jpg",
	"images/products/prod8.jpg",
	"images/products/prod9.jpg",
	"images/products/prod10.jpg",
	"images/products/prod11.jpg",
	"images/products/prod12.jpg",
	"images/products/prod13.jpg",
	"images/products/prod14.jpg",
	"images/products/prod15.jpg",
	"images/products/prod16.jpg",
	"images/products/prod17.jpg",
	"images/products/prod18.jpg",
	"images/products/prod19.jpg",
	"images/products/prod20.jpg",
}
var textProd = `Ýewropada we Aziýada eSIM tehnologiýasy barha meşhurlyk gazanýar we Apple kompaniýasy täze iPhone 14 smartfonlary üçin SIM-kartlaryň slotlaryndan ýüz öwrüp biler. Bu barada «Trend» habar berýär.

Ýagny Apple öz iPhone telefonlarynyň birnäçe modifikasiýalaryny öndürer diýlip çak edilýär, olar fiziki we sanly SIM-kartlary göteriji, şeýle-de diňe eSIM ulgamynda işleýän kysymlar bolar. Ýagny önümçilikçi köpçülikleýin bazar üçin SIM-kartlary ulanmak mümkin bolan modeli hem saklap galar.

Aragatnaşyga diňe eSIM arkaly birikdirilýän smartfonyň nusgasy diňe şu ady tutulan ulgam giňden ýaýran ýurtlarda elýeterli bolar.

Ýeri gelende aýtsak, eSIM tehnologiýasy ilkinji gezek iPhone Xs we Xr kysymlarda peýda bolupdy. Indi iPhone 13 smartfonynyň çykmagy bilen kompaniýa ulanyjylara birbada eSIM ulgamynyň ikisinden peýdalanmaga mümkinçilik berdi.`

func SeedProducts(c *fiber.Ctx) error {

	count, err := strconv.Atoi(c.Query("count", "1"))
	if err != nil {
		return c.JSON(models.ErrorResponse(err))
	}
	posts := config.MI.DB.Collection("posts")
	sellers := config.MI.DB.Collection("sellers")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sellerCursor, err := sellers.Find(ctx, bson.M{})
	var sellersResult []models.Seller

	for sellerCursor.Next(ctx) {
		var seller models.Seller
		err := sellerCursor.Decode(&seller)
		if err != nil {
			return c.JSON(models.ErrorResponse(err))
		}
		sellersResult = append(sellersResult, seller)
	}
	postsCount, err := posts.CountDocuments(ctx, bson.M{})
	if err != nil {
		return c.JSON(models.ErrorResponse(err))
	}
	rand.Seed(time.Now().Unix())
	for i := 0; i < count; i++ {
		objectId := primitive.NewObjectID()
		num := postsCount + int64(i) + 1
		// initialize global pseudo random generator
		selected := rand.Intn(int(len(sellersResult)))
		selectedImage := rand.Intn(int(len(imagesList)))
		selectedTextStart := rand.Intn(int(len(textLorem)) / 5)
		selectedTextEnd := rand.Intn(int(len(textLorem)) / 5)
		textStart, textEnd := 0, 0
		if selectedTextStart > selectedTextEnd {
			textStart = selectedTextEnd
			textEnd = selectedTextStart
		}
		fmt.Printf("seller index %v", selected)
		post := models.PostUpsert{
			Id: objectId,
			Title: models.Translation{
				Tm: fmt.Sprintf("post %v title tm", num),
				Ru: fmt.Sprintf("post %v title ru", num),
				En: fmt.Sprintf("post %v title en", num),
				Tr: fmt.Sprintf("post %v title tr", num),
			},
			Body: models.Translation{
				Tm: fmt.Sprintf("post %v body tm. %v.", num, textLorem[textStart:textEnd]),
				Ru: fmt.Sprintf("post %v body ru. %v.", num, textLorem[textStart:textEnd]),
				En: fmt.Sprintf("post %v body en. %v.", num, textLorem[textStart:textEnd]),
				Tr: fmt.Sprintf("post %v body tr. %v.", num, textLorem[textStart:textEnd]),
			},
			Viewed:          0,
			RelatedProducts: []primitive.ObjectID{},
			SellerId:        sellersResult[selected].Id,
			Image:           imagesList[selectedImage],
		}
		_, err := posts.InsertOne(ctx, post)
		if err != nil {
			return c.JSON(models.ErrorResponse(err))
		}

	}
	return c.JSON(models.Response[bool]{
		IsSuccess: true,
		Result:    true,
	})
}
