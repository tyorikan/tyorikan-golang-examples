package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	// Logging the start and end of each request
	r.Use(middleware.Logger)

	// Path, Method 単位で定義
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Post("/receiver", func(w http.ResponseWriter, r *http.Request) {
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
		// TODO set attributes from request body
		result := t.Publish(ctx, &pubsub.Message{
			Data: []byte("Received post data."),
			Attributes: map[string]string{
				"id":         "p000062",
				"shopNumber": "160",
				"hostname":   "LBCAM010",
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
	})

	// Listen port
	http.ListenAndServe(":8080", r)
}
