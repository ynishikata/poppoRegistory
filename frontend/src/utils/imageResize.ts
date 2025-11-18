/**
 * 画像をリサイズしてファイルサイズを削減
 * @param file 元の画像ファイル
 * @param maxWidth 最大幅（デフォルト: 1920px）
 * @param maxHeight 最大高さ（デフォルト: 1920px）
 * @param quality JPEG品質（0-1、デフォルト: 0.8）
 * @returns リサイズされたBlob
 */
export function resizeImage(
  file: File,
  maxWidth: number = 1280,  // Render無料プラン対応: より小さくリサイズ
  maxHeight: number = 1280,  // Render無料プラン対応: より小さくリサイズ
  quality: number = 0.75     // Render無料プラン対応: 品質を少し下げてファイルサイズを削減
): Promise<File> {
  return new Promise((resolve, reject) => {
    // 画像ファイルでない場合はそのまま返す
    if (!file.type.startsWith("image/")) {
      resolve(file);
      return;
    }

    // 2MB以下の場合はリサイズしない（既に小さい）
    // Renderの無料プランではタイムアウトが短いため、より小さなサイズに制限
    if (file.size <= 2 * 1024 * 1024) {
      resolve(file);
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();
      img.onload = () => {
        // リサイズが必要かチェック（1MB以下でリサイズ済みの場合はそのまま）
        if (img.width <= maxWidth && img.height <= maxHeight && file.size <= 1 * 1024 * 1024) {
          resolve(file);
          return;
        }

        // リサイズ後のサイズを計算
        let width = img.width;
        let height = img.height;

        if (width > height) {
          if (width > maxWidth) {
            height = Math.round((height * maxWidth) / width);
            width = maxWidth;
          }
        } else {
          if (height > maxHeight) {
            width = Math.round((width * maxHeight) / height);
            height = maxHeight;
          }
        }

        // Canvasでリサイズ
        const canvas = document.createElement("canvas");
        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext("2d");
        if (!ctx) {
          reject(new Error("Canvas context not available"));
          return;
        }

        ctx.drawImage(img, 0, 0, width, height);

        // Blobに変換
        canvas.toBlob(
          (blob) => {
            if (!blob) {
              reject(new Error("Failed to create blob"));
              return;
            }
            // Fileオブジェクトに変換（元のファイル名とタイプを保持）
            const resizedFile = new File([blob], file.name, {
              type: "image/jpeg",
              lastModified: Date.now(),
            });
            resolve(resizedFile);
          },
          "image/jpeg",
          quality
        );
      };
      img.onerror = () => {
        reject(new Error("Failed to load image"));
      };
      img.src = e.target?.result as string;
    };
    reader.onerror = () => {
      reject(new Error("Failed to read file"));
    };
    reader.readAsDataURL(file);
  });
}

