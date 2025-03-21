package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "demo/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedGreetingServiceServer
}

// SayHello は GreetingService の実装
func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("リクエストを受信: %v", req.GetName())
	return &pb.HelloResponse{Message: fmt.Sprintf("こんにちは, %s!", req.GetName())}, nil
}

func main() {
	// ポート設定
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// リスナーの作成
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("ポート %s のリッスンに失敗: %v", port, err)
	}

	// gRPCサーバーの作成
	s := grpc.NewServer()
	// サービス実装を登録
	pb.RegisterGreetingServiceServer(s, &server{})
	// gRPC リフレクションを有効化（デバッグに便利）
	reflection.Register(s)

	log.Printf("サーバーがポート %s で起動しました", port)
	// サーバーの起動
	if err := s.Serve(lis); err != nil {
		log.Fatalf("サーバーの起動に失敗: %v", err)
	}
}
