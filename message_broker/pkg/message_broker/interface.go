package message_broker

import (
	"context"
	brokerTopics "message_broker/pkg/message_broker/topics"
)

type IMBroker interface {
	Register(handlerMap map[brokerTopics.SubscriptionConfig]func(ctx context.Context, task []byte) error) error
	RegisterCallbackTask(taskName string, callBackFunc func(args ...string) error) error
	Send(config brokerTopics.PublisherConfig) error
	Retry(err error) error
}
