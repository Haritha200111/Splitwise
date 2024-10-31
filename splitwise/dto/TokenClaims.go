package dto

import (
	jwt "github.com/dgrijalva/jwt-go"
)

//  "github.com/golang-jwt/jwt/v4"

type TokenClaims struct {
	UserEmail string
	TokenID   string
	Jti       string
	Key       string
	jwt.StandardClaims
}
