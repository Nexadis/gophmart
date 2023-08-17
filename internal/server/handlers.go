package server

import (
	"errors"
	"io"
	"net/http"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/Nexadis/gophmart/internal/order"
	"github.com/Nexadis/gophmart/internal/server/auth"
	"github.com/Nexadis/gophmart/internal/user"
	"github.com/labstack/echo/v4"
)

const InvalidReq = "invalid request"

func (s *Server) UserRegister(c echo.Context) error {
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

	token, err := auth.NewToken(u.Login, JwtSecret)
	if err != nil {
		logger.Logger.Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	cookie := auth.CookieToken(token)
	c.SetCookie(cookie)

	return c.NoContent(http.StatusOK)
}

func (s *Server) UserLogin(c echo.Context) error {
	u := new(user.User)
	if err := c.Bind(u); err != nil {
		logger.Logger.Errorln(err)
		return c.String(http.StatusBadRequest, InvalidReq)
	}
	savedUser, err := s.db.GetUser(c.Request().Context(), u.Login)
	if err != nil {
		logger.Logger.Error(err)
		switch {
		case errors.Is(err, db.ErrUserNotFound):
			c.NoContent(http.StatusUnauthorized)
		default:
			c.NoContent(http.StatusInternalServerError)
		}
		return err
	}
	if !u.IsValidHash(savedUser.HashPass) {
		return c.NoContent(http.StatusUnauthorized)
	}
	token, err := auth.NewToken(u.Login, JwtSecret)
	if err != nil {
		logger.Logger.Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	logger.Logger.Infof("User '%s' authorized. Token:'%s'", savedUser.Login, token)
	cookie := auth.CookieToken(token)
	c.SetCookie(cookie)
	return c.NoContent(http.StatusOK)
}

func (s *Server) UserOrdersSave(c echo.Context) error {
	req := c.Request()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer req.Body.Close()
	logger.Logger.Info(body)
	header := req.Header.Get(echo.HeaderAuthorization)
	login, err := auth.GetLogin(header)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	regOrder, err := order.New(string(body), login)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrInvalidNum):
			return c.String(http.StatusNotAcceptable, err.Error())
		}
		return c.String(http.StatusBadRequest, err.Error())
	}
	logger.Logger.Infof("Save order for %s, order %s", login, regOrder.Number)
	err = s.db.AddOrder(req.Context(), regOrder)
	if err != nil {
		logger.Logger.Error(err)
		switch {
		case errors.Is(err, db.ErrOrderAdded):
			return c.String(http.StatusOK, err.Error())
		case errors.Is(err, db.ErrOtherUserOrder):
			return c.String(http.StatusConflict, err.Error())
		}
	}

	return c.NoContent(http.StatusAccepted)
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
