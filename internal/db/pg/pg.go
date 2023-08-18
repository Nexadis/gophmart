package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/Nexadis/gophmart/internal/order"
	"github.com/Nexadis/gophmart/internal/user"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const SchemaUsers = `CREATE TABLE Users(
"login" VARCHAR(256) PRIMARY KEY,
"hashpass" VARCHAR(256) NOT NULL);
`

const SchemaOrders = `CREATE TABLE Orders(
	"number" VARCHAR(256) PRIMARY KEY,
	"owner" VARCHAR(256),
	"status" VARCHAR(256) NOT NULL,
	"accrual" INT,
	"upload_at" TIMESTAMP NOT NULL);`

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
	_, err = pgx.Exec(SchemaUsers)
	if err != nil {
		logger.Logger.Errorln(err)
	}
	_, err = pgx.Exec(SchemaOrders)
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

func (pg *PG) GetUser(ctx context.Context, login string) (*user.User, error) {
	stmt, err := pg.db.Prepare("SELECT login, hashpass FROM Users WHERE login=$1")
	if err != nil {
		return nil, err
	}
	u := new(user.User)
	row := stmt.QueryRowContext(ctx, login)
	err = row.Scan(&u.Login, &u.HashPass)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", db.ErrSomeWrong, err)
	}
	return u, nil
}

func (pg *PG) AddOrder(ctx context.Context, o *order.Order) error {
	stmt, err := pg.db.Prepare("INSERT INTO Orders(number, owner, status, accrual, upload_at) values($1,$2,$3,$4,$5)")
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx,
		o.Number,
		o.Owner,
		o.Status,
		o.Accrual,
		o.UploadAt,
	)
	if err != nil {
		logger.Logger.Error(err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			existOrder, err := pg.GetOrder(ctx, o.Number)
			if err != nil {
				return fmt.Errorf("%s: %w", db.ErrSomeWrong, err)
			}
			if existOrder.Number == o.Number {
				return db.ErrOrderAdded
			}
			return db.ErrOtherUserOrder
		}
		return fmt.Errorf("%s: %w", db.ErrSomeWrong, err)
	}
	return nil
}

func (pg *PG) GetOrder(ctx context.Context, number string) (*order.Order, error) {
	stmt, err := pg.db.Prepare("SELECT number, owner, status, accrual, upload_at FROM Orders WHERE number=$1 ORDER BY upload_at")
	if err != nil {
		return nil, err
	}
	o := &order.Order{}
	row := stmt.QueryRowContext(ctx, number)
	err = row.Scan(&o.Number, &o.Owner, &o.Status, &o.Accrual, &o.UploadAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, db.ErrOrderNotFound
		}
		return nil, fmt.Errorf("%s: %w", db.ErrSomeWrong, err)
	}
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (pg *PG) GetOrders(ctx context.Context, owner string) ([]*order.Order, error) {
	stmt, err := pg.db.Prepare("SELECT number, owner, status, accrual, upload_at FROM Orders WHERE owner=$1 ORDER BY upload_at DESC")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	orders := make([]*order.Order, 0, len(columns))

	for rows.Next() {
		o := &order.Order{}
		err = rows.Scan(&o.Number, &o.Owner, &o.Status, &o.Accrual, &o.UploadAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", db.ErrSomeWrong, err)
		}
		orders = append(orders, o)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}
