package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
)

func main() {
	ctx := context.Background()

	// Firestore に接続するために必要な、ProjectId を環境変数から取得
	projectID := os.Getenv("PROJECT_ID")

	// Firestore に接続し、client 取得
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	// Collection 一覧取得（確認用）
	collectionRefs, err := client.Collections(ctx).GetAll()
	for _, v := range collectionRefs {
		log.Print(v.Path)
	}

	// Close client
	defer client.Close()
}
