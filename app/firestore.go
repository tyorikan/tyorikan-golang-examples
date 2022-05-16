package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

const platesCollection = "plates"

type Plate struct {
	QrID       string
	ShopNumber int64
	Hostname   string
	PopNumber  int16
	State      int8
}

type TimelineData struct {
	Plate      Plate
	Revision   int8
	CreateTime time.Time
	UpdateTime time.Time
}

var FirestoreClient *firestore.Client

func CreateClient(ctx context.Context, projectID string) (*firestore.Client, error) {
	if FirestoreClient != nil {
		return FirestoreClient, nil
	}

	if projectID == "" {
		return nil, fmt.Errorf("must be set `PROJECT_ID` to the environment variable")
	}
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	FirestoreClient = client
	return FirestoreClient, nil
}

func CloseClient(client *firestore.Client) error {
	return client.Close()
}

func AddPlateStates(ctx context.Context, plate Plate) (*TimelineData, error) {
	if FirestoreClient == nil {
		return nil, fmt.Errorf("firestore client hasn't initialized yet")
	} else if plate.ShopNumber == 0 {
		return nil, fmt.Errorf("shop number must be set")
	}

	collectionName := platesCollection + "-" + strconv.FormatInt(plate.ShopNumber, 10)
	col := FirestoreClient.Collection(collectionName)

	// UUID = DocID とする
	uuidObj, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	// Document 登録
	doc := col.Doc(uuidObj.String())
	now := time.Now()
	data := &TimelineData{
		Plate:      plate,
		Revision:   0, // Not used
		CreateTime: now,
		UpdateTime: now,
	}

	_, err = doc.Create(ctx, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
