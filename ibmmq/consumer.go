package ibmmq

import (
	"encoding/json"
	"errors"

	"github.com/grafana/sobek"
	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"go.k6.io/k6/js/common"
)

type ConsumerConfig struct {
	QMName      string `json:"qmName"`      // Queue Manager name
	Hostname    string `json:"hostname"`    // MQ host address
	PortNumber  int    `json:"portNumber"`  // Listener port
	ChannelName string `json:"channelName"` // SVRCONN channel
	QueueName   string `json:"queueName"`   // Queue name
	UserName    string `json:"userName"`    // MQ user ID
	Password    string `json:"password"`    // MQ password
	Timeout     int32  `json:"timeout"`     // Consumer receive timeout
	MsgLimit    int    `json:"msgLimit"`    // Number of messages to consume
}

func (c *Ibmmq) ConsumerJs(call sobek.ConstructorCall) *sobek.Object {
	runtime := c.vu.Runtime()

	if len(call.Arguments) != 1 {
		common.Throw(runtime, ErrNotEnoughArguments)
	}

	var consumerConfig *ConsumerConfig
	if params, ok := call.Argument(0).Export().(map[string]any); ok {
		b, err := json.Marshal(params)
		if err != nil {
			common.Throw(runtime, err)
		}

		if err = json.Unmarshal(b, &consumerConfig); err != nil {
			common.Throw(runtime, err)
		}
	}

	ctx, consumer := c.consumer(consumerConfig)

	// Set consumer to 'This' in JS
	consumerObject := runtime.NewObject()
	if err := consumerObject.Set("This", consumer); err != nil {
		common.Throw(runtime, err)
	}

	// Bind to JS methods
	if err := consumerObject.Set("consume", func(call sobek.FunctionCall) sobek.Value {
		messages, err := c.consume(consumerConfig, consumer)
		if err != nil {
			common.Throw(runtime, err)
		}

		return runtime.ToValue(messages)
	}); err != nil {
		common.Throw(runtime, err)
	}

	if err := consumerObject.Set("commit", func(call sobek.FunctionCall) sobek.Value {
		if err := c.commit(ctx); err != nil {
			common.Throw(runtime, err)
		}

		return sobek.Undefined()
	}); err != nil {
		common.Throw(runtime, err)
	}

	if err := consumerObject.Set("close", func(call sobek.FunctionCall) sobek.Value {
		c.close(ctx)
		return sobek.Undefined()
	}); err != nil {
		common.Throw(runtime, err)
	}

	freeze(consumerObject)

	return runtime.ToValue(consumerObject).ToObject(runtime)
}

func (c *Ibmmq) consumer(config *ConsumerConfig) (jms20subset.JMSContext, jms20subset.JMSConsumer) {
	runtime := c.vu.Runtime()

	cf := mqjms.ConnectionFactoryImpl{
		QMName:      config.QMName,
		Hostname:    config.Hostname,
		PortNumber:  config.PortNumber,
		ChannelName: config.ChannelName,
		UserName:    config.UserName,
		Password:    config.Password,
	}

	ctx, err := cf.CreateContext()
	if err != nil {
		common.Throw(runtime, err)
		return nil, nil
	}

	consumer, err := ctx.CreateConsumer(ctx.CreateQueue(config.QueueName))
	if err != nil {
		common.Throw(runtime, err)
		return nil, nil
	}

	return ctx, consumer
}

func (c *Ibmmq) consume(config *ConsumerConfig, consumer jms20subset.JMSConsumer) ([]map[string]any, error) {
	if config.MsgLimit <= 0 {
		config.MsgLimit = 1
	}

	messages := make([]map[string]any, 0)
	for i := 0; i < config.MsgLimit; i++ {
		payload, err := consumer.Receive(config.Timeout)
		if err != nil {
			return nil, err
		}

		var body string
		switch msg := payload.(type) {
		case jms20subset.TextMessage:
			body = *msg.GetText()
		default:
			return nil, errors.New("received non-text message")
		}

		properties, err := payload.GetPropertyNames()
		if err != nil {
			return nil, err
		}

		headers := make(map[string]string)
		for _, header := range properties {
			value, err := payload.GetStringProperty(header)
			if err != nil {
				return nil, err
			}
			headers[header] = *value
		}

		msg := map[string]any{
			"headers": headers,
			"body":    body,
		}

		messages = append(messages, msg)
	}

	return messages, nil
}
