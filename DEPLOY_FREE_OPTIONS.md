# クレジットカード不要のデプロイ先比較

このアプリ（Goバックエンド + Reactフロントエンド）をクレジットカード不要でデプロイできるサービスを比較します。

## 推奨順位

### 1. Render（既に設定済み）⭐️ 最推奨

**メリット:**
- ✅ クレジットカード不要
- ✅ バックエンド（Go）とフロントエンド（React）の両方をホスト可能
- ✅ 既に `render.yaml` で設定済み
- ✅ 無料プランで2つのサービスまで作成可能
- ✅ ディスク永続化対応（SQLiteデータベース用）

**デメリット:**
- ⚠️ 無料プランでは15分間非アクティブでスリープ（起動に時間がかかる）
- ⚠️ ビルド時間が長い場合がある

**料金:** 完全無料（クレジットカード不要）

**設定済み:** ✅ `render.yaml` が既に作成済み

---

### 2. Fly.io

**メリット:**
- ✅ クレジットカード不要
- ✅ バックエンドとフロントエンドの両方をホスト可能
- ✅ スリープなし（無料プランでも）
- ✅ グローバルCDN対応

**デメリット:**
- ⚠️ 設定がやや複雑
- ⚠️ 無料プランはリソース制限あり（3つの共有CPU、256MB RAM）

**料金:** 完全無料（クレジットカード不要、$5のクレジット付き）

**設定:** `fly.toml` を作成する必要あり

---

### 3. Railway

**メリット:**
- ✅ クレジットカード不要（$5の無料クレジット付き）
- ✅ バックエンドとフロントエンドの両方をホスト可能
- ✅ デプロイが簡単
- ✅ 環境変数の管理が簡単

**デメリット:**
- ⚠️ 無料クレジットを使い切ると課金される可能性（ただしクレジットカード登録は不要）
- ⚠️ 無料クレジットは月額リセット

**料金:** $5の無料クレジット/月（クレジットカード不要）

**設定:** `railway.json` またはGUIで設定

---

### 4. Vercel + 別のバックエンドサービス

**メリット:**
- ✅ フロントエンドのデプロイが非常に簡単
- ✅ クレジットカード不要
- ✅ 高速なCDN
- ✅ 自動HTTPS

**デメリット:**
- ⚠️ フロントエンドのみ（バックエンドは別途必要）
- ⚠️ バックエンドを別サービス（Render、Fly.ioなど）でホストする必要

**料金:** 完全無料（クレジットカード不要）

**設定:** `vercel.json` を作成する必要あり

---

### 5. Netlify

**メリット:**
- ✅ フロントエンドのデプロイが簡単
- ✅ クレジットカード不要
- ✅ 自動HTTPS
- ✅ フォーム機能など追加機能あり

**デメリット:**
- ⚠️ フロントエンドのみ（バックエンドは別途必要）
- ⚠️ バックエンドを別サービスでホストする必要

**料金:** 完全無料（クレジットカード不要）

**設定:** `netlify.toml` を作成する必要あり

---

## 推奨構成

### 構成1: Render（シンプル・推奨）

- **バックエンド**: Render Web Service（Go）
- **フロントエンド**: Render Web Service（Node.js）
- **メリット**: 1つのサービスで完結、既に設定済み
- **設定**: `render.yaml` を使用

### 構成2: Fly.io（スリープなし）

- **バックエンド**: Fly.io App（Go）
- **フロントエンド**: Fly.io App（Node.js）
- **メリット**: スリープなし、高速
- **設定**: `fly.toml` を作成

### 構成3: Vercel + Render（ハイブリッド）

- **フロントエンド**: Vercel（React）
- **バックエンド**: Render（Go）
- **メリット**: フロントエンドが高速、バックエンドはRenderで管理
- **設定**: 両方の設定ファイルが必要

---

## 各サービスの設定ファイル例

### Fly.io (`fly.toml`)

```toml
# fly.toml for backend
app = "poppo-backend"
primary_region = "nrt"

[build]
  builder = "paketobuildpacks/builder:base"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = false
  auto_start_machines = true
  min_machines_running = 0
```

### Vercel (`vercel.json`)

```json
{
  "buildCommand": "cd frontend && npm install && npm run build",
  "outputDirectory": "frontend/dist",
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ]
}
```

### Railway (`railway.json`)

```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "NIXPACKS"
  },
  "deploy": {
    "startCommand": "./bin/server",
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

---

## 結論

**最も簡単で推奨される方法: Render**

- ✅ 既に設定ファイル（`render.yaml`）が作成済み
- ✅ クレジットカード不要
- ✅ バックエンドとフロントエンドの両方をホスト可能
- ✅ 無料プランで十分

**スリープを避けたい場合: Fly.io**

- ✅ スリープなし
- ✅ クレジットカード不要
- ⚠️ 設定がやや複雑

現在の `render.yaml` 設定で、そのままRenderにデプロイできます！

