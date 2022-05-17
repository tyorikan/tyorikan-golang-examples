package main

import (
	"context"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
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

	// UUID = DocID とする
	uuidObj, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}

	col := client.Collection("plates-sample")
	doc := col.Doc(uuidObj.String())
	now := time.Now()
	data := &timelineData{
		Plate: plate{
			QrID:       "p0000062",
			ShopNumber: 160,
			Hostname:   "LBCAM010",
			PopNumber:  62,
			State:      0,
		},
		Revision:   0, // Not used
		CreateTime: now,
		UpdateTime: now,
	}

	// Document 登録
	wr, err := doc.Create(ctx, data)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Document created: ", wr.UpdateTime)

	// Close client
	defer client.Close()
}

type plate struct {
	QrID       string
	ShopNumber int64
	Hostname   string
	PopNumber  int16
	State      int8
}

type timelineData struct {
	Plate      plate
	Revision   int8
	CreateTime time.Time
	UpdateTime time.Time
}
