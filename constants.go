package main

import "time"

// Error messages (Japanese)
const (
	ErrAuthRequired          = "認証が必要です。ログインしてください。"
	ErrInvalidID             = "無効なIDです"
	ErrPlushieNotFound       = "ぬいぐるみが見つかりませんでした"
	ErrFailedToListPlushies  = "ぬいぐるみ一覧の取得に失敗しました"
	ErrFailedToGetPlushie    = "ぬいぐるみ情報の取得に失敗しました"
	ErrFailedToCreatePlushie = "データベースへの保存に失敗しました"
	ErrFailedToUpdatePlushie = "ぬいぐるみ情報の更新に失敗しました"
	ErrFailedToDeletePlushie = "ぬいぐるみの削除に失敗しました"
	ErrFailedToUpdateConv    = "会話履歴の更新に失敗しました"
	ErrFailedToChat          = "チャットの生成に失敗しました"
	ErrFailedToParseForm     = "フォームデータの解析に失敗しました"
	ErrFailedToSaveImage     = "画像の保存に失敗しました"
	ErrNameRequired           = "名前は必須です"
	ErrUserNotFound           = "ユーザー情報が見つかりません。再度ログインしてください。"
	ErrDataReadFailed         = "データの読み込みに失敗しました"
	ErrServerConfigError      = "サーバー設定エラー: SUPABASE_JWT_SECRETが設定されていません"
	ErrTokenExpired           = "トークンの有効期限が切れています。再度ログインしてください。"
	ErrTokenNotFound          = "認証トークンが見つかりません。ログインしてください。"
	ErrAuthFailed             = "認証に失敗しました"
	ErrOpenAIKeyNotSet        = "OPENAI_API_KEY not configured"
)

// Configuration constants
const (
	DefaultMaxUsers     = 3
	MaxMultipartFormSize = 10 << 20 // 10MB
	DefaultReadTimeout  = 15 * time.Second
	DefaultWriteTimeout = 15 * time.Second
	DefaultPort         = ":8080"
	UploadsDir          = "uploads"
	DBPath              = "./poppo.db"
)

