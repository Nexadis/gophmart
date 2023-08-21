package order

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

type Status string

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusInvalid    Status = "INVALID"
	SatusProcessed   Status = "PROCESSED"
)

var statuses []Status

func init() {
	statuses = []Status{
		StatusNew,
		SatusProcessed,
		StatusInvalid,
		StatusProcessing,
	}
}

var (
	ErrInvalidNum    = errors.New(`invalid order number`)
	ErrInvalidStatus = errors.New(`invalid status`)
)

type OrderNumber string

type Order struct {
	Owner      string      `json:"-"`
	Number     OrderNumber `json:"number"`
	Status     Status      `json:"status"`
	Accrual    *int64      `json:"accrual,omitempty"`
	UploadedAt *time.Time  `json:"uploaded_at"`
}

func New(number, owner string) (*Order, error) {
	upload := time.Now()
	order := &Order{
		Owner:      owner,
		Number:     OrderNumber(number),
		Status:     StatusNew,
		Accrual:    nil,
		UploadedAt: &upload,
	}
	if order.IsValid() {
		return order, nil
	}
	return nil, ErrInvalidNum
}

func (o Order) IsValid() bool {
	return o.Number.IsValid()
}

func (o OrderNumber) IsValid() bool {
	digits := strings.Split(strings.ReplaceAll(string(o), " ", ""), "")
	lengthOfString := len(digits)

	if lengthOfString < 2 {
		return false
	}

	sum := 0
	flag := false

	for i := lengthOfString - 1; i > -1; i-- {
		digit, _ := strconv.Atoi(digits[i])

		if flag {
			digit *= 2

			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		flag = !flag
	}

	return math.Mod(float64(sum), 10) == 0
}

type jsonOrder struct {
	Number     OrderNumber `json:"number"`
	Status     Status      `json:"status"`
	Accrual    *float64    `json:"accrual,omitempty"`
	UploadedAt *time.Time  `json:"uploaded_at"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	var accrual *float64
	if o.Accrual != nil {
		a := float64(*o.Accrual) / 100
		accrual = &a
	}
	j := &jsonOrder{
		Number:     o.Number,
		Status:     o.Status,
		Accrual:    accrual,
		UploadedAt: o.UploadedAt,
	}
	return json.Marshal(j)
}

func (o *Order) UnmarshalJSON(data []byte) error {
	j := &jsonOrder{}
	err := json.Unmarshal(data, j)
	if err != nil {
		return err
	}
	if !IsValidStatus(j.Status) {
		return ErrInvalidStatus
	}
	o.Status = j.Status
	if j.Accrual != nil {
		accrual := int64(*j.Accrual * 100)
		o.Accrual = &accrual
	}
	o.Number = j.Number
	o.UploadedAt = j.UploadedAt
	return nil
}

func IsValidStatus(status Status) bool {
	for _, validStatus := range statuses {
		if status == validStatus {
			return true
		}
	}
	return false
}
