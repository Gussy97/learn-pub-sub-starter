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
	handler func(T) Acktype,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("error declaring and binding: %v", err)
	}

	delCh, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error consuming channel: %v", err)
	}

	unmarsheller := func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}
	go func() {
		defer ch.Close()
		for del := range delCh {
			target, err := unmarsheller(del.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			switch handler(target) {
			case Ack:
				del.Ack(false)
				fmt.Println("Ack")
			case NackDiscard:
				del.Nack(false, false)
				fmt.Println("NackDiscard")
			case NackRequeue:
				del.Nack(false, true)
				fmt.Println("NackRequeue")
			}
		}
	}()

	return nil
}
