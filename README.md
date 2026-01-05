# Primind Remind Time Management

リマインダーの時間管理サービス。リマインダーの作成・取得・更新・削除、及びNATS JetStreamへのイベント発行を担当。

## エンドポイント

| メソッド | エンドポイント | 概要 |
|---------|------|------|
| POST | /api/v1/reminds | リマインダー作成 |
| GET | /api/v1/reminds | 時間範囲内のリマインダー取得 |
| POST | /api/v1/reminds/:id/throttled | スロットル状態の更新 |
| DELETE | /api/v1/reminds/:id | リマインダー削除 |
| POST | /api/v1/reminds/cancel | タスクIDでリマインダーをキャンセル |

## Proto定義

- `proto/remind/v1/remind.proto`
- `proto/common/v1/common.proto`
  - Enum: `TaskType`

## 依存

- PostgreSQL v18
- Cloud Pub/Sub / NATS 2.10 (JetStream)
