package app

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
)

const plateStateCollectionPrefix = "plate-states-"
const menuCountCollectionPrefix = "menu-count-"
const plateDocPrefix = "qrid-"
const countDocPrefix = "pop-number-"

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

	plateStateCollectionPath := plateStateCollectionPrefix + strconv.FormatInt(shopNumber, 10)
	menuCountCollectionPath := menuCountCollectionPrefix + strconv.FormatInt(shopNumber, 10)
	plateRef := FirestoreClient.Collection(plateStateCollectionPath).Doc(plateDocPrefix + plate.QrID)
	countRef := FirestoreClient.Collection(menuCountCollectionPath).Doc(countDocPrefix + strconv.FormatInt(int64(plate.PopNumber), 10))

	var data *PlateDocument

	// Plate の state 変化から、提供開始時刻 or 空になった時刻を指定。PopNumber の count up/down
	err := FirestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// 現在の Plate データ取得
		docSnap, err := tx.Get(plateRef)
		if docSnap == nil && err != nil {
			return err
		}

		// 新規データ
		if !docSnap.Exists() {
			data, err = createNewData(ctx, tx, plateRef, countRef, plate)
			if err != nil {
				return err
			}
			return nil
		}

		// データ更新
		data, err = updatePlateData(ctx, tx, plateRef, docSnap, countRef, plate)
		return err
	})

	if err != nil {
		return nil, err
	}
	return data, nil
}

func createNewData(ctx context.Context, tx *firestore.Transaction, plateRef *firestore.DocumentRef, countRef *firestore.DocumentRef, plate PlateStates) (*PlateDocument, error) {
	now := time.Now()
	data := &PlateDocument{
		PlateStates: plate,
		Revision:    0, // Not used
		UpdateTime:  now,
		DiscardFlag: notDiscard,
	}
	incrementValue := 0

	if plate.State == openState {
		data.EmptyTimestamp = now.Unix()
	} else if plate.State == closeState {
		data.ServedTimestamp = now.Unix()
		incrementValue = 1
	}

	err := tx.Create(plateRef, data)
	if err != nil {
		return nil, err
	}

	// 商品（= popNumber）数カウントアップ
	countData := map[string]interface{}{
		"count": firestore.Increment(incrementValue),
	}
	err = tx.Create(countRef, countData)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func updatePlateData(ctx context.Context, tx *firestore.Transaction, plateRef *firestore.DocumentRef, docSnap *firestore.DocumentSnapshot, countRef *firestore.DocumentRef, plate PlateStates) (*PlateDocument, error) {
	var preData PlateDocument
	err := docSnap.DataTo(&preData)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	data := &PlateDocument{
		PlateStates: plate,
		Revision:    0, // Not used
		UpdateTime:  now,
		DiscardFlag: notDiscard,
	}

	// デフォ以前の状態で上書き
	data.ServedTimestamp = preData.ServedTimestamp
	data.EmptyTimestamp = preData.EmptyTimestamp
	data.DiscardFlag = preData.DiscardFlag

	// 以前の State と、送信されたデータに差異があれば時刻を更新
	incrementValue := 0
	if preData.PlateStates.State != plate.State {
		if plate.State == openState {
			data.EmptyTimestamp = now.Unix()
			data.ServedTimestamp = 0
			data.DiscardFlag = notDiscard
			incrementValue = -1
		} else if plate.State == closeState {
			data.EmptyTimestamp = 0
			data.ServedTimestamp = now.Unix()
			incrementValue = 1
		}
	}

	// ServedTime が一定以上過去の場合に、廃棄 Flag を On
	if preData.ServedTimestamp > 0 && preData.ServedTimestamp+discardDuration < now.Unix() {
		data.DiscardFlag = discard
	}

	err = tx.Set(plateRef, data)
	if err != nil {
		return nil, err
	}

	// 商品（= popNumber）数カウントアップ/ダウン
	err = tx.Update(countRef, []firestore.Update{
		{Path: "count", Value: firestore.Increment(incrementValue)},
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}
