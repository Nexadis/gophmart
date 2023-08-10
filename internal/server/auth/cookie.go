package auth

import (
	"net/http"
	"time"
)

func CookieToken(value string) *http.Cookie {
	return &http.Cookie{
		Name:    "Authorization",
		Value:   "Bearer " + value,
		Expires: time.Now().Add(TokenExp),
	}
}
