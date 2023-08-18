package order

import (
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

var ErrInvalidNum = errors.New(`invalid order number`)

type Order struct {
	Owner    string     `json:"-"`
	Number   string     `json:"number"` // написать альтернативный маршалер в string
	Status   Status     `json:"status"`
	Accrual  *int64     `json:"accrual,omitempty"`
	UploadAt *time.Time `json:"upload_at"`
}

type OrderTime string

func (ot OrderTime) Time() (time.Time, error) {
	return time.Parse(time.RFC3339, string(ot))
}

func New(number, owner string) (*Order, error) {
	upload := time.Now()
	order := &Order{
		Owner:    owner,
		Number:   number,
		Status:   StatusNew,
		Accrual:  nil,
		UploadAt: &upload,
	}
	if order.IsValid() {
		return order, nil
	}
	return nil, ErrInvalidNum
}

func (o Order) IsValid() bool {
	digits := strings.Split(strings.ReplaceAll(o.Number, " ", ""), "")
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
