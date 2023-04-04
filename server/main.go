/*
	This main program generates random numbers and gives it as order id to rabbitmq which is setup_queue.
	We enable the program to generate order id by sending a request to /order end point.

	for testing u can use request.py file

*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func randomN(a int, b int) int {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return a + rng.Intn(b-a+1)
}

var client *redis.Client

func init() {

	//!Connect to Redis
	client = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Failed to connect to Redis:", err)
		return
	}

	fmt.Println("Connected to Redis:", pong)

}

// /order endpoint create order id and send setup_queue channel
func order() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//Random order id generator
		orderID := fmt.Sprintf("%d", randomN(10000, 10000000))
		ctx := context.Background()

		//connect rabbitmq
		conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ch, err := conn.Channel()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// make queue
		q, err := ch.QueueDeclare(
			"setup_queue",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//marshall order id
		msgBody := map[string]string{"orderid": orderID}
		msgBodyBytes, err := json.Marshal(msgBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//send order id in queue
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

		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"orderid": orderID}
		json.NewEncoder(w).Encode(response)
	}
}

func status() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// Read request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}

		// Parse JSON data from request body
		var jsonDataRaw interface{}
		err = json.Unmarshal(body, &jsonDataRaw)
		if err != nil {
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}
		jsonData := jsonDataRaw.(map[string]interface{})
		orderid := jsonData["orderid"].(string)

		// Set timeout for Redis query
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		fmt.Printf("Got orderid: %s\n", orderid)

		// Set response headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		// Query Redis for the order status and return as response
		value, _ := client.Get(ctx, orderid).Result()
		jsonData["status"] = value
		response := map[string]string{"status": value}
		json.NewEncoder(w).Encode(response)

	}
	return http.HandlerFunc(fn)
}

func main() {

	http.Handle("/order", order())
	http.Handle("/status", status())

	log.Print("Listening...")
	log.Fatal(http.ListenAndServe(":8787", nil))

}
