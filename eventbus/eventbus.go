package eventbus

import (
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	dispatcherv1 "github.com/videocoin/cloud-api/dispatcher/v1"
	validatorv1 "github.com/videocoin/cloud-api/validator/v1"
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

	err = e.mq.Consumer("validator.events", 1, false, e.handleValidatorEvent)
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
		span = tracer.StartSpan("eventbus.handleDispatcherEvent")
	} else {
		span = tracer.StartSpan("eventbus.handleDispatcherEvent", ext.RPCServerOption(spanCtx))
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
	case dispatcherv1.EventTypeSegementTranscoded:
		_, err := e.dm.CreateTransactionFromDispatcherEvent(ctx, req)
		if err != nil {
			logger.Errorf("failed to create transcation from dispatcher event: %s", err)
			return nil
		}
	}
	return nil
}

func (e *EventBus) handleValidatorEvent(d amqp.Delivery) error {
	var span opentracing.Span
	tracer := opentracing.GlobalTracer()
	spanCtx, err := tracer.Extract(opentracing.TextMap, mqmux.RMQHeaderCarrier(d.Headers))

	e.logger.Debugf("handling body: %+v", string(d.Body))

	if err != nil {
		span = tracer.StartSpan("eventbus.handleValidatorEvent")
	} else {
		span = tracer.StartSpan("eventbus.handleValidatorEvent", ext.RPCServerOption(spanCtx))
	}

	defer span.Finish()

	req := new(validatorv1.Event)
	err = json.Unmarshal(d.Body, req)
	if err != nil {
		tracerext.SpanLogError(span, err)
		return err
	}

	span.SetTag("event_type", req.Type.String())

	logger := e.logger.WithFields(logrus.Fields{
		"sca":        req.StreamContractAddress,
		"chunk_num":  req.ChunkNum,
		"event_type": req.Type.String(),
	})
	logger.Debugf("handling request %+v", req)

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	switch req.Type {
	case validatorv1.EventTypeValidatedProof:
		err := e.dm.MarkTransactionAsSuccededByValidatorEvent(ctx, req)
		if err != nil {
			logger.Errorf("failed to mark transaction as succeded by validator event: %s", err)
			return nil
		}
	case validatorv1.EventTypeScrapedProof:
		err := e.dm.MarkTransactionAsCanceledByValidatorEvent(ctx, req)
		if err != nil {
			logger.Errorf("failed to mark transaction as canceled by validator event: %s", err)
			return nil
		}
	}
	return nil
}
