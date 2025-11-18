# デプロイ手順（Render）

このドキュメントでは、Renderを使用してアプリケーションをデプロイする手順を説明します。

## 前提条件

- GitHubアカウント
- Renderアカウント（[https://render.com](https://render.com) で無料登録可能）
- Supabaseプロジェクト（認証用）
- OpenAI APIキー（会話機能を使用する場合）

## デプロイ方法

### 方法1: render.yamlを使用した自動デプロイ（推奨）

1. **GitHubリポジトリにpush**
   ```bash
   git add .
   git commit -m "Add deployment configuration"
   git push
   ```

2. **Render Dashboardで新規Blueprintを作成**
   - Render Dashboardにログイン
   - "New +" → "Blueprint" を選択
   - GitHubリポジトリを選択
   - `render.yaml` が自動的に検出されます

3. **環境変数を設定**
   
   **バックエンドサービス（poppo-backend）:**
   - `SUPABASE_JWT_SECRET`: Supabase Dashboard (Settings > API > JWT Secret)
   - `OPENAI_API_KEY`: OpenAI APIキー（会話機能を使用する場合）
   - `CORS_ORIGINS`: フロントエンドのURL（例: `https://poppo-frontend.onrender.com`）
   - `PORT`: `8080`（デフォルト）

   **フロントエンドサービス（poppo-frontend）:**
   - `VITE_SUPABASE_URL`: SupabaseプロジェクトのURL
   - `VITE_SUPABASE_ANON_KEY`: Supabase Anon Key
   - `VITE_API_BASE`: バックエンドのURL（例: `https://poppo-backend.onrender.com/api`）

4. **デプロイを開始**
   - "Apply" をクリック
   - 両方のサービスがビルド・デプロイされます

5. **デプロイ後の設定**
   - バックエンドとフロントエンドのURLを確認
   - フロントエンドの `VITE_API_BASE` をバックエンドのURLに更新
   - バックエンドの `CORS_ORIGINS` をフロントエンドのURLに更新
   - 両方のサービスを再デプロイ

### 方法2: 手動でサービスを作成

#### バックエンドのデプロイ

1. **新規Webサービスを作成**
   - Render Dashboard → "New +" → "Web Service"
   - GitHubリポジトリを接続

2. **設定**
   - **Name**: `poppo-backend`
   - **Environment**: `Go`
   - **Region**: `Oregon`（または最寄りのリージョン）
   - **Branch**: `main`
   - **Root Directory**: （空白、ルートディレクトリ）
   - **Build Command**: `go mod download && go build -o bin/server .`
   - **Start Command**: `./bin/server`

3. **環境変数を設定**
   - `SUPABASE_JWT_SECRET`: （必須）
   - `OPENAI_API_KEY`: （オプション、会話機能を使用する場合）
   - `PORT`: `8080`
   - `CORS_ORIGINS`: （後でフロントエンドURLを設定）

4. **Diskを追加**（データ永続化のため）
   - "Disks" タブ → "Add Disk"
   - **Name**: `poppo-data`
   - **Mount Path**: `/opt/render/project/src/data`
   - **Size**: 1GB

5. **デプロイ**
   - "Create Web Service" をクリック
   - ビルドとデプロイが開始されます
   - デプロイ完了後、URLをメモ（例: `https://poppo-backend.onrender.com`）

#### フロントエンドのデプロイ

1. **新規Webサービスを作成**
   - Render Dashboard → "New +" → "Web Service"
   - 同じGitHubリポジトリを選択

2. **設定**
   - **Name**: `poppo-frontend`
   - **Environment**: `Node`
   - **Region**: `Oregon`（バックエンドと同じリージョン推奨）
   - **Branch**: `main`
   - **Root Directory**: `frontend`
   - **Build Command**: `npm install && npm run build`
   - **Start Command**: `npx serve -s dist -l 3000`

3. **環境変数を設定**
   - `VITE_SUPABASE_URL`: （必須）
   - `VITE_SUPABASE_ANON_KEY`: （必須）
   - `VITE_API_BASE`: バックエンドのURL（例: `https://poppo-backend.onrender.com/api`）

4. **デプロイ**
   - "Create Web Service" をクリック
   - ビルドとデプロイが開始されます
   - デプロイ完了後、URLをメモ（例: `https://poppo-frontend.onrender.com`）

5. **CORS設定を更新**
   - バックエンドサービスの環境変数 `CORS_ORIGINS` を更新
   - フロントエンドのURLを設定（例: `https://poppo-frontend.onrender.com`）
   - バックエンドサービスを再デプロイ

## データベースとファイルストレージ

### SQLiteデータベース

Renderの無料プランでは、ディスクが永続化されますが、サービスがスリープする可能性があります。
本番環境では、PostgreSQLデータベースの使用を推奨します。

### 画像ファイル

現在、画像は `uploads/` ディレクトリに保存されます。Renderのディスクに保存されますが、
より堅牢なソリューションとして、S3やCloudinaryなどの外部ストレージサービスの使用を検討してください。

## トラブルシューティング

### ビルドエラー

- **Goビルドエラー**: `go.mod` と `go.sum` が最新であることを確認
- **Nodeビルドエラー**: `frontend/package-lock.json` が最新であることを確認

### 環境変数エラー

- すべての必須環境変数が設定されているか確認
- 環境変数の値に余分なスペースや引用符がないか確認

### CORSエラー

- バックエンドの `CORS_ORIGINS` にフロントエンドの正確なURLが設定されているか確認
- プロトコル（`https://`）を含めることを確認

### データベースエラー

- ディスクが正しくマウントされているか確認
- データベースファイルのパスが正しいか確認

## カスタムドメインの設定

1. Render Dashboardでサービスを選択
2. "Settings" → "Custom Domains"
3. ドメインを追加
4. DNS設定を更新（Renderが提供するCNAMEレコードを使用）

## モニタリングとログ

- Render Dashboardで各サービスのログを確認できます
- "Logs" タブでリアルタイムログを表示
- エラーが発生した場合、ログを確認して原因を特定

## コスト

- **無料プラン**: サービスが15分間非アクティブでスリープします
- **Starterプラン** ($7/月): スリープなし、より高速なビルド
- 詳細は [Render Pricing](https://render.com/pricing) を参照

## 次のステップ

- PostgreSQLデータベースへの移行を検討
- 画像ストレージにS3やCloudinaryを使用
- CI/CDパイプラインの設定
- カスタムドメインの設定

