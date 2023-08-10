package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type want struct {
	hash string
	err  error
}

type testCase struct {
	name string
	User User
	want want
}

var hashTests = []testCase{
	{
		name: "Simple password",
		User: User{
			Login:    "admin",
			Password: "123451245",
		},
		want: want{
			hash: "",
			err:  nil,
		},
	},
}

func TestIsValidPassword(t *testing.T) {
	for _, test := range hashTests {
		t.Run(test.name, func(t *testing.T) {
			hash, err := test.User.HashPassword()
			t.Logf("Hash:%s, len:%d", hash, len(hash))
			if assert.NoError(t, err) {
				assert.Equal(t, test.User.IsValidPassword(hash), true)
			}
		})
	}
}
