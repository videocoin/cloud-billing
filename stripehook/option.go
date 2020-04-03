package stripehook

import (
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-billing/manager"
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
