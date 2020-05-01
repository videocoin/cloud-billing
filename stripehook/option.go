package stripehook

import (
	"github.com/sirupsen/logrus"
	emitterv1 "github.com/videocoin/cloud-api/emitter/v1"
	"github.com/videocoin/cloud-billing/manager"
	"github.com/videocoin/cloud-pkg/grpcutil"
)

type Option func(*Server) error

func WithLogger(logger *logrus.Entry) Option {
	return func(s *Server) error {
		s.logger = logger
		return nil
	}
}

func WithSecret(secret string) Option {
	return func(s *Server) error {
		s.secret = secret
		return nil
	}
}

func WithDataManager(dm *manager.Manager) Option {
	return func(s *Server) error {
		s.dm = dm
		return nil
	}
}

func WithEmitterServiceClient(addr string) Option {
	return func(s *Server) error {
		conn, err := grpcutil.Connect(addr, s.logger.WithField("system", "emitter"))
		if err != nil {
			return err
		}
		s.emitter = emitterv1.NewEmitterServiceClient(conn)
		return nil
	}
}
