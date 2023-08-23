package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Nexadis/gophmart/internal/db"
	"github.com/Nexadis/gophmart/internal/logger"
	"github.com/Nexadis/gophmart/internal/order"
	"github.com/Nexadis/gophmart/internal/server/auth"
	"github.com/Nexadis/gophmart/internal/user"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"
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
	logger.Logger.Debugf("Register user:%v", *u)

	token, err := auth.NewToken(u.Login, JwtSecret)
	if err != nil {
		logger.Logger.Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	resp := echo.NewResponse(c.Response().Writer, s.e)
	resp.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	c.SetResponse(resp)
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (s *Server) UserLogin(c echo.Context) error {
	u := new(user.User)
	if err := c.Bind(u); err != nil {
		logger.Logger.Errorln(err)
		return c.String(http.StatusBadRequest, InvalidReq)
	}
	logger.Logger.Debug("User Login:", *u)
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
	resp := echo.NewResponse(c.Response().Writer, s.e)
	resp.Header().Add("Authorization", fmt.Sprintf("Bearer %s", token))
	c.SetResponse(resp)
	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (s *Server) UserOrdersSave(c echo.Context) error {
	req := c.Request()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer req.Body.Close()
	login, err := auth.GetLogin(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	regOrder, err := order.New(string(body), login)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrInvalidNum):
			return c.String(http.StatusUnprocessableEntity, err.Error())
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
	req := c.Request()
	login, err := auth.GetLogin(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	orders, err := s.db.GetOrders(req.Context(), login)
	if err != nil {
		logger.Logger.Error(err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if len(orders) == 0 {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, orders)
}

func (s *Server) UserBalance(c echo.Context) error {
	req := c.Request()
	login, err := auth.GetLogin(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	balance, err := getBalance(req.Context(), s.db, login)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, balance)
}

func (s *Server) UserBalanceWithdraw(c echo.Context) error {
	req := c.Request()
	login, err := auth.GetLogin(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	w := &order.Withdraw{}
	err = c.Bind(w)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if !w.Order.IsValid() {
		return c.NoContent(http.StatusUnprocessableEntity)
	}
	w.Owner = login
	t := time.Now()
	w.ProcessedAt = &t
	balance, err := getBalance(req.Context(), s.db, login)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if w.Sum > balance.Current {
		return c.String(http.StatusPaymentRequired, "not enough balance")
	}
	err = s.db.AddWithdrawal(req.Context(), w)
	if err != nil {
		if errors.Is(err, db.ErrWithdrawAdded) {
			return c.String(http.StatusConflict, err.Error())
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) UserWithdrawals(c echo.Context) error {
	req := c.Request()
	login, err := auth.GetLogin(c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	withdrawals, err := s.db.GetWithdrawals(req.Context(), login)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if len(withdrawals) == 0 {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, withdrawals)
}

func getBalance(ctx context.Context, db db.Database, owner string) (*user.Balance, error) {
	accruals, err := db.GetAccruals(ctx, owner)
	if err != nil {
		return nil, err
	}
	withdrawn, err := db.GetWithdrawn(ctx, owner)
	if err != nil {
		return nil, err
	}

	return &user.Balance{
		Current:   accruals - withdrawn,
		Withdrawn: withdrawn,
	}, nil
}
