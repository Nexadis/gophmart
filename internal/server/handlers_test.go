package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/order"
	"github.com/Nexadis/gophmart/internal/user"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var jsonHeaders = []http.Header{
	{
		echo.HeaderContentType: []string{
			echo.MIMEApplicationJSON,
		},
	},
}

var defaultDB = testDB{
	Users: map[string]*user.User{
		"admin": {
			Login:    "admin",
			Password: "secretpassword",
		},
	},
}

type testDB struct {
	Users  map[string]*user.User
	Orders map[string]*order.Order
}

func newTestDB() *testDB {
	users := make(map[string]*user.User, 10)
	orders := make(map[string]*order.Order, 10)
	db := testDB{
		users,
		orders,
	}
	return &db
}

func (tdb testDB) Open(addr string) error {
	return nil
}

func (tdb testDB) Close() error {
	return nil
}

func (tdb *testDB) AddUser(ctx context.Context, user *user.User) error {
	if _, ok := tdb.Users[user.Login]; ok {
		return db.ErrUserIsExist
	}
	tdb.Users[user.Login] = user
	return nil
}

func (tdb *testDB) GetUser(ctx context.Context, login string) (*user.User, error) {
	var u *user.User
	var ok bool
	var err error
	if u, ok = tdb.Users[login]; !ok {
		return nil, db.ErrUserNotFound
	}
	u.Password, err = u.HashPassword()
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (tdb *testDB) AddOrder(ctx context.Context, o *order.Order) error {
	return nil
}

func (tdb *testDB) AddOrders(ctx context.Context, orders []*order.Order) error {
	return nil
}

func (tdb *testDB) GetOrder(ctx context.Context, number string) (*order.Order, error) {
	return nil, nil
}

func (tdb *testDB) GetOrders(ctx context.Context, number []string) ([]*order.Order, error) {
	return nil, nil
}

type want struct {
	status   int
	response string
	err      error
	db       testDB
}

type request struct {
	method  string
	URI     string
	body    string
	headers []http.Header
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
			db:       defaultDB,
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
			db:       defaultDB,
		},
	},
}

func newTestServer() *Server {
	s := &Server{
		e: echo.New(),
	}
	s.MountHandlers()
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
	s := newTestServer()
	db := newTestDB()
	s.db = db
	for _, test := range testsUserRegister {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.r.method, test.r.URI, strings.NewReader(test.r.body))
			setHeaders(req, test.r.headers)
			rec := httptest.NewRecorder()
			c := s.e.NewContext(req, rec)
			if assert.NoError(t, s.UserRegister(c)) {
				assert.Equal(t, test.want.status, rec.Code)
				assert.Equal(t, test.want.db.Users, db.Users)
				body, _ := io.ReadAll(rec.Body)
				assert.Equal(t, test.want.response, string(body))
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
			db:     defaultDB,
		},
	}, {
		name: "Login invalid password",
		r: *newUserLogin(
			`{"login":"admin","password":"invalid"}`,
		),
		want: want{
			status: http.StatusUnauthorized,
			err:    nil,
			db:     defaultDB,
		},
	}, {
		name: "Login invalid user",
		r: *newUserLogin(
			`{"login":"user","password":"invalid"}`,
		),
		want: want{
			status: http.StatusUnauthorized,
			err:    db.ErrUserNotFound,
			db:     defaultDB,
		},
	},
}

func TestUserLogin(t *testing.T) {
	s := newTestServer()
	for _, test := range testsUserLogin {
		t.Run(test.name, func(t *testing.T) {
			s.db = &test.want.db
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
				assert.Equal(t, test.want.db.Users, test.want.db.Users)
			}
		})
	}
}
