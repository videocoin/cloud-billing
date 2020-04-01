package service

import (
	"github.com/stripe/stripe-go"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-billing/eventbus"
	"github.com/videocoin/cloud-billing/manager"
	"github.com/videocoin/cloud-billing/rpc"
)

type Service struct {
	cfg *Config
	rpc *rpc.Server
	eb  *eventbus.EventBus
}

func NewService(cfg *Config) (*Service, error) {
	ds, err := datastore.NewDatastore(cfg.DBURI)
	if err != nil {
		return nil, err
	}

	dm, err := manager.New(
		manager.WithLogger(cfg.Logger.WithField("system", "datamanager")),
		manager.WithDatastore(ds),
	)
	if err != nil {
		return nil, err
	}

	rpc, err := rpc.NewServer(
		rpc.WithAddr(cfg.RPCAddr),
		rpc.WithLogger(cfg.Logger.WithField("system", "rpc")),
		rpc.WithGRPCDefaultOpts(),
		rpc.WithHealthService(),
		rpc.WithAuthTokenSecret(cfg.AuthTokenSecret),
		rpc.WithDataManager(dm),
		rpc.WithUsersServiceClient(cfg.UsersRPCAddr),
	)
	if err != nil {
		return nil, err
	}

	eb, err := eventbus.NewEventBus(
		cfg.MQURI,
		eventbus.WithName(cfg.Name),
		eventbus.WithLogger(cfg.Logger.WithField("system", "eventbus")),
		eventbus.WithDataManager(dm),
	)
	if err != nil {
		return nil, err
	}

	stripe.Key = cfg.StripeKey

	svc := &Service{
		cfg: cfg,
		rpc: rpc,
		eb:  eb,
	}

	return svc, nil
}

func (s *Service) Start(errCh chan error) {
	go func() {
		s.cfg.Logger.Info("starting rpc server")
		errCh <- s.rpc.Start()
	}()

	go func() {
		s.cfg.Logger.Info("starting eventbus")
		errCh <- s.eb.Start()
	}()
}

func (s *Service) Stop() error {
	err := s.eb.Stop()
	if err != nil {
		return err
	}

	return nil
}
