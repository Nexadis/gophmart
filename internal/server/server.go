package server

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/db/pg"
	"github.com/Nexadis/gophmart/internal/logger"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
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
	err := db.Open(config.DbURI)
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
	return s.e.Start(s.config.RunAddress)
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
	s.e.POST(ApiUserRegister, s.UserRegister)
	s.e.POST(ApiUserLogin, s.UserLogin)
	r := s.e.Group(ApiRestricted)
	{
		r.Use(echojwt.JWT(JwtSecret))
		r.POST(ApiUserOrders, s.UserOrdersSave)
		r.GET(ApiUserOrders, s.UserOrdersGet)
		r.GET(ApiUserBalance, s.UserBalance)
		r.POST(ApiUserBalanceWithdraw, s.UserBalanceWithdraw)
		r.GET(ApiUserWithdrawals, s.UserWithdrawals)
	}
}
