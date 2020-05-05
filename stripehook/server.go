package stripehook

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	emitterv1 "github.com/videocoin/cloud-api/emitter/v1"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-billing/manager"
)

type Server struct {
	addr    string
	secret  string
	logger  *logrus.Entry
	e       *echo.Echo
	dm      *manager.Manager
	emitter emitterv1.EmitterServiceClient
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
			s.logger.WithError(err).Errorf("failed to unmarshal event")
			return echo.ErrBadRequest
		}

		s.logger.Debugf("payment intent %+v", pi)

		ctx := context.Background()

		logger := s.logger.WithField("pi_id", pi.ID)

		transaction, err := s.dm.GetTransactionByPaymentID(ctx, pi.ID)
		if err != nil {
			if err == datastore.ErrTxNotFound {
				return nil
			}
			logger.WithError(err).Errorf("failed to get transaction by payment id")
			return echo.ErrInternalServerError
		}

		err = s.dm.MarkTransactionAsSucceded(ctx, transaction)
		if err != nil {
			logger.WithError(err).Error("failed to mark transaction as succeded")
			return echo.ErrInternalServerError
		}

		logger = logger.WithField("to", transaction.To)

		account, err := s.dm.GetAccountByID(ctx, transaction.To)
		if err != nil {
			logger.WithError(err).Errorf("failed to mark transaction as succeded")
			return echo.ErrInternalServerError
		}

		logger = logger.
			WithField("account_id", account.ID).
			WithField("user_id", account.UserID)

		afReq := &emitterv1.AddFundsRequest{
			UserID:    account.UserID,
			AmountUsd: transaction.Amount / 100,
		}
		_, err = s.emitter.AddFunds(ctx, afReq)
		if err != nil {
			logger.WithError(err).Error("failed add funds")
			return echo.ErrInternalServerError
		}
	default:
		s.logger.Infof("webhook event %s", event.Type)
	}

	return nil
}
