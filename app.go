package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	DB           *sql.DB
	SessionStore *SessionStore
}

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never returned
	CreatedAt time.Time `json:"created_at"`
}

type Plushie struct {
	ID                  int64     `json:"id"`
	UserID              string    `json:"-"` // Changed to string (UUID) for Supabase
	Name                string    `json:"name"`
	Kind                string    `json:"kind"`
	AdoptedAt           string    `json:"adopted_at"` // ISO8601 (yyyy-mm-dd)
	ImageURL            string    `json:"image_url"`
	ConversationHistory string    `json:"conversation_history"`
	CreatedAt           time.Time `json:"created_at"`
	ModifiedAt          time.Time `json:"modified_at"`
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	log.Printf("API Error [%d]: %s", status, msg)
	respondJSON(w, status, map[string]string{"error": msg})
}

// AuthMiddleware ensures user is logged in (Supabase JWT)
func (a *App) AuthMiddleware(next http.Handler) http.Handler {
	return a.SupabaseAuthMiddleware(next)
}

// HandleRegister creates a new user (email + password)
func (a *App) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// Check user limit (default: 3, configurable via MAX_USERS env var)
	maxUsers := 3
	if maxUsersStr := os.Getenv("MAX_USERS"); maxUsersStr != "" {
		if parsed, err := strconv.Atoi(maxUsersStr); err == nil && parsed > 0 {
			maxUsers = parsed
		}
	}

	// Count existing users
	var count int
	err := a.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to check user count")
		return
	}

	if count >= maxUsers {
		respondError(w, http.StatusForbidden, fmt.Sprintf("user registration is disabled. maximum users (%d) reached", maxUsers))
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password required")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	now := time.Now().UTC()
	res, err := a.DB.Exec(`INSERT INTO users (email, password_hash, created_at) VALUES (?, ?, ?)`, req.Email, string(hashed), now)
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to create user (maybe email already used)")
		return
	}
	id, _ := res.LastInsertId()

	respondJSON(w, http.StatusCreated, map[string]any{
		"id":    id,
		"email": req.Email,
	})
}

// HandleLogin authenticates user and sets session cookie
func (a *App) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	var id int64
	var hashed string
	err := a.DB.QueryRow(`SELECT id, password_hash FROM users WHERE email = ?`, req.Email).Scan(&id, &hashed)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token := uuid.NewString()
	a.SessionStore.Set(token, id, time.Now().Add(7*24*time.Hour))

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure:   true, // enable in production with https
	})

	respondJSON(w, http.StatusOK, map[string]any{
		"id":    id,
		"email": req.Email,
	})
}

func (a *App) HandleLogout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookieName)
	if err == nil {
		a.SessionStore.Delete(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID := supabaseUserIDFromContext(r.Context())
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get user info from JWT token
	supabaseAuth := NewSupabaseAuth()
	authHeader := r.Header.Get("Authorization")
	var email string
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token, err := jwt.ParseWithClaims(parts[1], &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(supabaseAuth.JWTSecret), nil
			})
			if err == nil {
				if claims, ok := token.Claims.(*SupabaseClaims); ok {
					email = claims.Email
					// Ensure user exists in users table with supabase_user_id
					_, _ = a.DB.Exec(`
						INSERT OR IGNORE INTO users (supabase_user_id, email, password_hash, created_at)
						VALUES (?, ?, '', datetime('now'))
					`, userID, email)
					// Update email if user exists
					_, _ = a.DB.Exec(`
						UPDATE users SET email = ? WHERE supabase_user_id = ?
					`, email, userID)
					respondJSON(w, http.StatusOK, map[string]any{
						"id":    claims.Sub,
						"email": claims.Email,
					})
					return
				}
			}
		}
	}

	// Fallback: return user ID only
	respondJSON(w, http.StatusOK, map[string]any{
		"id":    userID,
		"email": email,
	})
}

func (a *App) HandleListPlushies(w http.ResponseWriter, r *http.Request) {
	supabaseUserID := supabaseUserIDFromContext(r.Context())
	if supabaseUserID == "" {
		respondError(w, http.StatusUnauthorized, "認証が必要です。ログインしてください。")
		return
	}
	// Use Supabase UUID directly (no conversion needed)
	userID := supabaseUserID
	rows, err := a.DB.Query(`
		SELECT id, user_id, name, kind, adopted_at, image_path, created_at, updated_at
		FROM plushies
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ぬいぐるみ一覧の取得に失敗しました")
		return
	}
	defer rows.Close()

	var items []Plushie
	for rows.Next() {
		var p Plushie
		var adoptedAt sql.NullString
		var imagePath sql.NullString
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Kind, &adoptedAt, &imagePath, &p.CreatedAt, &p.ModifiedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "データの読み込みに失敗しました")
			return
		}
		if adoptedAt.Valid {
			p.AdoptedAt = adoptedAt.String
		}
		if imagePath.Valid {
			p.ImageURL = "/uploads/" + imagePath.String
		}
		items = append(items, p)
	}

	respondJSON(w, http.StatusOK, items)
}

func (a *App) HandleGetPlushie(w http.ResponseWriter, r *http.Request) {
	supabaseUserID := supabaseUserIDFromContext(r.Context())
	if supabaseUserID == "" {
		respondError(w, http.StatusUnauthorized, "認証が必要です。ログインしてください。")
		return
	}
	// Use Supabase UUID directly (no conversion needed)
	userID := supabaseUserID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "無効なIDです")
		return
	}

	var p Plushie
	var adoptedAt sql.NullString
	var imagePath sql.NullString
	var conversationHistory sql.NullString
	err = a.DB.QueryRow(`
		SELECT id, user_id, name, kind, adopted_at, image_path, conversation_history, created_at, updated_at
		FROM plushies
		WHERE id = ? AND user_id = ?
	`, id, userID).Scan(&p.ID, &p.UserID, &p.Name, &p.Kind, &adoptedAt, &imagePath, &conversationHistory, &p.CreatedAt, &p.ModifiedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "ぬいぐるみが見つかりませんでした")
		} else {
			respondError(w, http.StatusInternalServerError, "ぬいぐるみ情報の取得に失敗しました")
		}
		return
	}
	if adoptedAt.Valid {
		p.AdoptedAt = adoptedAt.String
	}
	if imagePath.Valid {
		p.ImageURL = "/uploads/" + imagePath.String
	}
	if conversationHistory.Valid {
		p.ConversationHistory = conversationHistory.String
	}

	respondJSON(w, http.StatusOK, p)
}

func (a *App) HandleCreatePlushie(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC in HandleCreatePlushie: %v", err)
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("internal server error: %v", err))
		}
	}()

	supabaseUserID := supabaseUserIDFromContext(r.Context())
	if supabaseUserID == "" {
		log.Printf("ERROR: HandleCreatePlushie - no user ID in context")
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Use Supabase UUID directly (no conversion needed)
	userID := supabaseUserID

	// Ensure user exists in users table for foreign key constraint
	// Get email from JWT token if available
	supabaseAuth := NewSupabaseAuth()
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token, err := jwt.ParseWithClaims(parts[1], &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(supabaseAuth.JWTSecret), nil
			})
			if err == nil {
				if claims, ok := token.Claims.(*SupabaseClaims); ok {
					// Ensure user exists in users table
					_, _ = a.DB.Exec(`
						INSERT OR IGNORE INTO users (supabase_user_id, email, password_hash, created_at)
						VALUES (?, ?, '', datetime('now'))
					`, userID, claims.Email)
				}
			}
		}
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		log.Printf("ERROR: HandleCreatePlushie - failed to parse form: %v", err)
		respondError(w, http.StatusBadRequest, "フォームデータの解析に失敗しました")
		return
	}

	name := r.FormValue("name")
	kind := r.FormValue("kind")
	adoptedAt := r.FormValue("adopted_at")
	if name == "" {
		log.Printf("ERROR: HandleCreatePlushie - name is empty")
		respondError(w, http.StatusBadRequest, "名前は必須です")
		return
	}

	imagePath, err := saveUploadedFile(r, "image")
	if err != nil && !errors.Is(err, ErrNoFile) {
		log.Printf("ERROR: HandleCreatePlushie - failed to save image: %v", err)
		respondError(w, http.StatusBadRequest, fmt.Sprintf("画像の保存に失敗しました: %v", err))
		return
	}

	now := time.Now().UTC()
	res, err := a.DB.Exec(`
		INSERT INTO plushies (user_id, name, kind, adopted_at, image_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, userID, name, kind, nullIfEmpty(adoptedAt), nullIfEmpty(imagePath), now, now)
	if err != nil {
		log.Printf("ERROR: HandleCreatePlushie - failed to insert plushie: %v", err)
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			respondError(w, http.StatusBadRequest, "ユーザー情報が見つかりません。再度ログインしてください。")
		} else {
			respondError(w, http.StatusInternalServerError, "データベースへの保存に失敗しました")
		}
		return
	}
	id, _ := res.LastInsertId()
	respondJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (a *App) HandleUpdatePlushie(w http.ResponseWriter, r *http.Request) {
	supabaseUserID := supabaseUserIDFromContext(r.Context())
	if supabaseUserID == "" {
		respondError(w, http.StatusUnauthorized, "認証が必要です。ログインしてください。")
		return
	}
	// Use Supabase UUID directly (no conversion needed)
	userID := supabaseUserID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "無効なIDです")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "フォームデータの解析に失敗しました")
		return
	}
	name := r.FormValue("name")
	kind := r.FormValue("kind")
	adoptedAt := r.FormValue("adopted_at")

	// ensure it belongs to this user
	var existingImage sql.NullString
	err = a.DB.QueryRow(`SELECT image_path FROM plushies WHERE id = ? AND user_id = ?`, id, userID).Scan(&existingImage)
	if err != nil {
		respondError(w, http.StatusNotFound, "plushie not found")
		return
	}

	imagePath := ""
	filePath, err := saveUploadedFile(r, "image")
	if err != nil && !errors.Is(err, ErrNoFile) {
		respondError(w, http.StatusBadRequest, "画像の保存に失敗しました")
		return
	}
	if filePath != "" {
		imagePath = filePath
	} else if existingImage.Valid {
		imagePath = existingImage.String
	}

	_, err = a.DB.Exec(`
		UPDATE plushies
		SET name = ?, kind = ?, adopted_at = ?, image_path = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`, name, kind, nullIfEmpty(adoptedAt), nullIfEmpty(imagePath), time.Now().UTC(), id, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ぬいぐるみ情報の更新に失敗しました")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) HandleDeletePlushie(w http.ResponseWriter, r *http.Request) {
	supabaseUserID := supabaseUserIDFromContext(r.Context())
	if supabaseUserID == "" {
		respondError(w, http.StatusUnauthorized, "認証が必要です。ログインしてください。")
		return
	}
	// Use Supabase UUID directly (no conversion needed)
	userID := supabaseUserID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "無効なIDです")
		return
	}

	res, err := a.DB.Exec(`DELETE FROM plushies WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ぬいぐるみの削除に失敗しました")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		respondError(w, http.StatusNotFound, "plushie not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) HandleUpdateConversation(w http.ResponseWriter, r *http.Request) {
	supabaseUserID := supabaseUserIDFromContext(r.Context())
	if supabaseUserID == "" {
		respondError(w, http.StatusUnauthorized, "認証が必要です。ログインしてください。")
		return
	}
	// Use Supabase UUID directly (no conversion needed)
	userID := supabaseUserID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "無効なIDです")
		return
	}

	var req struct {
		ConversationHistory string `json:"conversation_history"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// ensure it belongs to this user
	var exists int
	err = a.DB.QueryRow(`SELECT 1 FROM plushies WHERE id = ? AND user_id = ?`, id, userID).Scan(&exists)
	if err != nil {
		respondError(w, http.StatusNotFound, "plushie not found")
		return
	}

	_, err = a.DB.Exec(`
		UPDATE plushies
		SET conversation_history = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`, req.ConversationHistory, time.Now().UTC(), id, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "会話履歴の更新に失敗しました")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) HandleChat(w http.ResponseWriter, r *http.Request) {
	supabaseUserID := supabaseUserIDFromContext(r.Context())
	if supabaseUserID == "" {
		respondError(w, http.StatusUnauthorized, "認証が必要です。ログインしてください。")
		return
	}
	// Use Supabase UUID directly (no conversion needed)
	userID := supabaseUserID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "無効なIDです")
		return
	}

	// Get plushie details
	var name, kind string
	var conversationHistory sql.NullString
	err = a.DB.QueryRow(`
		SELECT name, kind, conversation_history
		FROM plushies
		WHERE id = ? AND user_id = ?
	`, id, userID).Scan(&name, &kind, &conversationHistory)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "ぬいぐるみが見つかりませんでした")
		} else {
			respondError(w, http.StatusInternalServerError, "ぬいぐるみ情報の取得に失敗しました")
		}
		return
	}

	// Call OpenAI API
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		respondError(w, http.StatusInternalServerError, "OPENAI_API_KEY not configured")
		return
	}

	history := ""
	if conversationHistory.Valid {
		history = conversationHistory.String
	}

	prompt := buildChatPrompt(name, kind, history)
	message, err := callOpenAI(apiKey, prompt)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "チャットの生成に失敗しました: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": message})
}

func buildChatPrompt(name, kind, history string) string {
	prompt := fmt.Sprintf("あなたは「%s」という名前の%sのぬいぐるみです。", name, kind)
	if history != "" {
		prompt += fmt.Sprintf("\n\n過去の会話履歴:\n%s\n\n", history)
	}
	prompt += "このぬいぐるみのキャラクターとして、短い一言（1〜2文程度）を話してください。親しみやすく、温かみのある言葉を選んでください。"
	return prompt
}

func callOpenAI(apiKey, prompt string) (string, error) {
	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type Request struct {
		Model     string    `json:"model"`
		Messages  []Message `json:"messages"`
		MaxTokens int       `json:"max_tokens"`
	}

	reqBody := Request{
		Model: "gpt-4o-mini",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		MaxTokens: 100,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	type Choice struct {
		Message Message `json:"message"`
	}
	type Response struct {
		Choices []Choice `json:"choices"`
	}

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return apiResp.Choices[0].Message.Content, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
