package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"
)

func main() {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatalf("PROJECT_ID environment variable must be set.")
	}
	subID := os.Getenv("SUBSCRIPTION_ID")
	if subID == "" {
		log.Fatalf("SUBSCRIPTION_ID environment variable must be set.")
	}

	// アプリケーションのメインコンテキストを作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("Starting Pub/Sub pull worker for project '%s', subscription '%s'", projectID, subID)

	if err := pullMsgs(ctx, projectID, subID); err != nil {
		// context.Canceledはタイムアウトによる正常終了なのでエラーとして扱わない
		if err == context.Canceled {
			log.Println("Context canceled, worker finished pulling messages successfully.")
		} else {
			log.Fatalf("Failed to pull messages: %v", err)
		}
	}
	log.Println("Worker shutting down.")
}

// pullMsgsは指定されたサブスクリプションからメッセージをpullします
func pullMsgs(ctx context.Context, projectID, subID string) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %w", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)

	// Receiveはブロッキング呼び出しです。
	// contextがキャンセルされるか、致命的なエラーが発生するまでメッセージを受信し続けます。
	var mu sync.Mutex
	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		defer mu.Unlock()

		log.Printf("Got message: %s", string(msg.Data))

		// ここでメッセージに対するビジネスロジicを実装します
		// 例: データベースへの書き込み、別のAPIの呼び出しなど

		// 処理が成功したらメッセージをAckします
		msg.Ack()
	})

	// context.Canceledは期待されるエラーなので、呼び出し元で処理します
	if err != nil && err != context.Canceled {
		return fmt.Errorf("sub.Receive: %w", err)
	}

	return nil
}
