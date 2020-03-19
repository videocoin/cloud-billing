package service

import (
	"github.com/stripe/stripe-go"
	accountsv1 "github.com/videocoin/cloud-api/accounts/v1"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-billing/eventbus"
	"github.com/videocoin/cloud-billing/rpc"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
)

type Service struct {
	cfg *Config
	rpc *rpc.Server
	eb  *eventbus.EventBus
}

func NewService(cfg *Config) (*Service, error) {
	alogger := cfg.Logger.WithField("system", "accountcli")
	aGrpcDialOpts := grpcutil.ClientDialOptsWithRetry(alogger)
	accountsConn, err := grpc.Dial(cfg.AccountsRPCAddr, aGrpcDialOpts...)
	if err != nil {
		return nil, err
	}
	accounts := accountsv1.NewAccountServiceClient(accountsConn)

	ds, err := datastore.NewDatastore(cfg.DBURI)
	if err != nil {
		return nil, err
	}

	dm, err := datastore.NewDataManager(cfg.Logger.WithField("system", "datamanager"), ds)
	if err != nil {
		return nil, err
	}

	rpcConfig := &rpc.ServerOpts{
		Addr:     cfg.RPCAddr,
		Logger:   cfg.Logger,
		Accounts: accounts,
		DM:       dm,
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

func (s *Service) Start() error {
	s.cfg.Logger.Info("starting rpc server")
	go s.rpc.Start() //nolint

	s.cfg.Logger.Info("starting eventbus")
	go s.eb.Start() //nolint

	return nil
}

func (s *Service) Stop() error {
	err := s.eb.Stop()
	if err != nil {
		return err
	}

	return nil
}
