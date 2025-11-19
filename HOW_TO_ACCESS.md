# デプロイしたアプリケーションへのアクセス方法

## Renderでデプロイした場合

### 1. URLの確認

1. **Render Dashboardにログイン**
   - [https://dashboard.render.com](https://dashboard.render.com)

2. **サービス一覧を確認**
   - Dashboardのホーム画面に、デプロイしたサービスが表示されます
   - 例：
     - `poppo-backend`
     - `poppo-frontend`

3. **URLを確認**
   - 各サービスの名前をクリック
   - サービスの詳細ページで、上部にURLが表示されます
   - 例：
     - バックエンド: `https://poppo-backend.onrender.com`
     - フロントエンド: `https://poppo-frontend.onrender.com`

### 2. フロントエンドにアクセス

**フロントエンドのURLをブラウザで開く**

例: `https://poppo-frontend.onrender.com`

これがアプリケーションのメインURLです。

### 3. 初回アクセス時の注意

- **無料プランの場合**: サービスがスリープしている可能性があります
  - 15分間非アクティブでスリープします
  - 初回アクセス時に起動するまで30秒〜1分かかることがあります
  - しばらく待ってから再度アクセスしてください

### 4. 環境変数の確認

フロントエンドが正しく動作するためには、以下の環境変数が設定されている必要があります：

- `VITE_SUPABASE_URL`: SupabaseプロジェクトのURL
- `VITE_SUPABASE_ANON_KEY`: Supabase Anon Key
- `VITE_API_BASE`: バックエンドのURL（例: `https://poppo-backend.onrender.com/api`）

バックエンドが正しく動作するためには：

- `SUPABASE_JWT_SECRET`: Supabase JWT Secret
- `CORS_ORIGINS`: フロントエンドのURL（例: `https://poppo-frontend.onrender.com`）

## カスタムドメインの設定（オプション）

### Renderでカスタムドメインを設定

1. **サービスを選択**
2. **"Settings" タブを開く**
3. **"Custom Domains" セクションを開く**
4. **ドメインを追加**
   - 例: `poppo.yourdomain.com`
5. **DNS設定を更新**
   - Renderが提供するCNAMEレコードをDNSに追加

## トラブルシューティング

### ページが表示されない

1. **サービスが起動しているか確認**
   - Render Dashboardでサービスのステータスを確認
   - "Live" と表示されているか確認

2. **ログを確認**
   - "Logs" タブでエラーがないか確認

3. **環境変数が設定されているか確認**
   - "Environment" タブで必須の環境変数が設定されているか確認

### エラーページが表示される

1. **ブラウザの開発者ツールを開く**（F12キー）
2. **Consoleタブでエラーを確認**
3. **NetworkタブでAPIリクエストが失敗していないか確認**

### API接続エラー

- バックエンドのURLが正しく設定されているか確認
- CORS設定が正しいか確認（バックエンドの`CORS_ORIGINS`にフロントエンドのURLが含まれているか）

## アクセスURLの例

デプロイが成功した場合、以下のようなURLでアクセスできます：

```
フロントエンド: https://poppo-frontend.onrender.com
バックエンドAPI: https://poppo-backend.onrender.com/api
```

## 次のステップ

1. フロントエンドのURLをブラウザで開く
2. 新規登録またはログイン
3. ぬいぐるみを登録して動作確認

