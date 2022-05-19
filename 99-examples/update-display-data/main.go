package main

import (
	"context"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"go.uber.org/zap"
	"update-display-data/app"
)

func main() {
	// Create firestore client
	ctx := context.Background()
	client, err := app.CreateClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		app.Logger.Fatal("Failed to create firestore client", zap.Error(err))
	}

	// defer close
	defer func(firestoreClient *firestore.Client) {
		err := app.CloseClient(client)
		if err != nil {
			app.Logger.Warn("Failed to close firestore client", zap.Error(err))
		}
	}(client)

	// Listen port
	err = http.ListenAndServe(":8080", app.Router())
	if err != nil {
		app.Logger.Fatal("Failed to serve endpoints", zap.Error(err))
	}
}
