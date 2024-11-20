package seeders

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *OjoSeeder) CreateCustomer(n int) {
	pwd := helpers.HashPassword("12345678")
	avatar_images := s.getSeedImages("customers")
	phone_prefixes := []string{"61", "62", "63", "64", "65"}
	for i := 0; i < n; i++ {

		selected_image := s.selectImage(avatar_images)
		saved_path, err := s.saveImage(selected_image, config.FOLDER_USERS)
		if err != nil {
			fmt.Printf("\nError from SaveImage: %v\n", err)
			saved_path = os.Getenv(config.DEFAULT_USER_IMAGE)
		}
		fmt.Printf("\nSaved: %v\n", saved_path)
		phone_prefix := phone_prefixes[rand.Intn(len(phone_prefixes))]
		phone_suffix := int64(rand.Float64() * 1e6) // 61 - 02 61 68
		new_customer := models.CustomerForDb{
			CustomerWithPassword: models.CustomerWithPassword{
				Id:       primitive.NewObjectID(),
				Name:     fmt.Sprintf("Test Musderi %v", rand.Intn(10000)),
				Logo:     saved_path,
				Phone:    fmt.Sprintf("%v%v", phone_prefix, phone_suffix),
				SellerId: nil,
				Password: pwd,
			},
			CreatedAt:       time.Now(),
			UpdatedAt:       nil,
			DeletedProfiles: make([]primitive.ObjectID, 0),
			Status:          config.STATUS_ACTIVE,
		}
		s.customers = append(s.customers, new_customer)
	}
}
func (s *OjoSeeder) PrintCustomers() {

	for i, customer := range s.customers {
		fmt.Printf("\n%v: %v\n", i+1, customer)
		helpers.DeleteImageFile(customer.Logo) // biderek hdd doldurmasyn!
	}
}
