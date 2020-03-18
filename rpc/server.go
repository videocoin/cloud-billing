package rpc

import (
	"net"

	"github.com/sirupsen/logrus"
	accountsv1 "github.com/videocoin/cloud-api/accounts/v1"
	v1 "github.com/videocoin/cloud-api/billing/v1"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerOpts struct {
	Logger   *logrus.Entry
	Addr     string
	DM       *datastore.DataManager
	Accounts accountsv1.AccountServiceClient
}

type Server struct {
	logger   *logrus.Entry
	addr     string
	grpc     *grpc.Server
	listen   net.Listener
	v        *requestValidator
	dm       *datastore.DataManager
	accounts accountsv1.AccountServiceClient
}

func NewServer(opts *ServerOpts) (*Server, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
	grpcOpts = append(grpcOpts, grpc.MaxRecvMsgSize(1024*1024*1024))
	grpcOpts = append(grpcOpts, grpc.MaxSendMsgSize(1024*1024*1024))
	grpcServer := grpc.NewServer(grpcOpts...)

	healthService := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthService)

	listen, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, err
	}
	validator, err := newRequestValidator()
	if err != nil {
		return nil, err
	}
	rpcServer := &Server{
		addr:     opts.Addr,
		logger:   opts.Logger.WithField("system", "rpc"),
		v:        validator,
		grpc:     grpcServer,
		listen:   listen,
		accounts: opts.Accounts,
		dm:       opts.DM,
	}

	v1.RegisterBillingServiceServer(grpcServer, rpcServer)
	reflection.Register(grpcServer)

	return rpcServer, nil
}

func (s *Server) Start() error {
	s.logger.Infof("starting rpc server on %s", s.addr)
	return s.grpc.Serve(s.listen)
}
