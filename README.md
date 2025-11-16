## ぬいぐるみレジストリ

家にいるぬいぐるみを「名前・種類・お迎え日・写真」で登録して管理する小さなWebアプリです。  
バックエンドは Go + SQLite、フロントエンドは React(Vite) で構成されています。

### 構成

- `main.go` ほか: Go API サーバー
  - `/api/register` / `/api/login` / `/api/logout`
  - `/api/me`
  - `/api/plushies` (GET/POST/PUT/DELETE)
  - `uploads/` ディレクトリに画像ファイルを保存
- `frontend/`: React フロントエンド (Vite + TypeScript)

### 動かし方

#### 1. バックエンド(G0)を起動

```bash
cd /Users/y/go/src/github.com/ynishikata/poppoRegistory
go mod tidy
go run .
```

- ポート `:8080` で起動します。
- 起動時に `poppo.db`(SQLite) と `uploads/` ディレクトリが作られます。

#### 2. フロントエンド(React)を起動

別のターミナルで:

```bash
cd /Users/y/go/src/github.com/ynishikata/poppoRegistory/frontend
npm install   # or: pnpm install / yarn
npm run dev
```

- 通常 `http://localhost:5173` で開発サーバーが立ちます。
- API へのアクセス先はデフォルトで `http://localhost:8080/api` を見るようになっています。
  - 変更したい場合は `.env` を作成して `VITE_API_BASE` を設定してください。

```bash
# frontend/.env
VITE_API_BASE=http://localhost:8080/api
```

### 使い方

1. ブラウザで `http://localhost:5173` を開く
2. メールアドレスとパスワードで「新規登録」→そのままログイン
3. ログイン後、画面右側に「ぬいぐるみ一覧」、左側に「登録フォーム」が表示されます
4. 「名前」「種類」「お迎え日」「写真ファイル」を入力して登録
5. カード一覧から削除ボタンで削除できます

※ ログインしているユーザーごとに、ぬいぐるみ一覧は分かれています。

### 将来のPostgreSQL対応について

今は開発用として SQLite (`poppo.db`) を利用していますが、将来的に PostgreSQL へ移行しやすいように:

- SQL は基本的に SQLite / PostgreSQL どちらでも動く素朴な構文にしています
- 切り替える場合は:
  - `database/sql` の DSN を PostgreSQL 用に変更
  - `github.com/mattn/go-sqlite3` を PostgreSQL ドライバ(例: `github.com/jackc/pgx/v5/stdlib`)に変更
  - 必要に応じて `migrate.go` のテーブル定義を調整

### メモ

- 本番運用を想定する場合は、セッションの永続化(メモリではなくRedisなど)と HTTPS + `Secure` Cookie、画像容量制限などを追加してください。


