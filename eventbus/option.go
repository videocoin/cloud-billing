package eventbus

import (
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-billing/manager"
)

type Option func(*EventBus) error

func WithLogger(logger *logrus.Entry) Option {
	return func(e *EventBus) error {
		e.logger = logger
		return nil
	}
}

func WithDataManager(dm *manager.Manager) Option {
	return func(e *EventBus) error {
		e.dm = dm
		return nil
	}
}

func WithName(name string) Option {
	return func(e *EventBus) error {
		e.name = name
		return nil
	}
}
