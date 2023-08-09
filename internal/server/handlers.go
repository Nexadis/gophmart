package server

import (
	"errors"
	"net/http"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/Nexadis/gophmart/internal/user"
	"github.com/labstack/echo/v4"
)

const InvalidReq = "invalid request"

func (s *Server) UserRegister(c echo.Context) error {
	logger.Logger.Infoln("Got Register request")
	u := new(user.User)
	if err := c.Bind(u); err != nil {
		logger.Logger.Errorln(err)
		return c.String(http.StatusBadRequest, InvalidReq)
	}
	err := s.db.AddUser(c.Request().Context(), u)
	if err != nil {
		logger.Logger.Errorln(err)
		switch {
		case errors.Is(err, db.ErrUserIsExist):
			return c.String(http.StatusConflict, err.Error())
		case errors.Is(err, db.ErrSomeWrong):
			return c.String(http.StatusInternalServerError, err.Error())
		default:
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) UserLogin(c echo.Context) error {
	return nil
}

func (s *Server) UserOrdersSave(c echo.Context) error {
	return nil
}

func (s *Server) UserOrdersGet(c echo.Context) error {
	return nil
}

func (s *Server) UserBalance(c echo.Context) error {
	return nil
}

func (s *Server) UserBalanceWithdraw(c echo.Context) error {
	return nil
}

func (s *Server) UserWithdrawals(c echo.Context) error {
	return nil
}
