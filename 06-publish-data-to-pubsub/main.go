package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
)

func main() {
	ctx := context.Background()

	projectID := os.Getenv("PROJECT_ID")
	topicID := os.Getenv("TOPIC_ID")
	if topicID == "" {
		log.Fatal(fmt.Errorf("must be set TOPIC_ID to env variable"))
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte("Received plate data."),
		Attributes: map[string]string{
			"qrId":       "p000062",
			"shopNumber": "160",
			"hostname":   "LBCAM010",
			"popNumber":  "62",
			"state":      "0",
		},
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Published a message; msg ID: ", id)
}
