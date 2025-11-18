# デプロイ方法（統合版：フロントエンドをバックエンドから配信）

この方法では、フロントエンドをバックエンドの同じサービスから配信します。
1つのサービスでフルスタックアプリケーションとして動作します。

## メリット

- 1つのサービスで管理できる
- CORS設定が不要
- デプロイが簡単

## デプロイ手順

### 1. フロントエンドをビルド

```bash
cd frontend
npm install
npm run build
```

これにより `frontend/dist` ディレクトリが作成されます。

### 2. RenderでWebサービスを作成

1. **新規Webサービスを作成**
   - Render Dashboard → "New +" → "Web Service"
   - GitHubリポジトリを接続

2. **設定**
   - **Name**: `poppo-app`
   - **Environment**: `Go`
   - **Region**: `Oregon`（または最寄りのリージョン）
   - **Branch**: `main`
   - **Root Directory**: （空白、ルートディレクトリ）
   - **Build Command**: 
     ```bash
     cd frontend && npm install && npm run build && cd .. && go mod download && go build -o bin/server .
     ```
   - **Start Command**: `./bin/server`

3. **環境変数を設定**
   - `SUPABASE_JWT_SECRET`: （必須）
   - `OPENAI_API_KEY`: （オプション、会話機能を使用する場合）
   - `PORT`: `8080`
   - `VITE_SUPABASE_URL`: （フロントエンドビルド時に必要）
   - `VITE_SUPABASE_ANON_KEY`: （フロントエンドビルド時に必要）
   - `VITE_API_BASE`: `/api`（相対パスでOK）

4. **Diskを追加**（データ永続化のため）
   - "Disks" タブ → "Add Disk"
   - **Name**: `poppo-data`
   - **Mount Path**: `/opt/render/project/src/data`
   - **Size**: 1GB

5. **デプロイ**
   - "Create Web Service" をクリック
   - ビルドとデプロイが開始されます

## 注意事項

- フロントエンドのビルド時に環境変数が必要です
- ビルドコマンドが長くなるため、`build.sh` スクリプトを作成することを推奨します

## build.sh スクリプトの作成

`build.sh` ファイルを作成:

```bash
#!/bin/bash
set -e

# Build frontend
cd frontend
npm install
npm run build
cd ..

# Build backend
go mod download
go build -o bin/server .
```

そして、RenderのBuild Commandを `bash build.sh` に設定します。

