package internal

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"go.uber.org/zap"
)

type MessageBrokerInterface interface {
	QueueSubscribe(string, stan.MsgHandler, ...stan.SubscriptionOption) (stan.Subscription, error)
	Close() error
}

type MessageBroker struct {
	client    stan.Conn
	asyncMode bool
}

func newMessageBroker(config *config.NatsConfig, cancel context.CancelFunc) (MessageBrokerInterface, error) {
	opts := []nats.Option{
		nats.Name("NATS Streaming Publisher"),
	}

	mb := &MessageBroker{asyncMode: config.Async}

	if config.User != "" && config.Password != "" {
		opts = append(opts, nats.UserInfo(config.User, config.Password))
	}

	nc, err := nats.Connect(config.ServerUrls, opts...)
	if err != nil {
		return nil, err
	}

	mb.client, err = stan.Connect(
		config.ClusterId,
		config.ClientId,
		stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			zap.L().Error("connect to NATS Streaming server lost", zap.Error(err))
			cancel()
		}),
	)
	if err != nil {
		return nil, err
	}

	return mb, nil
}

func (c MessageBroker) QueueSubscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return c.client.QueueSubscribe(subject, "", cb, opts...)
}

func (c MessageBroker) Close() error {
	return c.client.Close()
}
