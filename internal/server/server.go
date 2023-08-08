package server

import "github.com/labstack/echo/v4"

type Server struct {
	e *echo.Echo
}

func New(config *Config) Server {
	e := echo.New()
	return Server{
		e: e,
	}
}
