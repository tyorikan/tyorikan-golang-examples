package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	pb "demo/proto"

	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	grpcMetadata "google.golang.org/grpc/metadata"
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

			// Create an identity token.
			audience := serverAddr
			if strings.Contains(audience, ":") {
				audience = strings.Split(audience, ":")[0]
			}
			audience = "https://" + audience
			tokenSource, err := idtoken.NewTokenSource(ctx, audience)
			if err != nil {
				log.Printf("認証トークンの生成に失敗: %v", err)
				http.Error(w, fmt.Sprintf("サーバーへの接続に失敗: %v", err), http.StatusInternalServerError)
				return
			}
			token, err := tokenSource.Token()
			if err != nil {
				log.Printf("認証トークンの取得に失敗: %v", err)
				http.Error(w, fmt.Sprintf("サーバーへの接続に失敗: %v", err), http.StatusInternalServerError)
				return
			}

			// Add token to gRPC Request.
			ctx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token.AccessToken)
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

		// https://inet-ip.info/ip にアクセスし、自身の IP アドレスを取得する
		egressIP, err := getEgressIP()
		if err != nil {
			log.Printf("IPアドレスの取得に失敗: %v", err)
			http.Error(w, fmt.Sprintf("IPアドレスの取得に失敗: %v", err), http.StatusInternalServerError)
			return
		}

		// gRPC クライアントの作成
		client := pb.NewGreetingServiceClient(conn)

		// RPC 呼び出し
		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			log.Printf("RPC呼び出しに失敗: %v", err)
			http.Error(w, fmt.Sprintf("RPC呼び出しに失敗: %v", err), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "サーバーからの応答: %s\nIP(gRPC client): %s", resp.GetMessage(), egressIP)
	})

	// HTTPサーバーの起動
	log.Printf("クライアントがポート %s で起動しました", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("HTTPサーバーの起動に失敗: %v", err)
	}
}

// インターネットアクセスに使用される IP アドレスを確認する
func getEgressIP() (string, error) {
	// https://inet-ip.info/ip にアクセスし、自身の IP アドレスを取得する
	resp, err := http.Get("https://inet-ip.info/ip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// レスポンスボディを読み込む
	buf := make([]byte, 32)
	n, err := resp.Body.Read(buf)
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}
