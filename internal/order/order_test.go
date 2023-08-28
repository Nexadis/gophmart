package order

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type want struct {
	isValid bool
}

var tests = []struct {
	name   string
	number string
	want   want
}{
	{
		"test1",
		"12345678902",
		want{
			false,
		},
	},
	{
		"test2",
		"25461716",
		want{
			true,
		},
	},
}

func TestOrderNumber(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			orNum := OrderNumber(test.number)
			valid := orNum.IsValid()
			assert.Equal(t, test.want.isValid, valid)
		})
	}
}

var (
	value       = Points(12312)
	res, _      = json.Marshal(123.12)
	checkPoints = []struct {
		name string
		p    *Points
		res  []byte
	}{
		{
			"Normal value",
			&value,
			res,
		},
		{
			"Nil value",
			nil,
			[]byte(`null`),
		},
	}
)

func TestPoints(t *testing.T) {
	for _, test := range checkPoints {
		t.Run(test.name, func(t *testing.T) {
			jsoned, err := json.Marshal(test.p)
			if assert.NoError(t, err) {
				assert.Equal(t, test.res, jsoned)
			}
		})
	}
}
