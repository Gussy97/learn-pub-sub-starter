package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("error declaring and binding: %v", err)
	}

	delCh, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error consuming channel: %v", err)
	}

	go func() {
		defer ch.Close()
		for del := range delCh {
			var obj T
			err = json.Unmarshal(del.Body, &obj)
			if err != nil {
				fmt.Printf("error unmarshling data: %v\n", err)
			}
			handler(obj)
			err = del.Ack(false)
			if err != nil {
				fmt.Printf("error acknowledging message: %v\n", err)
			}
		}
	}()

	return nil
}
