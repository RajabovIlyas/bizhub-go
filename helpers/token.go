package helpers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/devzatruk/bizhubBackend/config"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func CreateToken(ttl time.Duration, payload interface{}, privateKey string) (string, error) {
	decodedPrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", fmt.Errorf("\nError decoding key: %w\n", err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(decodedPrivateKey)
	if err != nil {
		return "", fmt.Errorf("\nError parsing key: %w\n", err)
	}
	now := time.Now()
	claims := make(jwt.MapClaims)
	claims["sub"] = payload
	claims["exp"] = now.Add(ttl).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		return "", fmt.Errorf("\nError creating token: %w\n", err)
	}
	return token, nil
}

func ValidateToken(token string, publicKey string) (interface{}, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("\nError decoding key: %w\n", err)
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(decodedPublicKey)
	if err != nil {
		return nil, fmt.Errorf("\nError parsing key: %w\n", err)
	}
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("\nUnexpected method: %s\n", t.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		// return nil, fmt.Errorf("\nValidate: %w\n", err)
		return nil, err
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("\nValidate: invalid token\n")
	}

	fmt.Printf("\ntoken claims: %v\n", claims)
	return claims["sub"], nil
}

func GetTokenFromHeader(c *fiber.Ctx) (string, error) {
	// fmt.Println("\nrequest headers =>", c.GetReqHeaders())
	authorizationHeader, ok := c.GetReqHeaders()["Authorization"]
	if ok != true {
		return "", errors.New("\nAuthorization header not found.")
	}

	fields := strings.Split(authorizationHeader, " ")

	if fields[0] != "Bearer" || len(fields) != 2 {
		return "", errors.New("\nAuthorization header invalid.")
	}

	token := fields[1]

	return token, nil

}

func CreateACCTForCustomer(anyUser interface{}) (string, error) {
	ttl, err := time.ParseDuration(os.Getenv(config.ACCT_EXPIREDIN))
	if err != nil {
		return "", fmt.Errorf("ParseDuration(acctexpiredin)- %v - %v", err.Error(), config.ACCT_TTL_NOT_VALID)
	}
	access_token, err := CreateToken(ttl, anyUser, os.Getenv(config.ACCT_PRIVATE_KEY))
	if err != nil {
		return "", fmt.Errorf("CreateToken(access_token) - %v - %v", err.Error(), config.ACCT_GENERATION_ERROR)
	}
	return access_token, nil
}
func CreateREFTForCustomer(anyUser interface{}) (string, error) {
	ttl, err := time.ParseDuration(os.Getenv(config.REFT_EXPIREDIN))
	if err != nil {
		return "", fmt.Errorf("ParseDuration(reftexpiredin)- %v - %v", err.Error(), config.REFT_TTL_NOT_VALID)
	}
	refresh_token, err := CreateToken(ttl, anyUser, os.Getenv(config.REFT_PRIVATE_KEY))
	if err != nil {
		return "", fmt.Errorf("CreateToken(refresh_token) - %v - %v", err.Error(), config.REFT_GENERATION_ERROR)
	}
	return refresh_token, nil
}
