package models

type Translation struct {
	Tm string `json:"tm" bson:"tm"`
	Ru string `json:"ru" bson:"ru"`
	En string `json:"en" bson:"en"`
	Tr string `json:"tr" bson:"tr"`
}

func (t Translation) HasEmptyFields() bool {
	if len(t.En) == 0 || len(t.Ru) == 0 || len(t.Tm) == 0 || len(t.Tr) == 0 {
		return true
	}
	return false
}
