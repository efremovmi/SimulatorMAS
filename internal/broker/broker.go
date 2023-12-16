package broker

import (
	"fmt"
	"go.uber.org/zap"
	"sync"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	. "SimulatorMAS/internal/config_parser"
	. "SimulatorMAS/internal/logger"
)

var Broker *broker

func init() {
	var err error
	Broker, err = newBroker(
		Cfg.Broker.BrokerAddress,
		Cfg.Broker.QueueCorpsesName,
		Cfg.Broker.QueueGreensReqToBornName,
		Cfg.Broker.QueueRodentReqToBornName,
		Cfg.Broker.RetryCount,
	)
	if err != nil {
		Log.Fatal("Can't create broker", zap.Error(err))
	}
}

type broker struct {
	mu              sync.RWMutex
	retryCount      int
	connectRabbitMQ *amqp.Connection
	channelRabbitMQ *amqp.Channel
	addr            string

	corpsesChan      <-chan amqp.Delivery
	queueCorpsesName string

	greensChan          <-chan amqp.Delivery
	greensReqToBornChan <-chan amqp.Delivery
	queueGreensName     string

	rodentReqToBornChan <-chan amqp.Delivery
	rodentGreensName    string
}

func newBroker(addr,
	queueCorpsesName,
	QueueGreensReqToBornName,
	QueueRodentReqToBornName string,
	retryCount int) (*broker, error) {
	// Create a new RabbitMQ connection.
	connectRabbitMQ, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		return nil, err
	}

	//  create queues
	{
		err = CreateQueues([]string{
			queueCorpsesName,
			QueueGreensReqToBornName,
			QueueRodentReqToBornName,
		}, channelRabbitMQ)
		if err != nil {
			return nil, err
		}
	}

	// get chans
	corpsesChan, err := GetSubscribeByQueue(queueCorpsesName, channelRabbitMQ)
	if err != nil {
		return nil, err
	}

	greensReqToBornChan, err := GetSubscribeByQueue(QueueGreensReqToBornName, channelRabbitMQ)
	if err != nil {
		return nil, err
	}

	rodentReqToBornChan, err := GetSubscribeByQueue(QueueRodentReqToBornName, channelRabbitMQ)
	if err != nil {
		return nil, err
	}

	b := &broker{
		connectRabbitMQ: connectRabbitMQ,
		channelRabbitMQ: channelRabbitMQ,
		addr:            addr,
		retryCount:      retryCount,
		// corpses
		corpsesChan:      corpsesChan,
		queueCorpsesName: queueCorpsesName,
		// greens
		greensReqToBornChan: greensReqToBornChan,
		// rodent
		rodentReqToBornChan: rodentReqToBornChan,
	}

	return b, nil
}

func CreateQueues(queueNames []string, channel *amqp.Channel) error {
	for _, queueName := range queueNames {
		// With the instance and declare Queues that we can publish and subscribe to.
		_, err := channel.QueueDeclare(
			queueName, // queue name
			true,      // durable
			false,     // auto delete
			false,     // exclusive
			false,     // no wait
			nil,       // arguments
		)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Can't create queue with name %s", queueName))
		}
	}

	return nil
}

func GetSubscribeByQueue(queueName string, channel *amqp.Channel) (<-chan amqp.Delivery, error) {
	// Subscribing to QueueService1 for getting messages.
	corpsesChan, err := channel.Consume(
		queueName, // broker name
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no local
		false,     // no wait
		nil,       // arguments
	)
	if err != nil {
		return nil, errors.Wrap(err, "Can't subscribe to queue")
	}

	return corpsesChan, nil
}

func (b *broker) Close() {
	_ = b.connectRabbitMQ.Close()
	_ = b.channelRabbitMQ.Close()
}

func (b *broker) PutMessageToQueue(queueName string, msg string) error {
	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	}

	// Attempt to publish a message to the broker.
	if err := b.channelRabbitMQ.Publish(
		"",        // exchange
		queueName, // broker name
		false,     // mandatory
		false,     // immediate
		message,   // message to publish
	); err == nil {
		return nil
	}

	if err := b.retry(); err != nil {
		return errors.Wrap(err, "Can't put corpse to broker")
	}

	return nil
}

func (b *broker) PutMessageToQueueNTimes(queueName string, msg string, n int) error {

	for i := 0; i < n; i++ {
		message := amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg + fmt.Sprintf("-%d", i)),
		}

		// Attempt to publish a message to the broker.
		if err := b.channelRabbitMQ.Publish(
			"",        // exchange
			queueName, // broker name
			false,     // mandatory
			false,     // immediate
			message,   // message to publish
		); err == nil {
			continue
		}
		//f()

		if err := b.retry(); err != nil {
			return errors.Wrap(err, "Can't put corpse to broker")
		}
	}

	return nil
}

func (b *broker) GetMessageFromQueue(queueName string) (string, error) {
	switch queueName {
	case b.queueCorpsesName:
		select {
		case msg, ok := <-b.corpsesChan:
			if ok {
				BrokerInfo(Cfg.Broker.Logging, "Broker: get message, body: %s", msg.Body)
				return string(msg.Body), nil
			}
			return "", errors.New("Broker: Error get message from broker")
		default:
			return "", errors.New("Broker: queue is empty")
		}
	}

	return "", nil
}

func (b *broker) GetGreensReqToBornChan() chan struct{} {
	getGreensReqToBornChan := make(chan struct{})
	go func() {
		for msg := range b.greensReqToBornChan {
			BrokerInfo(Cfg.Broker.Logging, "Broker: get message, body: %s", msg.Body)
			getGreensReqToBornChan <- struct{}{}
		}
	}()
	return getGreensReqToBornChan
}

func (b *broker) GetRodentReqToBornChan() chan struct{} {
	getRodentReqToBornChan := make(chan struct{})
	go func() {
		for msg := range b.rodentReqToBornChan {
			BrokerInfo(Cfg.Broker.Logging, "Broker: get message, body: %s", msg.Body)
			getRodentReqToBornChan <- struct{}{}
		}
	}()
	return getRodentReqToBornChan
}

func (b *broker) retry() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var connectRabbitMQ *amqp.Connection
	var channelRabbitMQ *amqp.Channel
	var err error
	for i := 0; i < b.retryCount; i++ {
		connectRabbitMQ, err = amqp.Dial(b.addr)
		if err != nil {
			continue
		}

		channelRabbitMQ, err = connectRabbitMQ.Channel()
		if err != nil {
			continue
		}
	}

	if err != nil {
		return errors.Wrap(err, "Can't get connection to broker")
	}

	b.connectRabbitMQ = connectRabbitMQ
	b.channelRabbitMQ = channelRabbitMQ

	return nil
}
