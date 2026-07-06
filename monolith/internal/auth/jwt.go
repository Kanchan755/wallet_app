package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func getSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		//only ofr local
		return []byte("super-secret-key-change")
	}
	return []byte(secret)
}

func GenerateToken(userID, email string, duration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        userID + "_" + time.Now().Format("20260102150405"),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecretKey())
}

func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (any, error) {
		return getSecretKey(), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Invalid token claims")
	}

	return claims, err
}
