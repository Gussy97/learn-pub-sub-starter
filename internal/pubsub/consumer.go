package pubsub

import (
	"bytes"
	"encoding/gob"
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
			case NackDiscard:
				del.Nack(false, false)
			case NackRequeue:
				del.Nack(false, true)
			}
		}
	}()

	return nil
}

func SubscribeGob[T any](
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

	if err = ch.Qos(10, 0, false); err != nil {
		return err
	}

	delCh, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("error consuming channel: %v", err)
	}

	decode := func(data []byte) (T, error) {
		buf := bytes.NewBuffer(data)
		var target T
		decoder := gob.NewDecoder(buf)
		err := decoder.Decode(&target)
		return target, err
	}

	go func() {
		defer ch.Close()
		for msg := range delCh {
			target, err := decode(msg.Body)
			if err != nil {
				fmt.Printf("could not decode message: %v\n", err)
				continue
			}
			switch handler(target) {
			case Ack:
				msg.Ack(false)
			case NackDiscard:
				msg.Nack(false, false)
			case NackRequeue:
				msg.Nack(false, true)
			}
		}
	}()

	return nil
}
