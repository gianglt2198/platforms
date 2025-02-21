package core

import (
	"context"
	"sync"
	"time"

	oblogger "github.com/gianglt2198/platforms/observability/logger"
	"github.com/gianglt2198/platforms/pkg/utils"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type (
	NatsConfig struct {
		Connection string `json:"connection"`
	}

	MqBroker[T any] struct {
		natsCon *nats.Conn
		logger  oblogger.ObLogger
	}
)

var (
	mqBrokerOnce sync.Once
)

func ProvideMqBroker[T any](logger oblogger.ObLogger, config *NatsConfig) *MqBroker[T] {
	var mqBroker *MqBroker[T]
	mqBrokerOnce.Do(func() {
		var err error
		var natsCon *nats.Conn
		for range 5 {
			natsCon, err = nats.Connect(config.Connection)

			if err == nil {
				break
			}
		}

		if err != nil {
			panic(err)
		}

		mqBroker = &MqBroker[T]{natsCon, logger}
	})

	return mqBroker
}

func (m *MqBroker[T]) CloseMQ() {
	if m.natsCon != nil {
		m.natsCon.Close()
	}
}

func (b *MqBroker[T]) RequestOperation(
	ctx context.Context,
	eventName string,
	payload any,
) ([]byte, error) {
	return b.requestOperation(ctx, eventName, payload, 2*time.Second)
}

func (b *MqBroker[T]) RequestOperationWithTimeout(
	ctx context.Context,
	eventName string,
	payload any,
	timeout time.Duration,
) ([]byte, error) {
	return b.requestOperation(ctx, eventName, payload, timeout)
}

func (b *MqBroker[T]) requestOperation(
	ctx context.Context,
	eventName string,
	payload any,
	timeout time.Duration,
) ([]byte, error) {

	b.logger.Info(ctx, "[MqBroker]RequestOperation: ", eventName)

	toBytes, err := utils.TransformToByteArray(payload)
	if err != nil {
		b.logger.Error(ctx, "[MqBroker]RequestOperation", err)
		return nil, err
	}

	// wait for 2 seconds
	msg, err := b.natsCon.RequestMsg(&nats.Msg{
		Subject: eventName,
		Data:    toBytes,
	}, timeout)

	if err != nil {
		b.logger.Error(ctx, "[MqBroker]RequestOperation", err)
		return nil, err
	}

	return msg.Data, nil
}

func (b *MqBroker[T]) ListenIndexOperation(ctx context.Context, event string, opFunc func(context.Context, []byte) error) {
	_, err := b.natsCon.QueueSubscribe(event, "query-worker", func(m *nats.Msg) {
		correlationId := m.Header.Get("correlation-id")
		b.logger.Info(ctx, "[MqBroker]SubscribeEvent", event, correlationId)
		err := opFunc(ctx, m.Data)
		if err != nil {
			b.logger.Error(ctx, "[MqBroker]SubscribeEvent: fail to execute subcribe logic", err)
			return
		}
	})
	if err != nil {
		b.logger.Error(ctx, "[MqBroker]SubscribeEvent: fail to subscribe event", err)
	}
}

func (b *MqBroker[T]) SubscribeOperation() func(context.Context, string, func(ctx context.Context, payload []byte) ([]byte, error)) {
	return func(ctx context.Context, eventName string, opFunc func(ctx context.Context, payload []byte) ([]byte, error)) {
		_, err := b.natsCon.QueueSubscribe(eventName, "gw-worker", func(m *nats.Msg) {
			b.logger.Info(ctx, "[MqBroker]SubscribeOperation", eventName)

			ReplyFunc := func(replyPayload []byte) {
				err := m.Respond(replyPayload)
				if err != nil {
					b.logger.Error(ctx, "[MqBroker]SubscribeOperation", err)
				}
			}

			replyPayload, err := opFunc(ctx, m.Data)

			if err != nil {
				b.logger.Error(ctx, "[MqBroker]SubscribeOperation", err)
			} else {
				ReplyFunc(replyPayload)
			}
		})

		if err != nil {
			b.logger.Error(ctx, "[MqBroker]SubscribeOperation", err)
		}
	}
}

func (b *MqBroker[T]) PublishEvent(ctx context.Context, eventName string, payload interface{}) error {
	correlationId := uuid.NewString()

	b.logger.Info(ctx, "[MqBroker]PublishEvent: ", eventName, correlationId)

	headers := nats.Header{}
	headers.Set("correlation-id", correlationId)

	sendBytes, err := utils.TransformToByteArray(payload)
	if err != nil {
		b.logger.Error(ctx, "[MqBroker]PublishEvent: fail to prepare data to publish", err)
		return err
	}
	if err = b.natsCon.PublishMsg(&nats.Msg{
		Subject: eventName,
		Header:  headers,
		Data:    sendBytes,
	}); err != nil {
		b.logger.Error(ctx, "[MqBroker]PublishEvent: fail to publish event", err)
		return err
	}

	return nil
}

func (b *MqBroker[T]) SubscribeEvent() func(context.Context, string, func(ctx context.Context, payload []byte) error) {
	return func(ctx context.Context, eventName string, opFunc func(ctx context.Context, payload []byte) error) {
		_, err := b.natsCon.QueueSubscribe(eventName, "gw-worker", func(m *nats.Msg) {
			correlationId := m.Header.Get("correlation-id")
			b.logger.Info(ctx, "[MqBroker]SubscribeEvent", eventName, correlationId)

			err := opFunc(ctx, m.Data)

			if err != nil {
				b.logger.Error(ctx, "[MqBroker]SubscribeEvent: fail to execute subcribe logic", err)
				return
			}
		})

		if err != nil {
			b.logger.Error(ctx, "[MqBroker]SubscribeEvent", err)
		}
	}
}
