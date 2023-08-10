package auth

import (
	"testing"

	"github.com/Nexadis/gophmart/internal/user"
	"github.com/stretchr/testify/assert"
)

type want struct {
	isValid bool
}

type testCase struct {
	name string
	user *user.User
	want want
}

var tests = []testCase{
	{
		name: "Simple login with password",
		user: &user.User{
			Login:    "test",
			Password: "test",
		},
		want: want{
			isValid: true,
		},
	},
}

func TestToken(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tokenString, err := NewToken(test.user)
			assert.NoError(t, err)
			isValid := IsValidToken(tokenString, []byte(test.user.Password))
			assert.Equal(t, test.want.isValid, isValid)
		})
	}
}
