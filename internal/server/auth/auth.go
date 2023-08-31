package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenExp = time.Hour * 2
)

var SignMethod = jwt.SigningMethodHS256

type TypeMethod = jwt.SigningMethodHMAC

type Claims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

func NewToken(login string, secret []byte) (string, error) {
	token := jwt.NewWithClaims(SignMethod, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		Login: login,
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func IsValidToken(tokenString string, key []byte) bool {
	token, err := GetToken(tokenString, key)
	if err != nil {
		return false
	}
	if !token.Valid {
		return false
	}
	return true
}

func GetToken(tokenString string, key []byte) (*jwt.Token, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*TypeMethod); !ok {
				return nil, errors.New(`invalid method`)
			}
			return key, nil
		})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func GetClaims(token *jwt.Token) *Claims {
	if claims, ok := token.Claims.(*Claims); ok {
		return claims
	}
	return nil
}
