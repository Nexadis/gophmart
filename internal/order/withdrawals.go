package order

import (
	"encoding/json"
	"time"
)

type Withdraw struct {
	Owner       string      `json:"-"`
	Order       OrderNumber `json:"order"`
	Sum         int64       `json:"sum"`
	ProcessedAt *time.Time  `json:"processed_at"`
}

type jsonWithdraw struct {
	Owner       string      `json:"-"`
	Order       OrderNumber `json:"order"`
	Sum         float64     `json:"sum"`
	ProcessedAt *time.Time  `json:"processed_at"`
}

func (w Withdraw) MarshalJSON() ([]byte, error) {
	j := &jsonWithdraw{
		Owner:       w.Owner,
		Order:       w.Order,
		Sum:         float64(w.Sum) / 100,
		ProcessedAt: w.ProcessedAt,
	}
	return json.Marshal(j)
}

func (w *Withdraw) UnmarshalJSON(data []byte) error {
	j := &jsonWithdraw{}
	err := json.Unmarshal(data, j)
	if err != nil {
		return err
	}
	if !j.Order.IsValid() {
		return ErrInvalidNum
	}
	w.Owner = j.Owner
	w.Order = j.Order
	w.Sum = int64(j.Sum * 100)
	if j.ProcessedAt == nil {
		t := time.Now()
		j.ProcessedAt = &t
	}
	w.ProcessedAt = j.ProcessedAt
	return nil
}
