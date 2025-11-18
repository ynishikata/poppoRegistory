# 認証トークンエラーのデバッグ

## 問題

バックエンドログに以下のエラーが繰り返し表示される：
```
Auth error: no valid token found: check Authorization header or login status
API Error [401]: 認証トークンが見つかりません。ログインしてください。
```

## 原因の可能性

### 1. フロントエンドの環境変数が設定されていない

**確認方法:**
1. ブラウザの開発者ツール（F12）を開く
2. Consoleタブで以下を確認：
   - `VITE_SUPABASE_URL` が設定されているか
   - `VITE_SUPABASE_ANON_KEY` が設定されているか
   - エラーメッセージが表示されていないか

**解決方法:**
1. Render Dashboardでフロントエンドサービスを選択
2. "Environment" タブを開く
3. 以下の環境変数を確認・設定：
   - `VITE_SUPABASE_URL`: SupabaseプロジェクトのURL
   - `VITE_SUPABASE_ANON_KEY`: Supabase Anon Key
4. サービスを再デプロイ

### 2. Supabaseのセッションが取得できていない

**確認方法:**
1. ブラウザの開発者ツール（F12）を開く
2. Consoleタブで以下を確認：
   - "No active session. Please log in." というメッセージ
   - "Session exists but no access_token found" というメッセージ

**解決方法:**
1. 一度ログアウトして、再度ログイン
2. ブラウザのApplicationタブでLocal Storageを確認
   - `sb-<project-id>-auth-token` が存在するか確認
3. セッションが保存されていない場合は、Supabaseの設定を確認

### 3. フロントエンドとバックエンドのドメインが異なる

**確認:**
- フロントエンド: `https://poppo-frontend.onrender.com`
- バックエンド: `https://poppo-backend.onrender.com`

**問題:**
- 異なるドメイン間でセッションが共有されない
- SupabaseのセッションはLocal Storageに保存されるため、ドメインごとに独立

**解決方法:**
- これは正常な動作です
- フロントエンドからバックエンドにAPIリクエストを送信する際に、Authorizationヘッダーにトークンを含める必要があります

### 4. CORS設定の問題

**確認:**
- バックエンドのログに `CORS allowed origins: [https://poppo-frontend.onrender.com]` が表示されている
- これは正しく設定されています

## デバッグ手順

### ステップ1: ブラウザのコンソールを確認

1. ブラウザでF12キーを押して開発者ツールを開く
2. Consoleタブで以下を確認：
   - Supabaseの設定エラー
   - セッション取得エラー
   - トークン取得エラー

### ステップ2: ネットワークタブを確認

1. Networkタブを開く
2. `/api/plushies` リクエストを確認
3. Request Headersを確認：
   - `Authorization: Bearer <token>` が含まれているか
   - 含まれていない場合は、トークンが取得できていない

### ステップ3: Applicationタブを確認

1. Applicationタブを開く
2. Local Storageを確認
3. `sb-<project-id>-auth-token` が存在するか確認
4. 存在する場合は、その値を確認（JSON形式）

### ステップ4: 環境変数を確認

Render Dashboardで：
1. フロントエンドサービス → "Environment" タブ
2. 以下の環境変数が設定されているか確認：
   - `VITE_SUPABASE_URL`
   - `VITE_SUPABASE_ANON_KEY`
   - `VITE_API_BASE`

## よくある解決方法

### 方法1: 環境変数を再設定

1. Render Dashboardでフロントエンドサービスを選択
2. "Environment" タブを開く
3. 環境変数を削除して再追加
4. サービスを再デプロイ

### 方法2: ログアウトして再ログイン

1. フロントエンドでログアウト
2. ブラウザのキャッシュをクリア
3. 再度ログイン

### 方法3: Supabaseの設定を確認

1. Supabase Dashboardでプロジェクトを確認
2. Settings > API で以下を確認：
   - Project URL
   - anon public key
3. これらがRenderの環境変数と一致しているか確認

## 確認チェックリスト

- [ ] `VITE_SUPABASE_URL` が設定されている
- [ ] `VITE_SUPABASE_ANON_KEY` が設定されている
- [ ] `VITE_API_BASE` が設定されている（バックエンドのURL）
- [ ] ブラウザのコンソールにSupabaseの設定エラーがない
- [ ] ログインしている
- [ ] Local Storageにセッションが保存されている
- [ ] ネットワークタブでAuthorizationヘッダーが送信されている

