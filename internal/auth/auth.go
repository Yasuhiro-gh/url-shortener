package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

const TOKEN_EXP = time.Minute * 10
const SECRET_KEY = "yasuhiro_gh"

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func BuildJWTString(newUserID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		UserID: newUserID,
	})
	return token.SignedString([]byte(SECRET_KEY))
}

func GetUserID(tokenString string) (int, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		return -1, errors.New("token parse error")
	}

	if !token.Valid {
		return -1, errors.New("token invalid")
	}

	return claims.UserID, nil
}
