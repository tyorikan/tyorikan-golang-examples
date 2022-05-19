package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
)

const collectionPrefix = "plate-states-"
const openState = 0
const closeState = 1
const notDiscard = 0
const discard = 1
const discardDuration = 60 * 60 // 廃棄 Duration 1 時間（仮）

type PlateStates struct {
	QrID      string `firestore:"qrId"`
	PopNumber int16  `firestore:"popNumber"`
	State     int8   `firestore:"state"`
}

type PlateDocument struct {
	PlateStates     PlateStates `firestore:"plateStates"`
	Revision        int8        `firestore:"revision"`
	ServedTimestamp int64       `firestore:"servedTimestamp"` // 0 -> 1
	EmptyTimestamp  int64       `firestore:"emptyTimestamp"`  // 1 -> 0
	UpdateTime      time.Time   `firestore:"updateTime"`
	DiscardFlag     int8        `firestore:"discardFlag"`
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

func UpdatePlates(ctx context.Context, shopNumber int64, plate PlateStates) (*PlateDocument, error) {
	if FirestoreClient == nil {
		return nil, fmt.Errorf("firestore client hasn't initialized yet")
	}

	collectionName := collectionPrefix + strconv.FormatInt(shopNumber, 10)
	col := FirestoreClient.Collection(collectionName)
	doc := col.Doc(plate.QrID)

	now := time.Now()
	data := &PlateDocument{
		PlateStates: plate,
		Revision:    0, // Not used
		UpdateTime:  now,
		DiscardFlag: notDiscard,
	}

	// 現在のデータ取得
	docSnap, err := doc.Get(ctx)
	if docSnap == nil && err != nil {
		return nil, err
	}

	// Plate の state 変化から、提供開始時刻 or 空になった時刻を指定

	// 新規データ
	if !docSnap.Exists() {
		if plate.State == openState {
			data.EmptyTimestamp = now.Unix()
		} else if plate.State == closeState {
			data.ServedTimestamp = now.Unix()
		}

		_, err = doc.Create(ctx, data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	var preData PlateDocument
	err = docSnap.DataTo(&preData)
	if err != nil {
		return nil, err
	}

	// デフォ以前の状態で上書き
	data.ServedTimestamp = preData.ServedTimestamp
	data.EmptyTimestamp = preData.EmptyTimestamp
	data.DiscardFlag = preData.DiscardFlag

	// 以前の State と、送信されたデータに差異があれば時刻を更新
	if preData.PlateStates.State != plate.State {
		if plate.State == openState {
			data.EmptyTimestamp = now.Unix()
			data.ServedTimestamp = 0
			data.DiscardFlag = notDiscard
		} else if plate.State == closeState {
			data.EmptyTimestamp = 0
			data.ServedTimestamp = now.Unix()
		}
	}

	// ServedTime が一定以上過去の場合に、廃棄 Flag を On
	if preData.ServedTimestamp > 0 && preData.ServedTimestamp+discardDuration < now.Unix() {
		data.DiscardFlag = discard
	}

	_, err = doc.Set(ctx, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
