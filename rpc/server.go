package rpc

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"
	accountsv1 "github.com/videocoin/cloud-api/accounts/v1"
	v1 "github.com/videocoin/cloud-api/billing/v1"
	"github.com/videocoin/cloud-api/rpc"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-pkg/auth"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type ServerOpts struct {
	Logger          *logrus.Entry
	Addr            string
	AuthTokenSecret string
	StripeOpts      *StripeOpts
	DM              *datastore.DataManager
	Users           usersv1.UserServiceClient
	Accounts        accountsv1.AccountServiceClient
}

type Server struct {
	logger          *logrus.Entry
	addr            string
	authTokenSecret string
	grpc            *grpc.Server
	listen          net.Listener
	v               *requestValidator
	dm              *datastore.DataManager
	accounts        accountsv1.AccountServiceClient
	users           usersv1.UserServiceClient
	stripeOpts      *StripeOpts
}

func NewServer(opts *ServerOpts) (*Server, error) {
	grpcOpts := grpcutil.DefaultServerOpts(opts.Logger)
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
		addr:            opts.Addr,
		authTokenSecret: opts.AuthTokenSecret,
		logger:          opts.Logger.WithField("system", "rpc"),
		v:               validator,
		grpc:            grpcServer,
		listen:          listen,
		accounts:        opts.Accounts,
		users:           opts.Users,
		dm:              opts.DM,
		stripeOpts:      opts.StripeOpts,
	}

	v1.RegisterBillingServiceServer(grpcServer, rpcServer)
	reflection.Register(grpcServer)

	return rpcServer, nil
}

func (s *Server) Start() error {
	s.logger.Infof("starting rpc server on %s", s.addr)
	return s.grpc.Serve(s.listen)
}

func (s *Server) authenticate(ctx context.Context) (string, error) {
	ctx = auth.NewContextWithSecretKey(ctx, s.authTokenSecret)
	ctx, jwtToken, err := auth.AuthFromContext(ctx)
	if err != nil {
		s.logger.Warningf("failed to auth from context: %s", err)
		return "", rpc.ErrRpcUnauthenticated
	}

	tokenType, ok := auth.TypeFromContext(ctx)
	if ok {
		if usersv1.TokenType(tokenType) == usersv1.TokenTypeAPI {
			_, err := s.users.GetApiToken(context.Background(), &usersv1.ApiTokenRequest{Token: jwtToken})
			if err != nil {
				s.logger.Errorf("failed to get api token: %s", err)
				return "", rpc.ErrRpcUnauthenticated
			}
		}
	}

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		s.logger.Warningf("failed to get user id from context: %s", err)
		return "", rpc.ErrRpcUnauthenticated
	}

	return userID, nil
}
