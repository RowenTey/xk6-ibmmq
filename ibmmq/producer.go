package ibmmq

import (
	"encoding/json"
	"errors"

	"github.com/grafana/sobek"
	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"go.k6.io/k6/js/common"
)

var (
	ErrNotEnoughArguments = errors.New("not enough arguments")
)

type ProducerConfig struct {
	QMName        string `json:"qmName"`        // Queue Manager name
	Hostname      string `json:"hostname"`      // MQ host address
	PortNumber    int    `json:"portNumber"`    // Listener port
	ChannelName   string `json:"channelName"`   // SVRCONN channel
	UserName      string `json:"userName"`      // MQ user ID
	Password      string `json:"password"`      // MQ password
	TLSEnabled    bool   `json:"tlsEnabled"`    // Use bool for clarity
	TLSCipherSpec string `json:"tlsCipherSpec"` // TLS cipher spec
	KeyRepoPath   string `json:"keyRepoPath"`   // Path to .kdb (without extension)
	CertLabel     string `json:"certLabel"`     // Label in .kdb
}

func (p *Ibmmq) ProducerJs(call sobek.ConstructorCall) *sobek.Object {
	runtime := p.vu.Runtime()

	if len(call.Arguments) != 1 {
		common.Throw(runtime, ErrNotEnoughArguments)
	}

	var producerConfig *ProducerConfig
	if params, ok := call.Argument(0).Export().(map[string]any); ok {
		b, err := json.Marshal(params)
		if err != nil {
			common.Throw(runtime, err)
		}

		if err = json.Unmarshal(b, &producerConfig); err != nil {
			common.Throw(runtime, err)
		}
	}

	ctx, producer := p.producer(producerConfig)

	// Set producer to 'This' in JS
	producerObject := runtime.NewObject()
	if err := producerObject.Set("This", producer); err != nil {
		common.Throw(runtime, err)
	}

	// Bind to JS methods
	if err := producerObject.Set("send", func(call sobek.FunctionCall) sobek.Value {
		if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
			common.Throw(runtime, ErrNotEnoughArguments)
		}

		queueName, ok := call.Argument(0).Export().(string)
		if !ok {
			common.Throw(runtime, errors.New("Missing queue name"))
		}

		jsonStrMsg, ok := call.Argument(1).Export().(string)
		if !ok {
			common.Throw(runtime, errors.New("Missing json string message"))
		}

		// Optional headers
		var headers map[string]string
		if len(call.Arguments) == 3 {
			if param, ok := call.Argument(2).Export().(map[string]any); ok {
				b, err := json.Marshal(param)
				if err != nil {
					common.Throw(runtime, errors.New("Headers must be a map[string]string"))
				}

				if err = json.Unmarshal(b, &headers); err != nil {
					common.Throw(runtime, errors.New("Headers must be a map[string]string"))
				}
			}
		}

		if err := p.send(ctx, producer, queueName, jsonStrMsg, headers); err != nil {
			common.Throw(runtime, err)
		}

		return sobek.Undefined()
	}); err != nil {
		common.Throw(runtime, err)
	}

	if err := producerObject.Set("commit", func(call sobek.FunctionCall) sobek.Value {
		if err := p.commit(ctx); err != nil {
			common.Throw(runtime, err)
		}

		return sobek.Undefined()
	}); err != nil {
		common.Throw(runtime, err)
	}

	if err := producerObject.Set("close", func(call sobek.FunctionCall) sobek.Value {
		p.close(ctx)
		return sobek.Undefined()
	}); err != nil {
		common.Throw(runtime, err)
	}

	freeze(producerObject)

	return runtime.ToValue(producerObject).ToObject(runtime)
}

func (p *Ibmmq) producer(config *ProducerConfig) (jms20subset.JMSContext, jms20subset.JMSProducer) {
	cf := mqjms.ConnectionFactoryImpl{
		QMName:      config.QMName,
		Hostname:    config.Hostname,
		PortNumber:  config.PortNumber,
		ChannelName: config.ChannelName,
		UserName:    config.UserName,
		Password:    config.Password,
	}

	if config.TLSEnabled {
		cf.TLSClientAuth = mqjms.TLSClientAuth_REQUIRED
		cf.TLSCipherSpec = config.TLSCipherSpec
		cf.KeyRepository = config.KeyRepoPath
		cf.CertificateLabel = config.CertLabel
	}

	ctx, err := cf.CreateContext()
	if err != nil {
		common.Throw(p.vu.Runtime(), err)
		return nil, nil
	}

	producer := ctx.CreateProducer()
	return ctx, producer
}

func (p *Ibmmq) send(ctx jms20subset.JMSContext, producer jms20subset.JMSProducer, queueName, msg string, headers map[string]string) jms20subset.JMSException {
	textMsg := ctx.CreateTextMessageWithString(msg)

	for key, value := range headers {
		err := textMsg.SetStringProperty(key, &value)
		if err != nil {
			return err
		}
	}

	return producer.Send(ctx.CreateQueue(queueName), textMsg)
}
