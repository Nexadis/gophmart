package server

import (
	"github.com/Nexadis/gophmart/internal/db"
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
	db := db.New()
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
	s.e.POST("/api/user/register", s.UserRegister)
	s.e.POST("/api/user/login", s.UserLogin)
	s.e.POST("/api/user/orders", s.UserOrdersSave)
	s.e.GET("/api/user/orders", s.UserOrdersGet)
	s.e.GET("/api/user/balance", s.UserBalance)
	s.e.POST("/api/user/balance/withdraw", s.UserBalanceWithdraw)
	s.e.GET("/api/user/withdrawals", s.UserWithdrawals)
}
