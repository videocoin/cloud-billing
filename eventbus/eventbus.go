package eventbus

import (
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	dispatcherv1 "github.com/videocoin/cloud-api/dispatcher/v1"
	"github.com/videocoin/cloud-billing/manager"
	"github.com/videocoin/cloud-pkg/mqmux"
	tracerext "github.com/videocoin/cloud-pkg/tracer"
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
	err := e.mq.Consumer("dispatcher.events", 1, false, e.handleDispatcherEvent)
	if err != nil {
		return err
	}

	return e.mq.Run()
}

func (e *EventBus) Stop() error {
	return e.mq.Close()
}

func (e *EventBus) handleDispatcherEvent(d amqp.Delivery) error {
	var span opentracing.Span
	tracer := opentracing.GlobalTracer()
	spanCtx, err := tracer.Extract(opentracing.TextMap, mqmux.RMQHeaderCarrier(d.Headers))

	e.logger.Debugf("handling body: %+v", string(d.Body))

	if err != nil {
		span = tracer.StartSpan("eventbus.handleTaskEvent")
	} else {
		span = tracer.StartSpan("eventbus.handleTaskEvent", ext.RPCServerOption(spanCtx))
	}

	defer span.Finish()

	req := new(dispatcherv1.Event)
	err = json.Unmarshal(d.Body, req)
	if err != nil {
		tracerext.SpanLogError(span, err)
		return err
	}

	span.SetTag("task_id", req.TaskID)
	span.SetTag("event_type", req.Type.String())

	logger := e.logger.WithFields(logrus.Fields{
		"task_id":    req.TaskID,
		"stream_id":  req.StreamID,
		"event_type": req.Type.String(),
	})
	logger.Debugf("handling request %+v", req)

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	switch req.Type {
	case dispatcherv1.EventTypeTaskCompleted:
		_, err := e.dm.CreateTransactionFromEvent(ctx, req)
		if err != nil {
			e.logger.Errorf("failed to create segment transcoded transcation: %s", err)
			return nil
		}
	}
	return nil
}
