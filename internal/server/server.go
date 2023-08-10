package server

import (
	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/db/pg"
	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/labstack/echo/v4"
)

type Server struct {
	e      *echo.Echo
	config *Config
	db     db.Database
}

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
	s.MountHandlers()
	return s.e.Start(s.config.RunAddress)
}

func (s *Server) MountHandlers() {
	s.e.POST(ApiUserRegister, s.UserRegister)
	s.e.POST(ApiUserLogin, s.UserLogin)
	s.e.POST(ApiUserOrders, s.UserOrdersSave)
	s.e.GET(ApiUserOrders, s.UserOrdersGet)
	s.e.GET(ApiUserBalance, s.UserBalance)
	s.e.POST(ApiUserBalanceWithdraw, s.UserBalanceWithdraw)
	s.e.GET(ApiUserWithdrawals, s.UserWithdrawals)
}
