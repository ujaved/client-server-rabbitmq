package internal

import (
	"encoding/json"
	"fmt"

	"client-server-rabbitmq/api"

	"github.com/go-playground/validator/v10"
	amqp "github.com/rabbitmq/amqp091-go"
)

var validate = validator.New()

type QueueState struct {
	connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      amqp.Queue
}

func (qs *QueueState) SafeClose() {
	if qs.Channel != nil {
		qs.Channel.Close()
	}
	if qs.connection != nil {
		qs.connection.Close()
	}
}

func (qs *QueueState) PublishRequest(req api.Request, clientId string) error {
	req.ClientId = clientId
	if req.Items != nil {
		req.Items[0].Id = req.RequestId
	}
	if err := validate.Struct(req); err != nil {
		return fmt.Errorf("failed to validate request: %w", err)
	}
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %w", err)
	}
	if err = qs.Channel.Publish("", qs.Queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}); err != nil {
		return err
	}
	return nil
}

func InitQueue(username, password, queueName string, port int) (*QueueState, error) {

	qs := &QueueState{}
	var err error
	qs.connection, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:%d/", username, password, port))
	if err != nil {
		qs.SafeClose()
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	qs.Channel, err = qs.connection.Channel()
	if err != nil {
		qs.SafeClose()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	qs.Queue, err = qs.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		qs.SafeClose()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}
	return qs, nil
}
