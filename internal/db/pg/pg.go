package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/Nexadis/gophmart/internal/user"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PG struct {
	db *sql.DB
}

func New() db.Database {
	db := &PG{
		db: &sql.DB{},
	}
	return db
}

func (pg *PG) Open(Addr string) error {
	pgx, err := sql.Open("pgx", Addr)
	if err != nil {
		return err
	}
	err = pgx.Ping()
	if err != nil {
		logger.Logger.Errorln("Can't connect to db")
		return err
	}
	pg.db = pgx
	_, err = pgx.Exec(db.Schema)
	if err != nil {
		logger.Logger.Errorln(err)
	}
	return nil
}

func (pg *PG) Close() error {
	return pg.db.Close()
}

func (pg *PG) AddUser(ctx context.Context, user *user.User) error {
	hash, err := user.HashPassword()
	if err != nil {
		return fmt.Errorf("%s: %w", db.ErrSomeWrong, err)
	}
	stmt, err := pg.db.Prepare("INSERT INTO Users(\"login\",\"hashpass\") values($1,$2)")
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
			return db.ErrUserIsExist
		}
		return fmt.Errorf("%s: %w", db.ErrSomeWrong, err)
	}
	logger.Logger.Infof("User:'%s' with hash '%s' added!", user.Login, hash)
	return nil
}
