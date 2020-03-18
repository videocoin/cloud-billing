package eventbus

import (
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-billing/datastore"
	"github.com/videocoin/cloud-pkg/mqmux"
)

type Config struct {
	Logger *logrus.Entry
	URI    string
	Name   string
	DM     *datastore.DataManager
}

type EventBus struct {
	logger *logrus.Entry
	mq     *mqmux.WorkerMux
	dm     *datastore.DataManager
}

func New(c *Config) (*EventBus, error) {
	mq, err := mqmux.NewWorkerMux(c.URI, c.Name)
	if err != nil {
		return nil, err
	}
	if c.Logger != nil {
		mq.Logger = c.Logger
	}
	return &EventBus{
		logger: c.Logger,
		mq:     mq,
		dm:     c.DM,
	}, nil
}

func (e *EventBus) Start() error {
	return e.mq.Run()
}

func (e *EventBus) Stop() error {
	return e.mq.Close()
}
