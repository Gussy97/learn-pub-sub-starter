package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const connectionString = "amqp://guest:guest@localhost:5672/"

	gamelogic.PrintServerHelp()

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("error opening connection: %s", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalf("error creating channel: %s", err)
	}

	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		fmt.Sprintf("%s.*", routing.GameLogSlug),
		pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatalf("error binding logs queue: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		cmd := words[0]

		switch cmd {
		case routing.PauseKey:
			fmt.Println("Publishing paused game state")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				})
			if err != nil {
				log.Fatalf("error publishing json: %v", err)
			}
		case routing.ResumeKey:
			fmt.Println("Publishing resumed game state")
			err = pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				})
			if err != nil {
				log.Fatalf("error sending resume command: %v", err)
			}
		case routing.QuitKey:
			fmt.Println("Closing the Server")
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
