package helpers

import (
	"math/rand"
	"time"
)

func RandomCodeGenerator() string {
	rand.Seed(time.Now().Unix())
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	generatedCode := ""
	for i := 0; i < 6; i++ {
		generatedCode = generatedCode + string(chars[rand.Intn(len(chars))])
	}
	return generatedCode
}
