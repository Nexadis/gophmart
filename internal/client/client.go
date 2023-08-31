package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/Nexadis/gophmart/internal/order"
)

const APIGetAccrual = `/api/orders/{number}`

var (
	ErrInternal      = errors.New(`internal error accrual`)
	ErrNotRegistered = errors.New(`order isn't registered in system`)
)

type Client struct {
	client *resty.Client
	Addr   string
	db     db.OrdersStore
	wait   time.Duration
}

func New(addr string, db db.OrdersStore, wait time.Duration) *Client {
	return &Client{
		client: resty.New().SetDebug(true),
		Addr:   addr,
		db:     db,
		wait:   wait,
	}
}

func (c *Client) GetAccruals(done chan struct{}, errors chan error) {
	orders := make(chan order.OrderNumber)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		t := time.NewTicker(c.wait)
		for {
			select {
			case <-done:
				close(orders)
				wg.Done()
				return
			case <-t.C:
				PrcessingOrders, err := c.db.GetWithStatus(context.Background(), order.StatusProcessing)
				if err != nil {
					logger.Logger.Error(err)
					continue
				}
				NewOrders, err := c.db.GetWithStatus(context.Background(), order.StatusNew)
				if err != nil {
					logger.Logger.Error(err)
					continue
				}
				Orders := append(PrcessingOrders, NewOrders...)
				for _, number := range Orders {
					orders <- number
				}
			}
		}
	}()
	for {
		select {
		case n := <-orders:
			o := &order.Order{
				Number: n,
			}
			a, err := c.getOrderStatus(n)
			switch err {
			case nil:
				o.Status = accrualToOrderStatus(a.Status)
				o.Accrual = a.Accrual
			case ErrNotRegistered:
				o.Status = order.StatusInvalid
			default:
				logger.Logger.Error(err)
				continue
			}
			err = c.db.UpdateOrder(context.Background(), o)
			if err != nil {
				logger.Logger.Error(err)
			}
		case <-done:
			wg.Wait()
			return
		}
	}
}

func (c *Client) getOrderStatus(number order.OrderNumber) (*Accrual, error) {
	endpoint := fmt.Sprintf("%s%s", c.Addr, APIGetAccrual)
	a := &Accrual{}
	resp, err := c.client.R().
		SetPathParam("number", string(number)).
		SetResult(a).
		Get(endpoint)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return a, nil
	case http.StatusTooManyRequests:
		logger.Logger.Info("Too many requests")
		time.Sleep(60 * time.Second)
	case http.StatusNoContent:
		return nil, ErrNotRegistered
	}
	return nil, fmt.Errorf(`invlaid status code for order: %s`, number)
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
