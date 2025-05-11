package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func InitKafkaProducer() {
	broker := os.Getenv("KAFKA_BROKER")
	topic := os.Getenv("KAFKA_TOPIC") // Optional: can be passed dynamically to PublishEvent instead

	writer = &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	fmt.Println("Kafka producer initialized.")
}

func PublishEvent(topic string, message string) {
	err := writer.WriteMessages(context.Background(),
		kafka.Message{
			Topic: topic,
			Value: []byte(message),
			Time:  time.Now(),
		},
	)
	if err != nil {
		fmt.Printf("Error producing message to %s: %s\n", topic, err)
	} else {
		fmt.Printf("Message successfully produced to %s!\n", topic)
	}
}