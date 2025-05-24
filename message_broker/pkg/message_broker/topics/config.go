package topics

type SubscriptionConfig struct {
	NATS     NatsSubscriber     `json:"nats"`
	RabbitMQ RabbitMQSubscriber `json:"rabbitmq"`
}

type NatsSubscriber struct {
	Subject string `json:"subject"`
	Durable string `json:"durable"`
	Stream  string `json:"stream"`
}

type RabbitMQSubscriber struct {
	QueueName   string `json:"queueName"`
	Exchange    string `json:"exchange,omitempty"`
	RoutingKey  string `json:"routingKey,omitempty"`
	Durable     bool   `json:"durable,omitempty"`     // survive broker restart
	AutoDelete  bool   `json:"autoDelete,omitempty"`  // auto-delete when unused
	Exclusive   bool   `json:"exclusive,omitempty"`   // exclusive to connection
	NoWait      bool   `json:"noWait,omitempty"`      // no wait for queue declare
	ConsumerTag string `json:"consumerTag,omitempty"` // consumer identifier
	Priority    int    `json:"priority,omitempty"`    // consumer priority (0-255)
}

type PublisherConfig struct {
	NATS     NatsPublisher     `json:"nats"`
	RabbitMQ RabbitMQPublisher `json:"rabbitmq"`
}

type NatsPublisher struct {
	Subject string `json:"subject"`
	Stream  string `json:"stream"`
	Payload []byte `json:"payload"`
}

type RabbitMQPublisher struct {
	Exchange        string                 `json:"exchange"`
	RoutingKey      string                 `json:"routingKey"`
	Payload         []byte                 `json:"payload"`
	Mandatory       bool                   `json:"mandatory,omitempty"`       // return if no route
	Immediate       bool                   `json:"immediate,omitempty"`       // immediate delivery
	DeliveryMode    uint8                  `json:"deliveryMode,omitempty"`    // 1=non-persistent, 2=persistent
	Priority        uint8                  `json:"priority,omitempty"`        // 0-9
	ContentType     string                 `json:"contentType,omitempty"`     // e.g. "application/json"
	ContentEncoding string                 `json:"contentEncoding,omitempty"` // e.g. "utf-8"
	CorrelationId   string                 `json:"correlationId,omitempty"`   // for RPC correlation
	ReplyTo         string                 `json:"replyTo,omitempty"`         // reply queue name
	Expiration      string                 `json:"expiration,omitempty"`      // message TTL in ms
	MessageId       string                 `json:"messageId,omitempty"`
	Timestamp       int64                  `json:"timestamp,omitempty"` // unix timestamp
	Type            string                 `json:"type,omitempty"`      // message type
	UserId          string                 `json:"userId,omitempty"`    // authenticated user id
	AppId           string                 `json:"appId,omitempty"`     // app id
	Headers         map[string]interface{} `json:"headers,omitempty"`   // custom headers
}

type NATSStreamSubject struct {
	Subject string
	Stream  string
}

var streamSubject []NATSStreamSubject = []NATSStreamSubject{
	{
		Subject: HealthSubject,
		Stream:  HealthStream,
	},
}

func GetStreamSubject() []NATSStreamSubject {
	return streamSubject
}

// NATS constants
const (
	HealthTopic   = "health-topic"
	HealthDurable = "health-durable"
	HealthStream  = "health-stream"
	HealthSubject = "health-subject"
)

// RabbitMQ constants
const (
	DefaultExchange   = "default-exchange"
	DefaultRoutingKey = "default-routing-key"
)
