# ファイルアップロードのタイムアウト対策

## 問題

大きな画像ファイルをアップロードする際に、応答が返ってくる前にタイムアウトが発生する。

## 原因

1. **サーバーのタイムアウト設定が短い**
   - `ReadTimeout`: 15秒
   - `WriteTimeout`: 15秒
   - 大きなファイルのアップロードには不十分

2. **Renderの無料プランの制限**
   - リクエストタイムアウトが30秒程度に制限されている可能性
   - 大きなファイルのアップロードには時間がかかる

3. **ファイルサイズ**
   - 現在の最大サイズ: 10MB
   - 高解像度の画像は数MBになることがある

## 実施した対策

### 1. タイムアウトの延長

`main.go`でタイムアウトを延長：
- `ReadTimeout`: 15秒 → 60秒
- `WriteTimeout`: 15秒 → 60秒
- `IdleTimeout`: 120秒（新規追加）

### 2. ファイルサイズ制限の維持

- 最大ファイルサイズ: 10MB（変更なし）
- これは適切なサイズです

## 追加の対策（推奨）

### 1. 画像のリサイズ（クライアント側）

フロントエンドで画像をリサイズしてからアップロード：

```typescript
// 画像をリサイズする関数（例）
function resizeImage(file: File, maxWidth: number, maxHeight: number): Promise<Blob> {
  return new Promise((resolve) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        const canvas = document.createElement('canvas');
        let width = img.width;
        let height = img.height;
        
        if (width > height) {
          if (width > maxWidth) {
            height *= maxWidth / width;
            width = maxWidth;
          }
        } else {
          if (height > maxHeight) {
            width *= maxHeight / height;
            height = maxHeight;
          }
        }
        
        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext('2d');
        ctx?.drawImage(img, 0, 0, width, height);
        canvas.toBlob(resolve, 'image/jpeg', 0.8);
      };
      img.src = e.target?.result as string;
    };
    reader.readAsDataURL(file);
  });
}
```

### 2. 画像のリサイズ（サーバー側）

バックエンドで画像をリサイズして保存（将来的な改善）：

- Goの画像処理ライブラリ（例: `github.com/disintegration/imaging`）を使用
- アップロード時に自動的にリサイズ

### 3. プログレスバーの表示

フロントエンドでアップロードの進捗を表示：

```typescript
const xhr = new XMLHttpRequest();
xhr.upload.addEventListener('progress', (e) => {
  if (e.lengthComputable) {
    const percentComplete = (e.loaded / e.total) * 100;
    // プログレスバーを更新
  }
});
```

### 4. チャンクアップロード（将来的な改善）

大きなファイルを小さなチャンクに分割してアップロード。

## Renderの制限について

Renderの無料プランでは：
- リクエストタイムアウト: 約30秒（公式ドキュメントに明記されていないが、一般的な制限）
- 大きなファイルのアップロードには不十分な場合がある

**推奨:**
- 画像をリサイズしてからアップロード（クライアント側で実装）
- 最大サイズを5MB程度に制限
- または、有料プランにアップグレード

## 確認方法

1. タイムアウトが解消されたか確認
2. 小さな画像（1MB以下）でテスト
3. 大きな画像（5MB以上）でテスト
4. ブラウザの開発者ツールでネットワークタブを確認

## トラブルシューティング

### まだタイムアウトが発生する場合

1. **画像サイズを確認**
   - ブラウザの開発者ツールでファイルサイズを確認
   - 10MBを超えている場合は、リサイズが必要

2. **ネットワーク速度を確認**
   - 遅いネットワークでは、タイムアウトが発生しやすい

3. **Renderのログを確認**
   - バックエンドサービスのログでエラーを確認

4. **画像をリサイズ**
   - フロントエンドで画像をリサイズしてからアップロード

