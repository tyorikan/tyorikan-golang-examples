# GEMINI.md: Cloud Run Worker Pool Pub/Sub Consumer

## 1. 概要

このプロジェクトは、Google Cloud Runの**Worker Pool**機能を活用し、Google Cloud Pub/Subから継続的にメッセージをPullして処理するバックグラウンドワーカーを実装します。

## 2. 要件定義

### 機能要件

-   **常時稼働**: アプリケーションはHTTPリクエストをリッスンせず、バックグラウンドで常時稼働し続けること。
-   **Pub/Sub連携**: 指定されたGoogle CloudプロジェクトのPub/Subサブスクリプションからメッセージを継続的に受信（Pull）すること。
-   **設定の外部化**: 以下の情報を環境変数経由で設定可能にすること。
    -   `PROJECT_ID`: Google CloudのプロジェクトID
    -   `SUBSCRIPTION_ID`: Pub/SubのサブスクリプションID
-   **グレースフルシャットダウン**: Cloud Runがコンテナを停止する際に送信する`SIGTERM`シグナルを捕捉し、処理中のメッセージを失うことなく安全にアプリケーションを終了できること。
-   **ロギング**: 受信したメッセージの内容を標準出力にログとして記録すること。
-   **メッセージ確認**: メッセージの処理が完了したら、Pub/Subに対して確認応答（Ack）を送信すること。

### 非機能要件

-   **プラットフォーム**: Google Cloud Run (Worker Pool)
-   **言語**: Go
-   **コンテナ化**: アプリケーションはDockerコンテナとしてパッケージ化すること。
-   **スケーラビリティ**: 負荷に応じてインスタンス数が自動でスケールするように構成できること（例: min-instances=1, max-instances=5）。
-   **効率性**: 軽量なベースイメージ（例: Alpine Linux）を使用した、効率的なコンテナイメージであること。

## 3. 実装指示書

### `main.go`

-   Go言語の`cloud.google.com/go/pubsub`ライブラリを使用する。
-   `main`関数でアプリケーションのルートとなる`context.Context`を`context.WithCancel`を用いて生成する。
-   `os/signal`と`syscall`パッケージを使い、`SIGINT`および`SIGTERM`シグナルを待機するgoroutineを起動する。
-   シグナルを受信したら、`context`の`cancel`関数を呼び出し、アプリケーションのシャットダウン処理を開始させる。
-   `pubsub.Subscription.Receive()`メソッドを、生成した`context`を渡して呼び出す。これにより、`context`がキャンセルされるまでメッセージの受信がブロックされる。
-   `Receive`のコールバック関数内で、メッセージの処理（この場合はロギング）と`msg.Ack()`の呼び出しを行う。
-   `Receive`から返されるエラーのうち、`context.Canceled`は正常な終了を示すため、エラーとして扱わない。

### `Dockerfile`

-   Goのビルド用と実行用の2ステージからなるマルチステージビルドを構成する。
-   **ビルダーステージ**:
    -   `golang:1.24-alpine`のような公式イメージをベースとする。
    -   `go.mod`と`go.sum`を先にコピーし、`go mod download`で依存関係をキャッシュさせる。
    -   ソースコードをコピーし、`CGO_ENABLED=0 GOOS=linux`の環境変数付きで`go build`を実行し、静的にリンクされたLinux実行ファイルを生成する。
-   **実行ステージ**:
    -   `alpine:latest`のような軽量イメージをベースとする。
    -   ビルダーステージから生成された実行ファイルのみをコピーする。
    -   `CMD`命令でその実行ファイルを指定する。
