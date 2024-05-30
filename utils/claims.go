package utils

import "github.com/golang-jwt/jwt"

type Claims struct {
	jwt.StandardClaims
	Role string `json:"role"`
}
