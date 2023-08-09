package db

import (
	"context"
	"errors"

	"github.com/Nexadis/gophmart/internal/user"
)

const Schema = `CREATE TABLE Users(
"login" VARCHAR(256) PRIMARY KEY,
"hashpass" VARCHAR(256) NOT NULL);
`

var (
	ErrUserIsExist = errors.New(`user is exist`)
	ErrSomeWrong   = errors.New(`some wrong`)
)

type Database interface {
	Open(Addr string) error
	AddUser(ctx context.Context, user *user.User) error
	Close() error
}
