package rest

import (
	"fmt"
	"split/splitwise/dto"
	"strings"

	Error "split/error"
	"split/splitwise/db"

	// "github.com/golang-jwt/jwt/v5"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GetKeyClaims(auth string) (*dto.TokenClaims, error) {
	if auth == "" {
		return nil, fmt.Errorf("authorization token is empty")
	}

	// No need to split; auth should be the token string directly.
	token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(viper.GetString("JWTSignedString")), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var userClaims dto.TokenClaims
		if err := mapstructure.Decode(claims, &userClaims); err != nil {
			return nil, fmt.Errorf("error decoding claims: %w", err)
		}

		return &userClaims, nil
	}
	return nil, fmt.Errorf("invalid token claims")
}

func Authenticate(rpcRequest *dto.AuthenticateRequest) (*dto.UserAccount, error) {
	credential := rpcRequest.Authorization

	if credential == "" || !strings.HasPrefix(credential, "Bearer ") {
		log.Println("Missing or invalid Authorization header")
		return nil, Error.ErrInvalidCredential
	}

	tokenString := strings.TrimPrefix(credential, "Bearer ")
	log.Println("tokenString", tokenString)

	// Validate and parse claims from JWT
	tokenClaims, err := GetKeyClaims(tokenString)
	if err != nil {
		log.Println("Invalid or expired token:", err)
		return nil, Error.ErrInvalidToken
	}

	log.Println("token claims:", tokenClaims)

	userName := tokenClaims.UserEmail

	existUser, err := validateUser(userName)
	if err != nil {
		return nil, err
	}

	return existUser, nil
}

func validateUser(email string) (*dto.UserAccount, error) {
	var existUser dto.UserAccount
	if err := db.DB.Find(&existUser, "emailid", email); err != nil {
		log.Println("errrrrr", err)
	}
	if existUser.EmailId == "" {
		return nil, Error.NOT_FOUND_USER
	}
	return &existUser, nil
}
