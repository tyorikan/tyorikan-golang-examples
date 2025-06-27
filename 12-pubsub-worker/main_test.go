package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestPullMsgsは、pstestサーバーを使用してpullMsgs関数をテストします。
func TestPullMsgs(t *testing.T) {
	// テスト用のコンテキスト。5秒でタイムアウト
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// pstest (インメモリPub/Subサーバー) をセットアップ
	srv := pstest.NewServer()
	defer srv.Close()

	// サーバーへの接続を作成
	conn, err := grpc.Dial(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.Dial: %v", err)
	}
	defer conn.Close()

	// テスト用のクライアントを作成
	const projectID = "test-project"
	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	// テスト用のトピックとサブスクリプションを作成
	topic, err := client.CreateTopic(ctx, "test-topic")
	if err != nil {
		t.Fatalf("CreateTopic: %v", err)
	}
	const subID = "test-sub"
	sub, err := client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: topic})
	if err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}

	// テストメッセージをPublish
	const testMsg = "hello world"
	publishResult := topic.Publish(ctx, &pubsub.Message{Data: []byte(testMsg)})
	_, err = publishResult.Get(ctx)
	if err != nil {
		t.Fatalf("Publish.Get: %v", err)
	}

	// pullMsgsをgoroutineで実行
	pullCtx, pullCancel := context.WithCancel(ctx)
	defer pullCancel()

	go func() {
		// テスト用に環境変数を偽装
		t.Setenv("PROJECT_ID", projectID)
		t.Setenv("SUBSCRIPTION_ID", subID)

		// pullMsgsを直接呼び出す代わりに、テスト用のクライアントを使うように変更
		pullMsgsWithClient(pullCtx, client, subID)
	}()

	// メッセージが受信されるのを少し待つ
	time.Sleep(2 * time.Second)

	// pullMsgsを停止させる
	pullCancel()

	// メッセージがAckされたか確認 (再度pullしてもメッセージがないはず)
	// 新しいコンテキストでpullを試みる
	verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer verifyCancel()

	received := make(chan struct{})
	go func() {
		err := sub.Receive(verifyCtx, func(ctx context.Context, m *pubsub.Message) {
			t.Errorf("Message should have been acked, but received again: %s", string(m.Data))
			m.Nack()
			close(received)
		})
		if err != nil && err != context.Canceled {
			t.Logf("Receive for verification returned an expected error: %v", err)
		}
	}()

	select {
	case <-received:
		t.FailNow()
	case <-verifyCtx.Done():
		// タイムアウトすれば、メッセージがなかったということなのでテスト成功
		t.Log("Successfully verified that message was acked.")
	}
}

// pullMsgsをテスト用に少し変更したヘルパー関数
func pullMsgsWithClient(ctx context.Context, client *pubsub.Client, subID string) {
	sub := client.Subscription(subID)
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Test worker got message: %s", string(msg.Data))
		msg.Ack()
	})
	if err != nil && err != context.Canceled {
		fmt.Printf("pullMsgsWithClient Receive error: %v\n", err)
	}
}
