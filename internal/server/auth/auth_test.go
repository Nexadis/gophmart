package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Nexadis/gophmart/internal/user"
)

type want struct {
	isValid bool
}

type testCase struct {
	name string
	user *user.User
	want want
}

var testSecret = []byte("secret")

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
			tokenString, err := NewToken(test.user.Login, testSecret)
			assert.NoError(t, err)
			isValid := IsValidToken(tokenString, testSecret)
			assert.Equal(t, test.want.isValid, isValid)
		})
	}
}
