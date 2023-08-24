package client

import (
	"github.com/Nexadis/gophmart/internal/order"
)

const (
	StatusRegistered = `REGISTERED`
	StatusInvalid    = `INVALID`
	StatusProcessing = `PROCESSING`
	StatusProcessed  = `PROCESSED`
)

type Accrual struct {
	Order   order.OrderNumber `json:"order"`
	Status  string            `json:"status"`
	Accrual *order.Points     `json:"accrual,omitempty"`
}
