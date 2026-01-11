package gamelogic

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (gs *GameState) CommandSpam(words []string, publishCh *amqp.Channel) error {
	if len(words) < 2 {
		return errors.New("usage: spam <n>")
	}
	n, err := strconv.Atoi(words[1])
	if err != nil {
		return fmt.Errorf("error converting %s to int: %v\n", words[1], err)
	}
	for i := 0; i < n; i++ {
		maliciousLog := GetMaliciousLog()
		if err := pubsub.PublishGameLog(publishCh, gs.GetUsername(), maliciousLog); err != nil {
			return fmt.Errorf("error publishing game log: %v\n", err)
		}
	}
	return nil
}
