FROM golang:1.24 as builder

WORKDIR /app

# 依存関係のコピーとダウンロード
COPY go.mod go.sum* ./
RUN go mod download

# ソースコードのコピー
COPY . .

# クライアントのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /client ./client

# 実行用の軽量イメージ
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=builder /client /client

# クライアントの実行
CMD ["/client"]