package helpers

import "go.mongodb.org/mongo-driver/bson/primitive"

func IsNilObjectID(objID primitive.ObjectID) bool {
	return objID == primitive.NilObjectID
}