package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/order"
	"github.com/go-resty/resty/v2"
)

const APIGetAccrual = `/api/orders/{number}`

var ErrInternal = errors.New(`internal error accrual`)

type Client struct {
	client *resty.Client
	Addr   string
	db     db.OrdersStore
}

func New(addr string, db db.OrdersStore) *Client {
	return &Client{
		client: resty.New().SetDebug(true),
		Addr:   addr,
		db:     db,
	}
}

func (c *Client) GetAccruals(errors chan error) {
	endpoint := fmt.Sprintf("http://%s%s", c.Addr, APIGetAccrual)
	a := &Accrual{}
	orders := make(chan order.OrderNumber)
	go func() {
		for {
			orderNumbers, err := c.db.GetWithStatus(context.Background(), order.StatusProcessing)
			if err != nil {
				errors <- err
			}
			for _, number := range orderNumbers {
				orders <- number
			}
			orderNumbers, err = c.db.GetWithStatus(context.Background(), order.StatusNew)
			if err != nil {
				errors <- err
			}
			for _, number := range orderNumbers {
				orders <- number
			}
		}
	}()
	for orderNumber := range orders {
		resp, err := c.client.R().
			SetPathParam("number", string(orderNumber)).
			SetResult(a).
			Get(endpoint)
		if err != nil {
			errors <- err
			continue
		}
		o := &order.Order{
			Number: orderNumber,
		}

		switch resp.StatusCode() {
		case http.StatusOK:
			o.Accrual = &a.Accrual
			o.Status = accrualToOrderStatus(a.Status)
			err = c.db.UpdateOrder(context.Background(), o)
			if err != nil {
				errors <- err
			}
		case http.StatusTooManyRequests:
			time.Sleep(60 * time.Second)
		case http.StatusNoContent:
			err = ErrInternal
			errors <- fmt.Errorf(`%s order: %v`, err, orderNumber)
		}
	}
}

func accrualToOrderStatus(status string) order.Status {
	switch status {
	case StatusRegistered:
		return order.StatusNew
	case StatusProcessing:
		return order.StatusProcessing
	case StatusInvalid:
		return order.StatusInvalid
	case StatusProcessed:
		return order.StatusProcessed
	}
	return order.StatusInvalid
}
