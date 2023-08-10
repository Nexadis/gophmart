package auth

import (
	"net/http"
	"time"
)

func CookieToken(value string) *http.Cookie {
	return &http.Cookie{
		Name:    "token",
		Value:   value,
		Expires: time.Now().Add(TokenExp),
	}
}
