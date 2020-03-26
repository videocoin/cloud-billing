package manager

import (
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-billing/datastore"
)

type Option func(*Manager) error

func WithLogger(logger *logrus.Entry) Option {
	return func(s *Manager) error {
		s.logger = logger
		return nil
	}
}

func WithDatastore(ds *datastore.Datastore) Option {
	return func(s *Manager) error {
		s.ds = ds
		return nil
	}
}
