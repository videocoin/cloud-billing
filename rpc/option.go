package rpc

import (
	"net"

	"github.com/sirupsen/logrus"
	accountsv1 "github.com/videocoin/cloud-api/accounts/v1"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-billing/manager"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

type Option func(*Server) error

func WithAddr(addr string) Option {
	return func(s *Server) error {
		s.addr = addr

		listen, err := net.Listen("tcp", s.addr)
		if err != nil {
			return err
		}

		s.listen = listen

		return nil
	}
}

func WithLogger(logger *logrus.Entry) Option {
	return func(s *Server) error {
		s.logger = logger
		return nil
	}
}

func WithAuthTokenSecret(secret string) Option {
	return func(s *Server) error {
		s.authTokenSecret = secret
		return nil
	}
}

func WithDataManager(dm *manager.Manager) Option {
	return func(s *Server) error {
		s.dm = dm
		return nil
	}
}

func WithStripeOpts(opts *StripeOpts) Option {
	return func(s *Server) error {
		s.stripeOpts = opts
		return nil
	}
}

func WithUsersServiceClient(addr string) Option {
	return func(s *Server) error {
		conn, err := grpcutil.Connect(addr, s.logger.WithField("system", "users"))
		if err != nil {
			return err
		}
		s.users = usersv1.NewUserServiceClient(conn)
		return nil
	}
}

func WithAccountsServiceClient(addr string) Option {
	return func(s *Server) error {
		conn, err := grpcutil.Connect(addr, s.logger.WithField("system", "accounts"))
		if err != nil {
			return err
		}
		s.accounts = accountsv1.NewAccountServiceClient(conn)
		return nil
	}
}

func WithGRPCDefaultOpts() Option {
	return func(s *Server) error {
		opts := grpcutil.DefaultServerOpts(s.logger)
		s.grpc = grpc.NewServer(opts...)
		return nil
	}
}

func WithHealthService() Option {
	return func(s *Server) error {
		healthService := health.NewServer()
		healthv1.RegisterHealthServer(s.grpc, healthService)
		return nil
	}
}

func WithValidator() Option {
	return func(s *Server) error {
		validator, err := newRequestValidator()
		if err != nil {
			return err
		}
		s.validator = validator
		return nil
	}
}
