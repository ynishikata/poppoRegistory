# RenderのGitHubリポジトリ制限を変更する方法

RenderでGitHubと接続する際に、表示されるリポジトリを制限してしまった場合の変更方法です。

## 方法1: Render Dashboardから変更（推奨）

### 手順

1. **Render Dashboardにログイン**
   - [https://dashboard.render.com](https://dashboard.render.com) にアクセス

2. **アカウント設定を開く**
   - 右上のアカウント名をクリック
   - "Account Settings" を選択

3. **GitHub連携設定を確認**
   - 左メニューから "Connected Accounts" または "GitHub" を選択
   - 現在のGitHub連携状態を確認

4. **GitHubアプリの設定を変更**
   - "Manage GitHub App" または "Configure GitHub App" をクリック
   - または、GitHubアプリの設定ページに直接アクセス

5. **リポジトリアクセスの変更**
   - GitHubの設定ページで、以下のいずれかを選択：
     - **All repositories**: すべてのリポジトリにアクセス
     - **Only select repositories**: 特定のリポジトリのみにアクセス
   - 特定のリポジトリを選択する場合は、`ynishikata/poppoRegistory` を選択

6. **変更を保存**
   - "Save" または "Update" をクリック

## 方法2: GitHubから直接変更

### 手順

1. **GitHubにログイン**
   - [https://github.com](https://github.com) にアクセス

2. **Settingsを開く**
   - 右上のプロフィール画像をクリック
   - "Settings" を選択

3. **Applicationsを開く**
   - 左メニューから "Applications" → "Installed GitHub Apps" を選択

4. **Renderアプリを探す**
   - "Render" を検索または一覧から選択

5. **設定を変更**
   - "Repository access" セクションで設定を変更：
     - **All repositories**: すべてのリポジトリ
     - **Only select repositories**: 特定のリポジトリのみ
   - 特定のリポジトリを選択する場合は、`poppoRegistory` にチェック

6. **変更を保存**
   - "Save" をクリック

## 方法3: GitHubアプリを再インストール

上記の方法で変更できない場合：

1. **Render DashboardでGitHub連携を解除**
   - Account Settings → Connected Accounts
   - GitHub連携を解除

2. **GitHubでRenderアプリを削除**
   - GitHub Settings → Applications → Installed GitHub Apps
   - Renderアプリを削除

3. **Render Dashboardで再連携**
   - "New +" → "Blueprint" または "Web Service"
   - "Connect GitHub" をクリック
   - リポジトリアクセスを "All repositories" または "Only select repositories" で選択
   - `ynishikata/poppoRegistory` を選択

## 確認方法

設定変更後、Render Dashboardで：

1. "New +" → "Blueprint" または "Web Service" をクリック
2. GitHubリポジトリを選択
3. `ynishikata/poppoRegistory` が表示されることを確認

## トラブルシューティング

### リポジトリが表示されない

- GitHubアプリの権限を "All repositories" に変更
- または、特定のリポジトリを選択する場合は、`poppoRegistory` にチェックが入っているか確認

### 権限エラーが発生する

- GitHubアプリを再インストール
- Render DashboardでGitHub連携を解除して再設定

### 設定が反映されない

- ブラウザのキャッシュをクリア
- Render Dashboardを再読み込み
- 数分待ってから再度確認

