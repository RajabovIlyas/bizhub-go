package seeders

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/devzatruk/bizhubBackend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var cities = []string{"Aşgabat", "Änew", "Abadan", "Tejen", "Baharly", "Gökdepe",
	"Balkanabat", "Bereket", "Gumdag", "Hazar", "Daşoguz", "Köneürgenç", "Atamyrat",
	"Gazojak", "Magdanly", "Seýdi", "Türkmenabat", "Mary", "Baýramaly", "Ýolöten",
	"Serhetabat", "Şatlyk", "Murgap"}

type CityModel struct {
	Id   primitive.ObjectID `bson:"_id"`
	Name models.Translation `bson:"name"`
}

func (c CityModel) String() string {
	return fmt.Sprintf("ID: %v - Name { Tm: %v - Tr: %v - En: %v - Ru: %v",
		c.Id, c.Name.Tm, c.Name.Tr, c.Name.En, c.Name.Ru)
}

func (s *OjoSeeder) CreateCity(n int) {
	var cityIndex int
	for i := 0; i < n; i++ {
		for {
			cityIndex = rand.Intn(len(cities))
			if _, exists := s.cities[cities[cityIndex]]; !exists {
				break
			}
		}
		cityName := cities[cityIndex]
		city := CityModel{
			Id: primitive.NewObjectID(),
			Name: models.Translation{
				Tm: fmt.Sprintf("%v - TM", cityName),
				Tr: fmt.Sprintf("%v - TR", cityName),
				En: fmt.Sprintf("%v - EN", cityName),
				Ru: fmt.Sprintf("%v - RU", cityName),
			},
		}
		s.cities[cityName] = city
	}
}
func (s *OjoSeeder) SaveCities() (int64, error) {
	if len(s.cities) == 0 {
		return 0, errors.New("No cities to save.")
	}
	ctx := context.Background()
	numberOfCities := len(s.cities)
	var batch = make([]interface{}, 0, numberOfCities)
	for _, city := range s.cities {
		batch = append(batch, city)
	}
	coll := s.db.Collection("cities") // eger yok collection bolsa, create eder!
	insertResult, err := coll.InsertMany(ctx, batch)
	if err != nil {
		panic("Couldn't insert cities into database.")
	}
	inserted := len(insertResult.InsertedIDs)
	if inserted < numberOfCities {
		return int64(inserted), fmt.Errorf("Wanted to insert %v cities, but %v inserted.",
			numberOfCities, inserted)
	}
	return int64(numberOfCities), nil
}
func (s *OjoSeeder) PrintCities() {

	for cityName, city := range s.cities {
		fmt.Printf("\nCity %v: %v\n", cityName, city)
	}
}
