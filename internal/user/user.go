package user

import (
	"encoding/hex"

	"github.com/Nexadis/gophmart/internal/order"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	HashPass string `json:"-"`
	Balance  order.Balance
}

func (u *User) HashPassword() (string, error) {
	if u.HashPass == "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		u.HashPass = hex.EncodeToString(hash)
	}
	return u.HashPass, nil
}

func (u User) IsValidHash(hash string) bool {
	binHash, err := hex.DecodeString(hash)
	if err != nil {
		panic(err)
	}
	err = bcrypt.CompareHashAndPassword(binHash, []byte(u.Password))
	return err == nil
}
