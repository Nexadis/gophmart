package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/order"
	"github.com/Nexadis/gophmart/internal/server/auth"
	"github.com/Nexadis/gophmart/internal/user"
	"github.com/Nexadis/gophmart/mocks"
)

var jwtTestSecret = `supersecret`

var jsonHeaders = []http.Header{
	{
		echo.HeaderContentType: []string{
			echo.MIMEApplicationJSON,
		},
	},
}

var defaultUser = &user.User{
	Login:    "admin",
	Password: "secretpassword",
}

var otherUser = &user.User{
	Login:    "otheruser",
	Password: "otherpassword",
}

type want struct {
	status   int
	response string
	body     string
	err      error
}

type request struct {
	method  string
	URI     string
	body    string
	headers []http.Header
	user    *user.User
}

type testCase struct {
	name string
	r    request
	want want
}

func newUserRegister(body string) *request {
	return &request{
		method:  http.MethodPost,
		URI:     APIUserRegister,
		body:    body,
		headers: jsonHeaders,
	}
}

var testsUserRegister = []testCase{
	{
		name: "Register new user",
		r: *newUserRegister(
			`{"login":"admin","password":"secretpassword"}`,
		),
		want: want{
			status:   http.StatusOK,
			response: "",
			err:      nil,
		},
	}, {
		name: "Duplicate user",
		r: *newUserRegister(
			`{"login":"admin","password":"secretpassword"}`,
		),
		want: want{
			status:   http.StatusConflict,
			response: db.ErrUserIsExist.Error(),
			err:      db.ErrUserIsExist,
		},
	},
}

func newTestServer() *Server {
	s := &Server{
		e:      echo.New(),
		config: &Config{},
	}
	JwtSecret = []byte(jwtTestSecret)
	prepareServer(s)
	return s
}

func setHeaders(r *http.Request, headers []http.Header) {
	for _, header := range headers {
		for key, values := range header {
			for _, value := range values {
				r.Header.Add(key, value)
			}
		}
	}
}

func TestUserRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockdb := mocks.NewMockDatabase(ctrl)
	s := newTestServer()
	gomock.InOrder(
		mockdb.EXPECT().AddUser(context.Background(), defaultUser).Return(nil),
		mockdb.EXPECT().AddUser(context.Background(), defaultUser).Return(db.ErrUserIsExist),
	)
	s.db = mockdb
	for _, test := range testsUserRegister {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.r.method, test.r.URI, strings.NewReader(test.r.body))
			setHeaders(req, test.r.headers)
			rec := httptest.NewRecorder()
			c := s.e.NewContext(req, rec)
			if assert.NoError(t, s.UserRegister(c)) {
				assert.Equal(t, test.want.status, rec.Code)
			}
		})
	}
}

func newUserLogin(body string) *request {
	return &request{
		method:  http.MethodPost,
		URI:     APIUserLogin,
		body:    body,
		headers: jsonHeaders,
	}
}

var testsUserLogin = []testCase{
	{
		name: "Login Valid User",
		r: *newUserLogin(
			`{"login":"admin","password":"secretpassword"}`,
		),
		want: want{
			status: http.StatusOK,
			err:    nil,
		},
	}, {
		name: "Login invalid password",
		r: *newUserLogin(
			`{"login":"admin","password":"invalid"}`,
		),
		want: want{
			status: http.StatusUnauthorized,
			err:    nil,
		},
	}, {
		name: "Login invalid user",
		r: *newUserLogin(
			`{"login":"user","password":"invalid"}`,
		),
		want: want{
			status: http.StatusUnauthorized,
			err:    db.ErrUserNotFound,
		},
	},
}

func TestUserLogin(t *testing.T) {
	s := newTestServer()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockdb := mocks.NewMockDatabase(ctrl)
	defaultUser.HashPassword()
	gomock.InOrder(
		mockdb.EXPECT().GetUser(context.Background(), defaultUser.Login).Return(defaultUser, nil),
		mockdb.EXPECT().GetUser(context.Background(), defaultUser.Login).Return(defaultUser, nil),
		mockdb.EXPECT().GetUser(context.Background(), `user`).Return(nil, db.ErrUserNotFound),
	)
	s.db = mockdb
	for _, test := range testsUserLogin {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.r.method, test.r.URI, strings.NewReader(test.r.body))
			setHeaders(req, test.r.headers)
			rec := httptest.NewRecorder()
			c := s.e.NewContext(req, rec)
			err := s.UserLogin(c)
			if test.want.err != nil {
				assert.Error(t, err)
				return
			}
			if assert.NoError(t, err) {
				assert.Equal(t, test.want.status, rec.Code)
			}
		})
	}
}

func setLogin(c echo.Context, login string) error {
	tokenString, err := auth.NewToken(login, JwtSecret)
	token, _ := auth.GetToken(tokenString, JwtSecret)
	token.Claims = jwt.MapClaims{
		"login": login,
	}
	c.Set("user", token)
	if err != nil {
		return err
	}
	return nil
}

var testOrdersSave = []testCase{
	{
		name: "Normal order number",
		r: request{
			method: http.MethodPost,
			URI:    APIRestricted + APIUserOrders,
			body:   "445084503850",
			user:   defaultUser,
		},
		want: want{
			status: http.StatusAccepted,
			body:   "",
			err:    nil,
		},
	},
	{
		name: "Normal order added by this user",
		r: request{
			method: http.MethodPost,
			URI:    APIRestricted + APIUserOrders,
			body:   "445084503850",
			user:   defaultUser,
		},
		want: want{
			status: http.StatusOK,
			body:   db.ErrOrderAdded.Error(),
			err:    nil,
		},
	},
	{
		name: "Order added by other user",
		r: request{
			method: http.MethodPost,
			URI:    APIRestricted + APIUserOrders,
			body:   "445084503850",
			user:   otherUser,
		},
		want: want{
			status: http.StatusConflict,
			body:   db.ErrOtherUserOrder.Error(),
			err:    nil,
		},
	},
	{
		name: "Invalid number",
		r: request{
			method: http.MethodPost,
			URI:    APIRestricted + APIUserOrders,
			body:   "445084503851",
			user:   otherUser,
		},
		want: want{
			status: http.StatusUnprocessableEntity,
			body:   order.ErrInvalidNum.Error(),
			err:    nil,
		},
	},
}

type orderMatcher struct {
	Number string
	Owner  string
}

func mockOrder(number, owner string) gomock.Matcher {
	return &orderMatcher{
		Number: number,
		Owner:  owner,
	}
}

func (om *orderMatcher) Matches(x interface{}) bool {
	o, ok := x.(*order.Order)
	if !ok {
		return false
	}
	if string(o.Number) != om.Number {
		return false
	}
	if o.Owner != om.Owner {
		return false
	}

	return true
}

func (om *orderMatcher) String() string {
	return fmt.Sprintf("Match order, owner: %s, number: %s", om.Owner, om.Number)
}

func TestOrdersSave(t *testing.T) {
	s := newTestServer()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockdb := mocks.NewMockDatabase(ctrl)
	s.db = mockdb
	defaultUser.HashPassword()
	gomock.InOrder(
		mockdb.EXPECT().AddOrder(
			context.Background(),
			mockOrder("445084503850", defaultUser.Login),
		).Return(nil),
		mockdb.EXPECT().AddOrder(
			context.Background(),
			mockOrder("445084503850", defaultUser.Login),
		).Return(db.ErrOrderAdded),
		mockdb.EXPECT().AddOrder(
			context.Background(),
			mockOrder("445084503850", otherUser.Login),
		).Return(db.ErrOtherUserOrder),
	)
	for _, test := range testOrdersSave {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.r.method, test.r.URI, strings.NewReader(test.r.body))
			setHeaders(req, test.r.headers)
			rec := httptest.NewRecorder()
			c := s.e.NewContext(req, rec)
			setLogin(c, test.r.user.Login)
			err := s.UserOrdersSave(c)
			if test.want.err != nil {
				assert.Equal(t, test.want.err, err)
				return
			}
			if assert.NoError(t, err) {
				assert.Equal(t, test.want.status, rec.Code)
			}
			body, _ := io.ReadAll(rec.Body)
			assert.Equal(t, []byte(test.want.body), body)
		})
	}
}

func TestOrdersGet(t *testing.T) {
	s := newTestServer()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockdb := mocks.NewMockDatabase(ctrl)
	s.db = mockdb
	defaultUser.HashPassword()
	gomock.InOrder(
		mockdb.EXPECT().AddOrder(
			context.Background(),
			gomock.Any(),
		).Return(nil),
		mockdb.EXPECT().AddOrder(
			context.Background(),
			gomock.Any(),
		).Return(db.ErrOrderAdded),
		mockdb.EXPECT().AddOrder(
			context.Background(),
			gomock.Any(),
		).Return(db.ErrOtherUserOrder),
	)
	for _, test := range testOrdersSave {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.r.method, test.r.URI, strings.NewReader(test.r.body))
			setHeaders(req, test.r.headers)
			rec := httptest.NewRecorder()
			c := s.e.NewContext(req, rec)
			setLogin(c, defaultUser.Login)
			err := s.UserOrdersSave(c)
			if test.want.err != nil {
				assert.Equal(t, test.want.err, err)
				return
			}
			if assert.NoError(t, err) {
				assert.Equal(t, test.want.status, rec.Code)
			}
			body, _ := io.ReadAll(rec.Body)
			assert.Equal(t, []byte(test.want.body), body)
		})
	}
}

func TestUserBalance(t *testing.T) {
}

func TestUserBalanceWithdraw(t *testing.T) {
}

func TestUserWithDrawals(t *testing.T) {
}
