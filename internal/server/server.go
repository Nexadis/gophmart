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
}
