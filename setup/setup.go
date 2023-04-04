/*
	take order id in rabbitmq and push redis order id status: preparing
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func main() {

	//connect to Rabbit
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	//redis connection
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	//connect order_queue
	q, err := ch.QueueDeclare(
		"setup_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	//take the order id in rabbitmq
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	for msg := range msgs {
		if msg.Body == nil {
			fmt.Println("error")
		}
		log.Println("Message received...")
		fmt.Printf("unmarshalling message: ")

		//unmarshal the message
		var data map[string]string
		if err := json.Unmarshal(msg.Body, &data); err != nil {
			log.Printf("Error decoding message body: %v\n", err)
			continue
		}
		orderid, ok := data["orderid"]

		if !ok {
			log.Printf("Missing orderid field in message body: %s\n", msg.Body)
			continue
		}

		//into thread for take order id from program and send status in redis.
		go func(orderid string) {
			fmt.Printf("orderid %s", orderid)

			ctx := context.Background()
			err := client.Set(ctx, orderid, "preparing", 0).Err()
			if err != nil {
				fmt.Printf("Error setting initial order status for order %s: %s\n", orderid, err)
				return
			}
			//status wait 30 second in redis then start delivering status
			time.Sleep(30 * time.Second)

			//create another queue
			q, err := ch.QueueDeclare(
				"delivering_queue",
				false,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				log.Fatalf("Failed to declare a queue: %v", err)
			}

			//marshall orderid and send delivering_queue
			msgBody := map[string]string{"orderid": orderid}
			msgBodyBytes, err := json.Marshal(msgBody)
			if err != nil {
				fmt.Println(err)
				return
			}

			_ = ch.PublishWithContext(
				ctx,
				"",
				q.Name,
				false,
				false,
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        msgBodyBytes,
				},
			)

		}(orderid)
	}

}
