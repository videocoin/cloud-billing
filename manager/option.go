package manager

import (
	"github.com/sirupsen/logrus"
	emitterv1 "github.com/videocoin/cloud-api/emitter/v1"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-pkg/grpcutil"
)

type Option func(*Manager) error

func WithLogger(logger *logrus.Entry) Option {
	return func(m *Manager) error {
		m.logger = logger
		return nil
	}
}

func WithDatastore(ds *datastore.Datastore) Option {
	return func(m *Manager) error {
		m.ds = ds
		return nil
	}
}

func WithUsersServiceClient(addr string) Option {
	return func(m *Manager) error {
		conn, err := grpcutil.Connect(addr, m.logger.WithField("system", "users"))
		if err != nil {
			return err
		}
		m.users = usersv1.NewUserServiceClient(conn)
		return nil
	}
}

func WithEmitterServiceClient(addr string) Option {
	return func(m *Manager) error {
		conn, err := grpcutil.Connect(addr, m.logger.WithField("system", "emitter"))
		if err != nil {
			return err
		}
		m.emitter = emitterv1.NewEmitterServiceClient(conn)
		return nil
	}
}
