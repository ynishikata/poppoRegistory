# フロントエンドデプロイのトラブルシューティング

## よくあるエラーと解決方法

### 1. 環境変数が設定されていない

**エラーメッセージ:**
```
VITE_SUPABASE_URL is not defined
```

**解決方法:**
- Render Dashboardでフロントエンドサービスの環境変数を確認
- `VITE_SUPABASE_URL` と `VITE_SUPABASE_ANON_KEY` が設定されているか確認
- 環境変数は**ビルド時**に必要です（Viteはビルド時に環境変数を埋め込みます）

### 2. serveパッケージが見つからない

**エラーメッセージ:**
```
command not found: serve
```

**解決方法:**
- `npx serve` を使用しているため、通常は問題ありません
- もしエラーが出る場合は、`package.json` に `serve` を追加：
  ```json
  "devDependencies": {
    "serve": "^14.2.0"
  }
  ```

### 3. ビルドエラー

**エラーメッセージ:**
```
Build failed
```

**解決方法:**
- Render Dashboardのログを確認
- ローカルで `cd frontend && npm install && npm run build` を実行してエラーを確認
- TypeScriptエラーの場合は修正

### 4. ポートエラー

**エラーメッセージ:**
```
Port 3000 is already in use
```

**解決方法:**
- `render.yaml` で `$PORT` 環境変数を使用するように修正済み
- Renderが自動的にポートを割り当てます

## 環境変数の設定手順

1. **Render Dashboardでフロントエンドサービスを選択**
2. **"Environment" タブを開く**
3. **以下の環境変数を追加:**

   ```
   VITE_SUPABASE_URL=https://your-project.supabase.co
   VITE_SUPABASE_ANON_KEY=your-anon-key
   VITE_API_BASE=https://poppo-backend.onrender.com/api
   ```

4. **"Save Changes" をクリック**
5. **サービスを再デプロイ**

## ビルドログの確認方法

1. Render Dashboardでフロントエンドサービスを選択
2. "Logs" タブを開く
3. ビルドログを確認
4. エラーメッセージを特定

## ローカルでのビルドテスト

デプロイ前にローカルでビルドをテスト：

```bash
cd frontend

# 環境変数を設定
export VITE_SUPABASE_URL=https://your-project.supabase.co
export VITE_SUPABASE_ANON_KEY=your-anon-key
export VITE_API_BASE=http://localhost:8080/api

# ビルド
npm install
npm run build

# ビルドが成功するか確認
```

## チェックリスト

デプロイ前に以下を確認：

- [ ] `VITE_SUPABASE_URL` が設定されている
- [ ] `VITE_SUPABASE_ANON_KEY` が設定されている
- [ ] `VITE_API_BASE` が設定されている（またはデフォルト値で問題ない）
- [ ] ローカルでビルドが成功する
- [ ] `package.json` の依存関係が最新

