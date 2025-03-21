FROM golang:1.24 as builder

WORKDIR /app

# 依存関係のコピーとダウンロード
COPY go.mod go.sum* ./
RUN go mod download

# ソースコードのコピー
COPY . .

# サーバーのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./server

# 実行用の軽量イメージ
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=builder /server /server

# サーバーの実行
CMD ["/server"]