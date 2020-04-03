package stripehook

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/videocoin/cloud-billing/datastore"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"github.com/videocoin/cloud-billing/manager"
)

type Server struct {
	addr   string
	secret string
	logger *logrus.Entry
	e      *echo.Echo
	dm     *manager.Manager
}

func NewServer(addr string, opts ...Option) (*Server, error) {
	s := &Server{
		addr: addr,
		e:    echo.New(),
	}
	s.e.HideBanner = true
	s.e.HidePort = true
	s.e.DisableHTTP2 = true

	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Server) initRoutes() {
	s.e.POST("/api/v1/stripe/hooks", s.postHook)
}

func (s *Server) Start() error {
	s.logger.Infof("stripe hook server listening on %s", s.addr)
	s.initRoutes()
	return s.e.Start(s.addr)
}

func (s *Server) Stop() error {
	return nil
}

func (s *Server) postHook(c echo.Context) error {
	payload, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		s.logger.Errorf("failed to read body: %s", err)
		return echo.ErrBadRequest
	}

	event, err := webhook.ConstructEvent(payload, c.Request().Header.Get("Stripe-Signature"), s.secret)
	if err != nil {
		s.logger.Errorf("failed to construct event: %s", err)
		return echo.ErrBadRequest
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &pi)
		if err != nil {
			s.logger.Errorf("failed to unmarshal event: %s", err)
			return echo.ErrBadRequest
		}

		s.logger.Debugf("payment intent %+v", pi)

		ctx := context.Background()
		transaction, err := s.dm.GetTransactionByPaymentID(ctx, pi.ID)
		if err != nil {
			if err == datastore.ErrTxNotFound {
				return nil
			}
			s.logger.Errorf("failed to get transaction by payment id: %s", err)
			return echo.ErrInternalServerError
		}

		err = s.dm.MarkTransactionAsSucceded(ctx, transaction)
		if err != nil {
			s.logger.Errorf("failed to mark transaction as succeded: %s", err)
			return echo.ErrInternalServerError
		}
	default:
		s.logger.Infof("webhook event %s", event.Type)
	}

	return nil
}
