package service

import (
	"github.com/stripe/stripe-go"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-billing/eventbus"
	"github.com/videocoin/cloud-billing/manager"
	"github.com/videocoin/cloud-billing/prpc"
	"github.com/videocoin/cloud-billing/rpc"
	"github.com/videocoin/cloud-billing/stripehook"
)

type Service struct {
	cfg  *Config
	rpc  *rpc.Server
	prpc *prpc.Server
	eb   *eventbus.EventBus
	dm   *manager.Manager
	shs  *stripehook.Server
}

func NewService(cfg *Config) (*Service, error) {
	ds, err := datastore.NewDatastore(cfg.DBURI)
	if err != nil {
		return nil, err
	}

	dm, err := manager.New(
		manager.WithLogger(cfg.Logger.WithField("system", "datamanager")),
		manager.WithDatastore(ds),
		manager.WithUsersServiceClient(cfg.UsersRPCAddr),
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

	prpc, err := prpc.NewServer(
		prpc.WithAddr(cfg.PRPCAddr),
		prpc.WithLogger(cfg.Logger.WithField("system", "prpc")),
		prpc.WithGRPCDefaultOpts(),
		prpc.WithDataManager(dm),
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

	shs, err := stripehook.NewServer(
		cfg.StripeHookServerAddr,
		stripehook.WithLogger(cfg.Logger.WithField("system", "stripehook")),
		stripehook.WithSecret(cfg.StripeWHSecret),
		stripehook.WithDataManager(dm),
	)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		cfg:  cfg,
		rpc:  rpc,
		prpc: prpc,
		eb:   eb,
		dm:   dm,
		shs:  shs,
	}

	return svc, nil
}

func (s *Service) Start(errCh chan error) {
	go func() {
		s.cfg.Logger.Info("starting rpc server")
		errCh <- s.rpc.Start()
	}()

	go func() {
		s.cfg.Logger.Info("starting private rpc server")
		errCh <- s.prpc.Start()
	}()

	go func() {
		s.cfg.Logger.Info("starting stripe hook server")
		errCh <- s.shs.Start()
	}()

	go func() {
		s.cfg.Logger.Info("starting eventbus")
		errCh <- s.eb.Start()
	}()

	go func() {
		s.cfg.Logger.Info("starting manager")
		s.dm.Start()
	}()
}

func (s *Service) Stop() error {
	s.dm.Stop()

	err := s.eb.Stop()
	if err != nil {
		return err
	}

	err = s.shs.Stop()
	if err != nil {
		return err
	}

	return nil
}
