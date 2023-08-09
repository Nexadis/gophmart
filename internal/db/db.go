package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/Nexadis/gophmart/internal/user"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const schema = `CREATE TABLE Users(
"login" VARCHAR(256) PRIMARY KEY,
"hashpass" VARCHAR(256) NOT NULL);
`

var (
	ErrUserIsExist = errors.New(`user is exist`)
	ErrSomeWrong   = errors.New(`some wrong`)
)

type DB struct {
	db *sql.DB
}

func New() Database {
	db := &DB{
		db: &sql.DB{},
	}
	return db
}

type Database interface {
	Open(Addr string) error
	AddUser(ctx context.Context, user *user.User) error
	Close() error
}

func (db *DB) Open(Addr string) error {
	pgx, err := sql.Open("pgx", Addr)
	if err != nil {
		return err
	}
	err = pgx.Ping()
	if err != nil {
		logger.Logger.Errorln("Can't connect to db")
		return err
	}
	db.db = pgx
	_, err = pgx.Exec(schema)
	if err != nil {
		logger.Logger.Errorln(err)
	}
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) AddUser(ctx context.Context, user *user.User) error {
	hash, err := user.HashPassword()
	if err != nil {
		return fmt.Errorf("%s: %w", ErrSomeWrong, err)
	}
	stmt, err := db.db.Prepare("INSERT INTO Users(\"login\",\"hashpass\") values($1,$2)")
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx,
		user.Login,
		hash,
	)
	if err != nil {
		logger.Logger.Error(err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrUserIsExist
		}
		return fmt.Errorf("%s: %w", ErrSomeWrong, err)
	}
	logger.Logger.Infof("User:'%s' with hash '%s' added!", user.Login, hash)
	return nil
}
