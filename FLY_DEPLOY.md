# Fly.ioでのデプロイ手順（クレジットカード不要・永続ストレージ対応）

Fly.ioは、クレジットカード不要で永続ストレージを使用できるホスティングサービスです。

## メリット

- ✅ クレジットカード不要
- ✅ 永続ストレージ（ボリューム）が無料プランで使用可能
- ✅ スリープなし（無料プランでも）
- ✅ 高速なグローバルCDN

## 前提条件

- Fly.ioアカウント（[https://fly.io](https://fly.io) で無料登録）
- Fly CLIのインストール

## Fly CLIのインストール

```bash
# macOS
curl -L https://fly.io/install.sh | sh

# または Homebrew
brew install flyctl
```

## デプロイ手順

### 1. Fly.ioにログイン

```bash
fly auth login
```

### 2. バックエンドのデプロイ

```bash
# アプリを作成（初回のみ）
fly apps create poppo-backend

# 永続ボリュームを作成（データベースとアップロードファイル用）
fly volumes create poppo_data --region nrt --size 1

# 環境変数を設定
fly secrets set SUPABASE_JWT_SECRET=your-secret
fly secrets set OPENAI_API_KEY=your-key  # オプション
fly secrets set CORS_ORIGINS=https://poppo-frontend.fly.dev

# デプロイ
fly deploy
```

### 3. データベースパスの変更

`main.go` でデータベースとアップロードディレクトリを `/data` に変更する必要があります：

```go
// 変更前
db, err := sql.Open("sqlite3", "./poppo.db?_foreign_keys=on")

// 変更後
db, err := sql.Open("sqlite3", "/data/poppo.db?_foreign_keys=on")
```

### 4. フロントエンドのデプロイ

フロントエンドは別のアプリとしてデプロイするか、Vercel/Netlifyを使用することを推奨します。

## データベースパスの変更が必要

Fly.ioで永続ボリュームを使用する場合、`main.go` を修正する必要があります。

現在の設定では、データベースとアップロードファイルは `/data` ディレクトリに保存されます。

## コスト

- **完全無料**: クレジットカード不要
- **リソース制限**: 3つの共有CPU、256MB RAM
- **ボリューム**: 無料で使用可能

## トラブルシューティング

### ボリュームがマウントされない

```bash
# ボリュームの状態を確認
fly volumes list

# アプリにボリュームをアタッチ
fly volumes attach poppo_data
```

### 環境変数の確認

```bash
fly secrets list
```

