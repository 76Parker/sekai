package dto

import "github.com/golang-jwt/jwt/v5"

type KeycloakClaims struct {
	jwt.RegisteredClaims
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
}
