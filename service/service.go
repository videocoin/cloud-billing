package service

import (
	"github.com/stripe/stripe-go"
	accountsv1 "github.com/videocoin/cloud-api/accounts/v1"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-billing/eventbus"
	"github.com/videocoin/cloud-billing/rpc"
	"github.com/videocoin/cloud-pkg/grpcutil"
)

type Service struct {
	cfg *Config
	rpc *rpc.Server
	eb  *eventbus.EventBus
}

func NewService(cfg *Config) (*Service, error) {
	conn, err := grpcutil.Connect(cfg.AccountsRPCAddr, cfg.Logger.WithField("system", "accountscli"))
	if err != nil {
		return nil, err
	}
	accounts := accountsv1.NewAccountServiceClient(conn)

	conn, err = grpcutil.Connect(cfg.UsersRPCAddr, cfg.Logger.WithField("system", "userscli"))
	if err != nil {
		return nil, err
	}
	users := usersv1.NewUserServiceClient(conn)

	ds, err := datastore.NewDatastore(cfg.DBURI)
	if err != nil {
		return nil, err
	}

	dm, err := datastore.NewDataManager(cfg.Logger.WithField("system", "datamanager"), ds)
	if err != nil {
		return nil, err
	}

	rpcConfig := &rpc.ServerOpts{
		Logger:          cfg.Logger,
		Addr:            cfg.RPCAddr,
		AuthTokenSecret: cfg.AuthTokenSecret,
		Accounts:        accounts,
		Users:           users,
		DM:              dm,
		StripeOpts: &rpc.StripeOpts{
			BaseCallbackURL: cfg.StripeBaseCallbackURL,
		},
	}

	rpc, err := rpc.NewServer(rpcConfig)
	if err != nil {
		return nil, err
	}

	ebConfig := &eventbus.Config{
		URI:    cfg.MQURI,
		Name:   cfg.Name,
		Logger: cfg.Logger.WithField("system", "eventbus"),
		DM:     dm,
	}
	eb, err := eventbus.New(ebConfig)
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
