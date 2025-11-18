# deployブランチでのデプロイ手順

## 方法1: render.yamlを使用した自動デプロイ（推奨）

### 1. Render DashboardでBlueprintを作成

1. [Render Dashboard](https://dashboard.render.com) にログイン
2. "New +" → "Blueprint" を選択
3. GitHubリポジトリを選択: `ynishikata/poppoRegistory`
4. **重要**: ブランチを `deploy` に設定
   - "Branch" フィールドに `deploy` を入力
5. `render.yaml` が自動的に検出されます

### 2. 環境変数を設定

Blueprint作成時に、各サービスの環境変数を設定します。

#### バックエンドサービス（poppo-backend）

**必須:**
- `SUPABASE_JWT_SECRET`: Supabase Dashboard (Settings > API > JWT Secret) から取得
- `PORT`: `8080`

**オプション:**
- `OPENAI_API_KEY`: 会話機能を使用する場合のみ
- `CORS_ORIGINS`: デプロイ後にフロントエンドURLを設定（例: `https://poppo-frontend.onrender.com`）

#### フロントエンドサービス（poppo-frontend）

**必須:**
- `VITE_SUPABASE_URL`: SupabaseプロジェクトのURL（例: `https://xxxxx.supabase.co`）
- `VITE_SUPABASE_ANON_KEY`: Supabase Anon Key
- `VITE_API_BASE`: デプロイ後にバックエンドURLを設定（例: `https://poppo-backend.onrender.com/api`）

### 3. デプロイを開始

1. "Apply" をクリック
2. 両方のサービスがビルド・デプロイされます
3. デプロイ完了まで数分かかります

### 4. デプロイ後の設定

1. **バックエンドとフロントエンドのURLを確認**
   - 各サービスのDashboardでURLを確認
   - 例: 
     - バックエンド: `https://poppo-backend.onrender.com`
     - フロントエンド: `https://poppo-frontend.onrender.com`

2. **環境変数を更新**
   
   **フロントエンドサービス:**
   - `VITE_API_BASE` を `https://poppo-backend.onrender.com/api` に更新
   - サービスを再デプロイ（"Manual Deploy" → "Deploy latest commit"）
   
   **バックエンドサービス:**
   - `CORS_ORIGINS` を `https://poppo-frontend.onrender.com` に更新
   - サービスを再デプロイ

## 方法2: 手動でサービスを作成（deployブランチ指定）

### バックエンドのデプロイ

1. **新規Webサービスを作成**
   - Render Dashboard → "New +" → "Web Service"
   - GitHubリポジトリを接続: `ynishikata/poppoRegistory`

2. **設定**
   - **Name**: `poppo-backend`
   - **Environment**: `Go`
   - **Region**: `Oregon`（または最寄りのリージョン）
   - **Branch**: `deploy` ⚠️ **重要: deployブランチを指定**
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

### フロントエンドのデプロイ

1. **新規Webサービスを作成**
   - Render Dashboard → "New +" → "Web Service"
   - 同じGitHubリポジトリを選択: `ynishikata/poppoRegistory`

2. **設定**
   - **Name**: `poppo-frontend`
   - **Environment**: `Node`
   - **Region**: `Oregon`（バックエンドと同じリージョン推奨）
   - **Branch**: `deploy` ⚠️ **重要: deployブランチを指定**
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

## 注意事項

- **ブランチ指定**: 必ず `deploy` ブランチを指定してください
- **環境変数**: デプロイ前にすべての必須環境変数を設定してください
- **CORS設定**: フロントエンドとバックエンドのURLが確定してから、相互に設定を更新してください
- **無料プラン**: Renderの無料プランでは、15分間非アクティブでサービスがスリープします

## トラブルシューティング

### ビルドエラー

- ブランチが `deploy` に設定されているか確認
- ログを確認してエラー内容を確認

### 環境変数エラー

- すべての必須環境変数が設定されているか確認
- 値に余分なスペースや引用符がないか確認

### CORSエラー

- バックエンドの `CORS_ORIGINS` にフロントエンドの正確なURLが設定されているか確認
- プロトコル（`https://`）を含めることを確認

