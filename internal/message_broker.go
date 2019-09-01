package internal

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"go.uber.org/zap"
	"sync"
	"time"
)

type MessageBrokerInterface interface {
	Publish(string, interface{}, bool) error
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

func (c MessageBroker) Publish(subject string, msg interface{}, async bool) error {
	var (
		glock sync.Mutex
		guid  string
		ch    = make(chan bool)
	)

	message, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	acb := func(lguid string, err error) {
		glock.Lock()
		defer glock.Unlock()

		if err != nil {
			zap.L().Fatal(
				"Error in server ack for guid in the message broker",
				zap.Error(err),
				zap.Any("lguid", lguid),
			)
		}

		if lguid != guid {
			zap.L().Fatal(
				"Expected a matching guid in ack callback in the message broker",
				zap.Any("guid", guid),
				zap.Any("lguid", lguid),
			)
		}
		ch <- true
	}

	if !async {
		if err = c.client.Publish(subject, message); err != nil {
			return err
		}
	} else {
		glock.Lock()

		if guid, err = c.client.PublishAsync(subject, message, acb); err != nil {
			return err
		}

		glock.Unlock()

		if guid == "" {
			zap.L().Fatal("Expected non-empty guid to be returned from the message broker")
		}

		select {
		case <-ch:
			break
		case <-time.After(5 * time.Second):
			zap.L().Fatal("Timeout to publish message to the message broker")
		}
	}

	return nil
}

func (c MessageBroker) QueueSubscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return c.client.QueueSubscribe(subject, "", cb, opts...)
}

func (c MessageBroker) Close() error {
	return c.client.Close()
}
