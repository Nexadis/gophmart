package user

import (
	"crypto/sha256"
	"encoding/hex"
)

type User struct {
	Login    *string `json:"login"`
	Password *string `json:"password"`
}

func (u User) HashPassword() string {
	sum := sha256.Sum256([]byte(*u.Password))
	return hex.EncodeToString(sum[:])
}
