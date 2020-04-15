package prpc

import (
	"net"

	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-billing/manager"
	"github.com/videocoin/cloud-pkg/grpcutil"
	"google.golang.org/grpc"
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

func WithDataManager(dm *manager.Manager) Option {
	return func(s *Server) error {
		s.dm = dm
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
