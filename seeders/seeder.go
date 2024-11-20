package seeders

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/devzatruk/bizhubBackend/helpers"
	"github.com/devzatruk/bizhubBackend/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OjoSeeder struct {
	db          *mongo.Database
	collections map[string]string
	cities      map[string]CityModel
	attributes  []Attribute
	customers   []models.CustomerForDb
	sellers     []models.NewSeller
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
func NewSeeder(database *mongo.Database) *OjoSeeder {
	ctx := context.Background()
	collections, err := database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		panic("Couldn't get collection names in the database.")
	}
	seeder := &OjoSeeder{
		db:          database,
		collections: make(map[string]string),
		cities:      make(map[string]CityModel),
		attributes:  make([]Attribute, 0),
		customers:   make([]models.CustomerForDb, 0),
		sellers:     make([]models.NewSeller, 0),
	}
	for _, c := range collections {
		seeder.collections[c] = c
	}
	return seeder
}

func (s *OjoSeeder) getSeedImages(folder string) []string {
	root := config.RootPath
	dir := filepath.Join(root, "seeders", "seed_images", folder)
	path, err := filepath.Abs(dir)
	if err != nil {
		panic("Error from Abs()...")
	}
	files, err := os.ReadDir(path)
	if err != nil {
		panic("Error from ReadDir()...")
	}
	file_names := []string{}
	for _, file := range files {
		file_path := filepath.Join(path, file.Name())
		file_names = append(file_names, file_path)
	}
	return file_names
}

func (s *OjoSeeder) selectImage(images []string) string {
	selected_index := rand.Intn(len(images))
	return images[selected_index]
}
func (seeder *OjoSeeder) saveImage(source_path string, folder string) (string, error) {
	split := strings.Split(source_path, string(os.PathSeparator))
	file_name := split[len(split)-1]
	s := strings.Split(file_name, ".")
	extension := strings.ToLower(s[len(s)-1])
	if !helpers.SliceContains(config.IMAGE_EXTENSIONS, extension) {
		return "", fmt.Errorf("File: %v - is not a valid image file.", file_name)
	}
	rootPath := config.RootPath
	newFileUuid := uuid.New()
	splitUuid := strings.Split(newFileUuid.String(), "-")
	newFileName := strings.Join(splitUuid, "") + "." + extension
	newFilePath := path.Join("images", folder, newFileName)
	absPath := path.Join(rootPath, "public", newFilePath)
	_, err := helpers.CopyFile(source_path, absPath)
	if err != nil {
		return "", err
	}
	return newFilePath, nil
}
func (s *OjoSeeder) getAcity() CityModel {
	if len(s.cities) == 0 {
		return CityModel{}
	}
	len_cities := len(s.cities)
	selected_city, i := rand.Intn(len_cities), 0
	for _, city := range s.cities {
		if i == selected_city {
			return city
		}
		i++
	}
	return CityModel{} // this shouldn't execute!
}
func (s *OjoSeeder) becameSeller(selected_customer *models.CustomerForDb) bool {
	for _, seller := range s.sellers {
		if seller.OwnerId == (*selected_customer).Id {
			return true
		}
	}
	return false
}
func (s *OjoSeeder) getAcustomer() models.CustomerForDb {
	if len(s.customers) == 0 {
		return models.CustomerForDb{}
	}
	len_customers := len(s.customers)
	for {
		selected_customer := rand.Intn(len_customers)
		if !s.becameSeller(&s.customers[selected_customer]) {
			return s.customers[selected_customer]
		}
	}
}
