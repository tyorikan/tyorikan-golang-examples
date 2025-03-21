package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	pb "demo/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// ポート設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// サーバーのアドレス
	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = "localhost:8081" // ローカルテスト用のデフォルト値
	}

	// HTTP ハンドラの設定
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "ゲスト"
		}

		// gRPC クライアント接続の作成
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err := grpc.DialContext(ctx, serverAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock())
		if err != nil {
			log.Printf("サーバーへの接続に失敗: %v", err)
			http.Error(w, "サーバーへの接続に失敗", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// クライアントの作成
		client := pb.NewGreetingServiceClient(conn)

		// RPC 呼び出し
		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			log.Printf("RPC呼び出しに失敗: %v", err)
			http.Error(w, "RPC呼び出しに失敗", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "サーバーからの応答: %s", resp.GetMessage())
	})

	// HTTPサーバーの起動
	log.Printf("クライアントがポート %s で起動しました", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("HTTPサーバーの起動に失敗: %v", err)
	}
}
