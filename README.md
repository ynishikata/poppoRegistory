## ぬいぐるみレジストリ

家にいるぬいぐるみを「名前・種類・お迎え日・写真」で登録して管理する小さなWebアプリです。  
各ぬいぐるみに会話履歴を記録し、LLM APIを使って「一言話す」機能も利用できます。  
バックエンドは Go + SQLite、フロントエンドは React(Vite) で構成されています。

### 構成

- `main.go` ほか: Go API サーバー
  - 認証: Supabase Auth (JWT)
  - `/api/me` - 現在のユーザー情報取得
  - `/api/plushies` (GET/POST/PUT/DELETE) - ぬいぐるみCRUD
  - `/api/plushies/{id}` (GET) - ぬいぐるみ詳細取得
  - `/api/plushies/{id}/conversation` (PUT) - 会話履歴の更新
  - `/api/plushies/{id}/chat` (POST) - LLM APIを使った一言生成
  - `uploads/` ディレクトリに画像ファイルを保存
- `frontend/`: React フロントエンド (Vite + TypeScript + React Router)
  - Supabase Auth クライアントを使用

### 動かし方

#### 1. 環境変数の設定

プロジェクトルートに `.env` ファイルを作成し、以下の環境変数を設定してください:

```bash
# .env
SUPABASE_JWT_SECRET=your-supabase-jwt-secret
OPENAI_API_KEY=sk-...  # 会話機能を使う場合のみ
```

- `SUPABASE_JWT_SECRET`: Supabase Dashboard (Settings > API > JWT Secret) から取得
- `OPENAI_API_KEY`: [OpenAI Platform](https://platform.openai.com/api-keys) から取得（会話機能を使う場合のみ）

#### 2. バックエンド(Go)を起動

```bash
cd /Users/y/go/src/github.com/ynishikata/poppoRegistory
go mod tidy
go run .
```

- ポート `:8080` で起動します。
- 起動時に `poppo.db`(SQLite) と `uploads/` ディレクトリが作られます。
- **会話機能を使う場合**: `.env` ファイルに `OPENAI_API_KEY` を設定してください。
  - 設定しない場合、「話す」機能はエラーになりますが、その他の機能は正常に動作します。

#### 3. フロントエンド(React)を起動

別のターミナルで:

```bash
cd /Users/y/go/src/github.com/ynishikata/poppoRegistory/frontend
npm install   # or: pnpm install / yarn
```

`frontend/.env` ファイルを作成し、Supabaseの認証情報を設定してください:

```bash
# frontend/.env
VITE_SUPABASE_URL=https://your-project.supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key
```

Supabase Dashboard (Settings > API) から取得できます。

```bash
npm run dev
```

- 通常 `http://localhost:5173` で開発サーバーが立ちます。
- API へのアクセス先はデフォルトで `http://localhost:8080/api` を見るようになっています。

### 使い方

#### 基本的な使い方

1. ブラウザで `http://localhost:5173` を開く
2. メールアドレスとパスワードで「新規登録」→そのままログイン（Supabase Authを使用）
3. ログイン後、画面右側に「ぬいぐるみ一覧」、左側に「登録フォーム」が表示されます
4. 「名前」「種類」「お迎え日」「写真ファイル」を入力して登録
5. カード一覧から「編集」「削除」ができます

※ ログインしているユーザーごとに、ぬいぐるみ一覧は分かれています。
※ 認証は Supabase Auth を使用しています。

#### 会話機能の使い方

1. ぬいぐるみ一覧のカードから「詳細を見る」をクリック
2. 詳細ページで「会話履歴」テキストエリアに、そのぬいぐるみの性格や過去の会話内容を自由に記録
   - 例: 「こんにちは！今日はいい天気ですね。一緒に遊びましょう。」
3. 「会話履歴を保存」ボタンで保存
4. 「話す」ボタンを押すと、LLMが会話履歴を参考に、そのぬいぐるみのキャラクターとして短い一言を生成します

**注意**: 会話機能を使うには、バックエンド起動時に `OPENAI_API_KEY` 環境変数を設定する必要があります。

### 将来のPostgreSQL対応について

今は開発用として SQLite (`poppo.db`) を利用していますが、将来的に PostgreSQL へ移行しやすいように:

- SQL は基本的に SQLite / PostgreSQL どちらでも動く素朴な構文にしています
- 切り替える場合は:
  - `database/sql` の DSN を PostgreSQL 用に変更
  - `github.com/mattn/go-sqlite3` を PostgreSQL ドライバ(例: `github.com/jackc/pgx/v5/stdlib`)に変更
  - 必要に応じて `migrate.go` のテーブル定義を調整

### メモ

- **認証**: Supabase Auth を使用しています。JWT トークンで認証を行います。
- **セキュリティ**: 本番運用を想定する場合は、HTTPS + `Secure` Cookie、画像容量制限などを追加してください。
- **OpenAI API**: 会話機能は OpenAI API (gpt-4o-mini) を使用しています。APIキーは環境変数で管理し、リポジトリにコミットしないでください。
- **コスト**: OpenAI APIの使用には料金がかかります。詳細は [OpenAI Pricing](https://openai.com/api/pricing/) を参照してください。
- **エラーメッセージ**: エラーメッセージは日本語で表示されます。


