package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func CookieToken(value string) *http.Cookie {
	return &http.Cookie{
		Name:    echo.HeaderAuthorization,
		Value:   "Bearer " + value,
		Expires: time.Now().Add(TokenExp),
	}
}

func GetLogin(cookie string) (string, error) {
	cookie = strings.TrimSpace(cookie)
	tokenString := strings.Split(cookie, " ")[1]
	token, err := GetToken(tokenString, []byte(""))
	if err != nil {
		return "", nil
	}
	claims := GetClaims(token)
	return claims.Login, nil
}
