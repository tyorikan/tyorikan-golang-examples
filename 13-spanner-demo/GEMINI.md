# GEMINI.md: Cloud Run & Spanner Demo Application v2

## 1. 概要

このプロジェクトは、Google Cloud RunとGoogle Cloud Spannerを利用して、検索プラットフォーム向けの製品マスタデータを受け取り、保存・更新・削除するWeb APIを開発します。

クライアントからの単一エンドポイントへのHTTP POSTリクエストで、`CREATE`, `UPDATE`, `DELETE` の操作種別を判別し、Spannerデータベースにデータを永続化します。また、すべての成功したデータベース操作は、監査ログとしてPub/Subトピックに発行され、BigQueryなどでの後続処理に利用されることを想定しています。

## 2. 要件定義

### 2.1. 機能要件

-   **HTTPサーバー**: 単一の `POST /products` エンドポイントを公開すること。
-   **CRUD操作**: リクエストボディ内の `operation` フィールドに基づき、製品データの作成、更新、削除を実行すること。
-   **データ受信と検証**: JSON形式のデータを受け取り、操作種別に応じて必須フィールドの存在を検証すること。
-   **データベース永続化**: 検証済みのデータをSpannerデータベースの`Products`テーブルに反映すること。
-   **監査ログ**: データベース操作成功後、操作内容を含む監査メッセージをPub/Subトピックに発行すること。
-   **レスポンス**: 処理の成功・失敗に応じて、適切なHTTPステータスコードとJSONレスポンスを返すこと。
-   **ロギング**: リクエストの受信、処理結果、エラーなどを標準出力にログとして記録すること。

### 2.2. 非機能要件

-   **プラットフォーム**: Google Cloud Run
-   **データベース**: Google Cloud Spanner
-   **メッセージング**: Google Cloud Pub/Sub
-   **言語**: Go (chiルーター, go-playground/validatorライブラリ使用)
-   **コンテナ化**: アプリケーションはDockerコンテナとしてパッケージ化すること。
-   **設定の外部化**: 以下の情報を環境変数経由で設定可能にすること。
    -   `GCP_PROJECT_ID`: Google CloudのプロジェクトID
    -   `SPANNER_INSTANCE_ID`: SpannerのインスタンスID
    -   `SPANNER_DATABASE_ID`: SpannerのデータベースID
    -   `PUBSUB_TOPIC_ID`: 監査ログを発行するPub/SubトピックID
-   **テスト**: DBとPub/Subをモック化したユニットテストを実装し、正常系・異常系の両方をカバーすること。

## 3. APIエンドポイント

### `POST /products`

-   **説明**: 製品データのCRUD操作を実行します。
-   **リクエストボディ**: `operation`フィールドで操作を決定します。

    ```json
    {
      "operation": "CREATE", // "CREATE", "UPDATE", "DELETE" のいずれか
      "product_id": "product-12345",
      "product_name": "高性能ウィジェット", // DELETE時は不要
      "description": "最新技術を駆使した高性能なウィジェットです。",
      "attributes": {
        "color": "blue",
        "weight_kg": 0.5
      }
    }
    ```

-   **レスポンス (成功時)**:
    -   ステータスコード: `200 OK`
    -   ボディ:
        ```json
        {
          "status": "success",
          "operation": "CREATE",
          "product_id": "product-12345"
        }
        ```

## 4. データベース設計 (Google Cloud Spanner)

### `Products` テーブル (変更なし)

| カラム名 | 型 | 説明 |
| --- | --- | --- |
| ProductId | STRING(36) | 製品ID (主キー) |
| ProductName | STRING(255) | 製品名 |
| Description | STRING(1024) | 製品説明 |
| Attributes | JSON | 製品の属性情報 |
| CreatedAt | TIMESTAMP | 作成日時 (Commit Timestamp) |
| UpdatedAt | TIMESTAMP | 更新日時 (Commit Timestamp) |

## 5. Pub/Subメッセージ設計

BigQueryのサブスクリプションで直接取り込めるよう、JSON形式でログを出力します。

### `AuditLogMessage`

```json
{
  "operation": "CREATE",
  "product_id": "product-12345",
  "request_data": { ... }, // 受信したリクエストボディ全体
  "timestamp": "2025-07-09T12:00:00Z" // ISO 8601形式のUTCタイムスタンプ
}
```

## 6. 実装指示書

### `main.go`

-   ルーターには `go-chi/chi` を使用し、リクエストロギングのミドルウェアを導入する。
-   リクエストボディの検証には `go-playground/validator` を使用する。
-   データベース操作とPub/Sub発行処理は、テストでモック化できるようインターフェースを定義し、ハンドラはインターフェースに依存する設計とする。
-   `main`関数で、SpannerとPub/Subのクライアントを初期化し、ハンドラにDI（依存性注入）する。

### `Dockerfile` (変更なし)

-   Goのビルド用と実行用の2ステージからなるマルチステージビルドを構成する。
