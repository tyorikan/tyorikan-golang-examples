package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
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

	col := client.Collection("plates-sample")

	// 10 秒間の皿データ取得
	now := time.Now()
	t, err := subDate(now, 10, "second")
	if err != nil {
		log.Fatal(err)
	}
	iter := col.Where("timestamp", ">=", t.Unix()).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(doc.Data())
	}

	// Close client
	defer client.Close()
}

type plate struct {
	QrID       string `firestore:"qrId"`
	ShopNumber int64  `firestore:"shopNumber"`
	Hostname   string `firestore:"hostname"`
	PopNumber  int16  `firestore:"popNumber"`
	State      int8   `firestore:"state"`
}

type timelineData struct {
	Plate      plate     `firestore:"plate"`
	Revision   int8      `firestore:"revision"`
	Timestamp  int64     `firestore:"timestamp"`
	CreateTime time.Time `firestore:"createTime"`
	UpdateTime time.Time `firestore:"updateTime"`
}

// subDate subtracts n*unit duration from t and return the result.
func subDate(t time.Time, n int, unit string) (time.Time, error) {
	if strings.HasSuffix(unit, "s") {
		unit = string([]byte(unit)[:len(unit)-1])
	}
	switch unit {
	case "year":
		return t.AddDate(-1*n, 0, 0), nil
	case "month":
		return t.AddDate(0, -1*n, 0), nil
	case "week":
		return t.AddDate(0, 0, -7*n), nil
	case "day":
		return t.AddDate(0, 0, -1*n), nil
	case "hour":
		return t.Add(time.Duration(-1*n) * time.Hour), nil
	case "minute":
		return t.Add(time.Duration(-1*n) * time.Minute), nil
	case "second":
		return t.Add(time.Duration(-1*n) * time.Second), nil
	default:
		return t, fmt.Errorf("unsupported time unit: %v", unit)
	}
}
