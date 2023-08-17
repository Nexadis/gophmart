package db

import (
	"context"
	"errors"

	"github.com/Nexadis/gophmart/internal/order"
	"github.com/Nexadis/gophmart/internal/user"
)

var (
	ErrUserIsExist    = errors.New(`user is exist`)
	ErrUserNotFound   = errors.New(`user not found`)
	ErrOrderNotFound  = errors.New(`order not found`)
	ErrOrderAdded     = errors.New(`order was added`)
	ErrOtherUserOrder = errors.New(`order was added by other user`)
	ErrSomeWrong      = errors.New(`some wrong`)
)

type UserStore interface {
	AddUser(ctx context.Context, user *user.User) error
	GetUser(ctx context.Context, login string) (*user.User, error)
}

type OrdersStore interface {
	AddOrder(ctx context.Context, o *order.Order) error
	AddOrders(ctx context.Context, orders []*order.Order) error
	GetOrder(ctx context.Context, number string) (*order.Order, error)
	GetOrders(ctx context.Context, number []string) ([]*order.Order, error)
}

type Database interface {
	Open(Addr string) error
	UserStore
	OrdersStore
	Close() error
}
