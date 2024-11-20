package seeders

import (
	"context"
	"fmt"
	"strings"

	"github.com/devzatruk/bizhubBackend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ATTRIBUTE TYPES
const (
	STRING = "string"
	NUMBER = "number"
	COLOR  = "color"
)

type Attribute struct {
	Id         primitive.ObjectID `bson:"_id,omitempty"`
	Name       models.Translation `bson:"name"`
	Type       string             `bson:"type"`
	UnitsArray []string           `bson:"units_array"`
}

func (a Attribute) String() string {
	return fmt.Sprintf("ID: %v - TM: %v - TR: %v - EN: %v - RU: %v - TYPE: %v - UNITS: %v", a.Id, a.Name.Tm, a.Name.Tr, a.Name.En, a.Name.Ru, a.Type, a.UnitsArray)
}

var sample_attributes = []Attribute{
	{Id: primitive.NewObjectID(), Name: models.Translation{Tm: "Ram-TM", Tr: "Ram-TR", En: "Ram-EN", Ru: "Ram-RU"}, Type: STRING, UnitsArray: []string{"KB", "MB", "GB", "TB", "PB"}},
	{Id: primitive.NewObjectID(), Name: models.Translation{Tm: "Yat-TM", Tr: "Hafiza-TR", En: "Storage-EN", Ru: "Yat-RU"}, Type: STRING, UnitsArray: strings.Split("2,4,8,16,32,64,128,256", ",")},
	{Id: primitive.NewObjectID(), Name: models.Translation{Tm: "Renk-TM", Tr: "Renk-TR", En: "Color-EN", Ru: "Svet-RU"}, Type: COLOR, UnitsArray: strings.Split("FFE15D,DC3535,B01E68,BA94D1,A0E4CB,0D4C92,B6E2A1,EB6440,497174,7743DB,FF8FB1,474E68,F0FF42,82CD47,7DE5ED", ",")},
	{Id: primitive.NewObjectID(), Name: models.Translation{Tm: "Sim-TM", Tr: "Sim-TR", En: "Sim-EN", Ru: "Sim-RU"}, Type: NUMBER, UnitsArray: strings.Split("0,1,2", ",")},
	{Id: primitive.NewObjectID(), Name: models.Translation{Tm: "Esik Materialy-TM", Tr: "Elbise Malzeme-TR", En: "Cloth Material-EN", Ru: "Esik Material-RU"}, Type: STRING, UnitsArray: strings.Split("yun,yupek,pagta,polyester,deri", ",")},
	{Id: primitive.NewObjectID(), Name: models.Translation{Tm: "Material-TM", Tr: "Malzeme-TR", En: "Material-EN", Ru: "Material-RU"}, Type: STRING, UnitsArray: strings.Split("metal,poslamayan metal,agac,plastmas,ayna,rezin", ",")},
	{Id: primitive.NewObjectID(), Name: models.Translation{Tm: "Un Sorty-TM", Tr: "Un Cesidi-TR", En: "Un Kind-EN", Ru: "Sort Muka-RU"}, Type: STRING, UnitsArray: strings.Split("birinji,ikinji,ucunji,wyssiy,gara un", ",")},
	{Id: primitive.NewObjectID(), Name: models.Translation{Tm: "Teker-TM", Tr: "Teker-TR", En: "Tire-EN", Ru: "Pokryska-RU"}, Type: STRING, UnitsArray: strings.Split("225x60x16,215x60x16,215x65x16,22565x16,215x65x15,225x65x15", ",")},
}

func (s *OjoSeeder) SaveAttributes() (int64, error) {
	var batch []interface{}
	for _, a := range sample_attributes {
		batch = append(batch, a)
	}
	coll := s.db.Collection("attributes")
	insertResult, err := coll.InsertMany(context.Background(), batch)
	if err != nil {
		panic("Couldn't insert attributes.")
	}
	inserted := len(insertResult.InsertedIDs)
	num_attrs := len(sample_attributes)
	if inserted < num_attrs {
		return int64(inserted), fmt.Errorf("Wanted to insert %v attributes, but %v inserted.",
			num_attrs, inserted)
	}
	s.attributes = sample_attributes
	return int64(num_attrs), nil
}
