package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var (
	ErrJwt           = errors.New("jwt token missing or invalid")
	ErrLoginNotFound = errors.New("login not found in jwt")
)

func CookieToken(value string) *http.Cookie {
	return &http.Cookie{
		Name:    echo.HeaderAuthorization,
		Value:   "Bearer " + value,
		Expires: time.Now().Add(TokenExp),
	}
}

func GetLogin(c echo.Context) (string, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "", ErrJwt
	}
	logger.Logger.Infof("got token %s", token.Claims)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("can't cast to claims")
	}
	login, ok := claims["login"].(string)
	if !ok {
		return "", ErrLoginNotFound
	}
	return login, nil
}
