package server

import "github.com/labstack/echo/v4"

type Server struct {
	e      *echo.Echo
	config *Config
}

func New(config *Config) Server {
	e := echo.New()
	return Server{
		e:      e,
		config: config,
	}
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
