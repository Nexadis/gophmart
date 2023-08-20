package order

import (
	"time"
)

type Withdraw struct {
	Owner       string      `json:"-"`
	Order       OrderNumber `json:"order"`
	Sum         int64       `json:"sum"`
	ProcessedAt *time.Time  `json:"processed_at"`
}
