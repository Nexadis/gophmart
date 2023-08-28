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

type Points int64

const (
	StatusNew        Status = "NEW"
	StatusProcessing Status = "PROCESSING"
	StatusInvalid    Status = "INVALID"
	StatusProcessed  Status = "PROCESSED"
)

var Statuses []Status

func init() {
	Statuses = []Status{
		StatusNew,
		StatusProcessed,
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
	Accrual    *Points     `json:"accrual,omitempty"`
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

func (p Points) MarshalJSON() ([]byte, error) {
	points := float64(p) / 100
	return json.Marshal(points)
}

func (p *Points) UnmarshalJSON(data []byte) error {
	var points *float64
	err := json.Unmarshal(data, *points)
	if err != nil {
		return err
	}
	ptmp := Points(*points * 100)
	*p = ptmp
	return nil
}
