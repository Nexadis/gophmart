package server

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/Nexadis/gophmart/internal/client"
	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/db/pg"
	"github.com/Nexadis/gophmart/internal/logger"
)

type Server struct {
	e      *echo.Echo
	config *Config
	db     db.Database
}

const secretLen = 32

var JwtSecret []byte

func New(config *Config) (*Server, error) {
	e := echo.New()
	db := pg.New()
	err := db.Open(config.DBURI)
	if err != nil {
		logger.Logger.Infoln(`can't connect to DB`)
		return nil, err
	}
	return &Server{
		e:      e,
		config: config,
		db:     db,
	}, nil
}

func (s *Server) Run() error {
	prepareServer(s)
	errors := make(chan error)
	done := make(chan struct{})
	client := client.New(s.config.AccrualSystemAddress, s.db, time.Duration(s.config.Wait)*time.Second)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		client.GetAccruals(done, errors)
		wg.Done()
	}()
	err := s.e.Start(s.config.RunAddress)
	close(done)
	wg.Wait()
	close(errors)
	return err
}

func prepareServer(s *Server) {
	JwtSecret = []byte(s.config.JwtSecret)
	if s.config.JwtSecret == "" {
		secret := make([]byte, secretLen)
		rand.Read(secret)
		JwtSecret = secret
		logger.Logger.Infof("Set Secret '%s'", hex.EncodeToString(JwtSecret))
	}
	s.MountHandlers()
}

func (s *Server) MountHandlers() {
	s.e.Use(middleware.Logger())
	s.e.Use(middleware.Gzip())
	s.e.POST(APIUserRegister, s.UserRegister)
	s.e.POST(APIUserLogin, s.UserLogin)
	r := s.e.Group(APIRestricted)
	{
		r.Use(echojwt.JWT(JwtSecret))
		r.POST(APIUserOrders, s.UserOrdersSave)
		r.GET(APIUserOrders, s.UserOrdersGet)
		r.GET(APIUserBalance, s.UserBalance)
		r.POST(APIUserBalanceWithdraw, s.UserBalanceWithdraw)
		r.GET(APIUserWithdrawals, s.UserWithdrawals)
	}
}
