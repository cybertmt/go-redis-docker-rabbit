package main

import (
	"net/http"
	"os"
	"os/signal"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var rabbit_host = os.Getenv("RABBIT_HOST")
var rabbit_port = os.Getenv("RABBIT_PORT")
var rabbit_user = os.Getenv("RABBIT_USERNAME")
var rabbit_password = os.Getenv("RABBIT_PASSWORD")

func main() {
	router := httprouter.New()

	router.POST("/publish/:message", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		submit(w, r, p)
	})

	go func() {
		log.Println("Server running @ :8011")
		log.Fatal(http.ListenAndServe(":8011", router))
	}()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
	log.Println("Publisher stopped")
}

func submit(writer http.ResponseWriter, request *http.Request, p httprouter.Params) {
	message := p.ByName("message")

	log.Println("Received message: " + message)

	conn, err := amqp.Dial("amqp://" + rabbit_user + ":" + rabbit_password + "@" + rabbit_host + ":" + rabbit_port + "/")

	if err != nil {
		log.Printf("%s: %s", "Failed to connect to RabbitMQ", err)
	}

	defer conn.Close()

	ch, err := conn.Channel()

	if err != nil {
		log.Printf("%s: %s", "Failed to open a channel", err)
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		"publisher", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare a queue", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})

	if err != nil {
		log.Fatalf("%s: %s", "Failed to publish a message", err)
	}

	log.Println("publish success!")
}
