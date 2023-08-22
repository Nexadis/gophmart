package client

import (
	"encoding/json"

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
	Accrual int64             `json:"accrual"`
}

type jsonAccrual struct {
	Order   order.OrderNumber `json:"order"`
	Status  string            `json:"status"`
	Accrual float64           `json:"accrual"`
}

func (a Accrual) MarshalJSON() ([]byte, error) {
	j := jsonAccrual{
		a.Order,
		a.Status,
		float64(a.Accrual) / 100,
	}
	return json.Marshal(j)
}

func (a *Accrual) UnmarshalJSON(data []byte) error {
	j := &jsonAccrual{}
	err := json.Unmarshal(data, j)
	if err != nil {
		return nil
	}
	a.Order = j.Order
	a.Status = j.Status
	a.Accrual = int64(j.Accrual * 100)
	return nil
}
