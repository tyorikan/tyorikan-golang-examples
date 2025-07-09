

# Cloud Run & Spanner API Demo

このプロジェクトは、Google Cloud Run上で動作するGo言語製のWeb APIです。リクエストを受け取ってGoogle Cloud Spannerにデータを保存し、監査ログをGoogle Cloud Pub/Subに発行します。

## ✨ 主な機能

- **CRUD API**: 単一のエンドポイント (`POST /products`) で製品データの作成・更新・削除（CRUD）を実現します。
- **Spanner連携**: Google Cloud Spannerをデータベースとして使用し、製品情報を永続化します。
- **Pub/Sub監査ログ**: すべてのデータベース操作の成功時に、監査ログをJSON形式でPub/Subトピックに発行します。これにより、BigQueryなどでのデータ分析が容易になります。
- **コンテナベース**: Dockerfileが付属しており、Cloud Runへのデプロイが簡単に行えます。

---

## 🛠️ セットアップ

### 1. 前提条件

- Go (1.22 以上)
- Docker
- Google Cloud SDK (`gcloud` CLI)
- 上記コンポーネントを利用可能な権限を持つGoogle Cloudプロジェクト

### 2. Google Cloud環境の準備

1.  **APIの有効化**: プロジェクトで以下のAPIを有効にします。
    - Cloud Spanner API
    - Cloud Pub/Sub API
    - Cloud Run Admin API
    - Cloud Build API

2.  **Spannerデータベースの作成**
    - Spannerインスタンスとデータベースを作成します。
    - 以下のDDL（Data Definition Language）を実行して、`Products`テーブルを作成します。

      ```sql
      CREATE TABLE Products (
          ProductId   STRING(36) NOT NULL,
          ProductName STRING(255),
          Description STRING(1024),
          Attributes  JSON,
          CreatedAt   TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
          UpdatedAt   TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true)
      ) PRIMARY KEY (ProductId);
      ```

3.  **Pub/Subトピックの作成**
    - 監査ログを送信するためのPub/Subトピックを1つ作成します。

### 3. アプリケーションの構成

アプリケーションは、起動時に以下の環境変数を読み込みます。ローカルで実行する場合やCloud Runにデプロイする際に設定してください。

- `GCP_PROJECT_ID`: あなたのGoogle CloudプロジェクトID
- `SPANNER_INSTANCE_ID`: 作成したSpannerインスタンスのID
- `SPANNER_DATABASE_ID`: 作成したSpannerデータベースのID
- `PUBSUB_TOPIC_ID`: 作成したPub/SubトピックのID
- `PORT`: (オプション) アプリケーションがリッスンするポート。デフォルトは`8080`です。

---

## 🚀 実行方法

### ローカルでの実行

1.  リポジトリをクローンし、このディレクトリに移動します。

2.  必要なGoパッケージをダウンロードします。
    ```sh
    go mod tidy
    ```

3.  環境変数を設定します。
    ```sh
    export GCP_PROJECT_ID="your-gcp-project-id"
    export SPANNER_INSTANCE_ID="your-spanner-instance-id"
    export SPANNER_DATABASE_ID="your-spanner-db-id"
    export PUBSUB_TOPIC_ID="your-pubsub-topic-id"
    ```

4.  アプリケーションを実行します。
    ```sh
    go run .
    ```
    サーバーが`localhost:8080`で起動します。

### Dockerでの実行

1.  Dockerイメージをビルドします。
    ```sh
    docker build -t spanner-api-demo .
    ```

2.  ビルドしたイメージを環境変数を設定して実行します。
    ```sh
    docker run -p 8080:8080 \
      -e GCP_PROJECT_ID="your-gcp-project-id" \
      -e SPANNER_INSTANCE_ID="your-spanner-instance-id" \
      -e SPANNER_DATABASE_ID="your-spanner-db-id" \
      -e PUBSUB_TOPIC_ID="your-pubsub-topic-id" \
      spanner-api-demo
    ```

---

## ⚙️ API利用方法

`curl`コマンドの例を以下に示します。

### CREATE: 新規製品の作成

```sh
curl -X POST http://localhost:8080/products \
-H "Content-Type: application/json" \
-d '{
  "operation": "CREATE",
  "product_id": "product-001",
  "product_name": "プレミアム・ウィジェット",
  "description": "最高品質の素材のみを使用したプレミアムなウィジェットです。",
  "attributes": {
    "color": "gold",
    "weight_kg": 1.2
  }
}'
```

### UPDATE: 既存製品の更新

```sh
curl -X POST http://localhost:8080/products \
-H "Content-Type: application/json" \
-d '{
  "operation": "UPDATE",
  "product_id": "product-001",
  "product_name": "プレミアム・ウィジェット V2",
  "description": "さらに改良された第二世代のウィジェット。"
}'
```

### DELETE: 既存製品の削除

```sh
curl -X POST http://localhost:8080/products \
-H "Content-Type: application/json" \
-d '{
  "operation": "DELETE",
  "product_id": "product-001"
}'
```
