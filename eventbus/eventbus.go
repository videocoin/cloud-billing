package eventbus

import (
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-billing/manager"
	"github.com/videocoin/cloud-pkg/mqmux"
)

type EventBus struct {
	logger *logrus.Entry
	uri    string
	name   string
	mq     *mqmux.WorkerMux
	dm     *manager.Manager
}

func NewEventBus(uri string, opts ...Option) (*EventBus, error) {
	eb := &EventBus{
		uri: uri,
	}
	for _, o := range opts {
		if err := o(eb); err != nil {
			return nil, err
		}
	}

	mq, err := mqmux.NewWorkerMux(eb.uri, eb.name)
	if err != nil {
		return nil, err
	}

	eb.mq = mq

	return eb, nil
}

func (e *EventBus) Start() error {
	return e.mq.Run()
}

func (e *EventBus) Stop() error {
	return e.mq.Close()
}
