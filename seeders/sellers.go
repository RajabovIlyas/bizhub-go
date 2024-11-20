package seeders

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *OjoSeeder) CreateSeller(n int) {
	if len(s.customers) == 0 { // to prevent panic
		fmt.Printf("\nNo customers in seeder.customers[] - Aborting CreateSeller()\n")
		return
	}
	if len(s.cities) == 0 {
		fmt.Printf("\nNo cities in seeder.cities[] - Aborting CreateSeller()\n")
		return
	}
	bio_text := "a global company producing quality products at honest pricing."
	seller_types := []string{config.SELLER_TYPE_MANUFACTURER, config.SELLER_TYPE_REGULAR}
	seller_logos := s.getSeedImages("sellers")
	for i := 0; i < n; i++ {
		selected_image := s.selectImage(seller_logos)
		saved_path, err := s.saveImage(selected_image, config.FOLDER_SELLERS)
		if err != nil {
			fmt.Printf("\nError from SaveImage: %v\n", err)
			saved_path = os.Getenv(config.DEFAULT_SELLER_IMAGE)
		}
		fmt.Printf("\nSaved: %v\n", saved_path)

		selected_customer := s.getAcustomer()
		selected_city := s.getAcity()
		seller_type := rand.Intn(2)
		seller_suffix := rand.Intn(1e4)
		now := time.Now()
		last_in1 := now.Add(-time.Hour * time.Duration(rand.Intn(3*24)+5))
		last_in2 := last_in1.Add(time.Minute * time.Duration(rand.Intn(2*60)+60))
		newWallet := models.SellerWallet{
			SellerId:  primitive.NewObjectID(),
			Balance:   0,
			CreatedAt: now,
			ClosedAt:  nil,
			Status:    config.STATUS_ACTIVE,
			InAuction: make([]models.InAuctionObject, 0),
		}

		newPackageHistory := models.SellerPackageHistoryFull{
			Id:             primitive.NewObjectID(),
			SellerId:       newWallet.SellerId,
			From:           now,
			To:             now.AddDate(0, 0, 7),
			AmountPaid:     0,
			Action:         config.PACKAGE_CHANGE,
			CurrentPackage: config.PACKAGE_TYPE_BASIC,
			Text:           "Basic package - 200 TMT.",
			CreatedAt:      now,
		}

		newSeller := models.NewSeller{
			Id:   newWallet.SellerId,
			Name: fmt.Sprintf("Test Seller %v", seller_suffix),
			Address: models.Translation{
				Tm: fmt.Sprintf("TM test address: %v", seller_suffix),
				Tr: fmt.Sprintf("TR test address: %v", seller_suffix),
				En: fmt.Sprintf("EN test address: %v", seller_suffix),
				Ru: fmt.Sprintf("RU test address: %v", seller_suffix),
			},
			Logo: saved_path,
			Bio: models.Translation{
				Tm: fmt.Sprintf("TM test bio %v - %v", seller_suffix, bio_text),
				Tr: fmt.Sprintf("TR test bio %v - %v", seller_suffix, bio_text),
				En: fmt.Sprintf("EN test bio %v - %v", seller_suffix, bio_text),
				Ru: fmt.Sprintf("RU test bio %v - %v", seller_suffix, bio_text),
			},
			OwnerId:       selected_customer.Id,
			CityId:        selected_city.Id,
			Type:          seller_types[seller_type],
			Likes:         0, //int64(rand.Intn(50)),
			PostsCount:    0, //int64(rand.Intn(5)),
			Categories:    []primitive.ObjectID{},
			ProductsCount: 0, //int64(rand.Intn(5)),
			Status:        config.SELLER_STATUSES[rand.Intn(len(config.SELLER_STATUSES))],
			LastIn:        []time.Time{last_in1, last_in2},
			Transfers:     make([]primitive.ObjectID, 0),
			Package: models.SellerCurrentPackage{
				PackageHistoryId: newPackageHistory.Id,
				To:               newPackageHistory.To,
				Type:             newPackageHistory.CurrentPackage,
			},
		}
		fmt.Printf("\nNew Seller: %v\n", newSeller)
	}
}
