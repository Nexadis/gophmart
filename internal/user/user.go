package user

import (
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u User) HashPassword() (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash), nil
}

func (u User) IsValidPassword(hash string) bool {
	binHash, err := hex.DecodeString(hash)
	if err != nil {
		panic(err)
	}
	err = bcrypt.CompareHashAndPassword(binHash, []byte(u.Password))
	return err == nil
}
