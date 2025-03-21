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
	"google.golang.org/grpc/credentials"
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

	// https 環境かどうかを serverAddr で判断
	isProduction := !(len(serverAddr) >= 9 && serverAddr[:9] == "localhost")

	// HTTP ハンドラの設定
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "ゲスト"
		}

		// タイムアウト付きのコンテキスト
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var dialOpts []grpc.DialOption

		if isProduction {
			// Cloud Run環境用の設定
			// TLS認証情報を使用
			creds := credentials.NewClientTLSFromCert(nil, "")
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
		} else {
			// ローカル開発環境用の設定
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		// gRPC サーバーへの接続
		conn, err := grpc.DialContext(ctx, serverAddr, dialOpts...)
		if err != nil {
			log.Printf("サーバーへの接続に失敗: %v", err)
			http.Error(w, fmt.Sprintf("サーバーへの接続に失敗: %v", err), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// クライアントの作成
		client := pb.NewGreetingServiceClient(conn)

		// RPC 呼び出し
		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			log.Printf("RPC呼び出しに失敗: %v", err)
			http.Error(w, fmt.Sprintf("RPC呼び出しに失敗: %v", err), http.StatusInternalServerError)
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
