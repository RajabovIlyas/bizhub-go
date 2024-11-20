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
var imagesList = []string{
	"images/posts/apple-2.webp",
	"images/posts/boy-1.jpg",
	"images/posts/camera1.webp",
	"images/posts/camera-2.jpg",
	"images/posts/coffee1.jpg",
	"images/posts/desk1.webp",
	"images/posts/digital-marketing1.jpg",
	"images/posts/internet-1.jpg",
	"images/posts/laptop-1.webp",
	"images/posts/man-1.jpg",
	"images/posts/nature-1.jpg",
	"images/posts/navigation1.webp",
	"images/posts/office-1.webp",
	"images/posts/office-2.webp",
	"images/posts/phone.webp",
	"images/posts/receptionists-1.jpg",
	"images/posts/samsung-1.webp",
	"images/posts/selfie-1.jpg",
	"images/posts/woman-1.webp",
	"images/posts/work-1",
}
var textLorem = `Ýewropada we Aziýada eSIM tehnologiýasy barha meşhurlyk gazanýar we Apple kompaniýasy täze iPhone 14 smartfonlary üçin SIM-kartlaryň slotlaryndan ýüz öwrüp biler. Bu barada «Trend» habar berýär.

Ýagny Apple öz iPhone telefonlarynyň birnäçe modifikasiýalaryny öndürer diýlip çak edilýär, olar fiziki we sanly SIM-kartlary göteriji, şeýle-de diňe eSIM ulgamynda işleýän kysymlar bolar. Ýagny önümçilikçi köpçülikleýin bazar üçin SIM-kartlary ulanmak mümkin bolan modeli hem saklap galar.

Aragatnaşyga diňe eSIM arkaly birikdirilýän smartfonyň nusgasy diňe şu ady tutulan ulgam giňden ýaýran ýurtlarda elýeterli bolar.

Ýeri gelende aýtsak, eSIM tehnologiýasy ilkinji gezek iPhone Xs we Xr kysymlarda peýda bolupdy. Indi iPhone 13 smartfonynyň çykmagy bilen kompaniýa ulanyjylara birbada eSIM ulgamynyň ikisinden peýdalanmaga mümkinçilik berdi.`

func SeedPosts(c *fiber.Ctx) error {

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
