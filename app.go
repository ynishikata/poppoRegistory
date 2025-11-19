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
	"strings"
	"time"

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
	maxUsers := DefaultMaxUsers
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
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	// Get user info from JWT token
	email, err := getEmailFromJWT(r)
	if err == nil && email != "" {
		// Ensure user exists in users table
		if err := a.ensureUserExists(userID, email); err != nil {
			log.Printf("WARNING: Failed to ensure user exists: %v", err)
		}
		respondJSON(w, http.StatusOK, map[string]any{
			"id":    userID,
			"email": email,
		})
		return
	}

	// Fallback: return user ID only
	respondJSON(w, http.StatusOK, map[string]any{
		"id":    userID,
		"email": "",
	})
}

func (a *App) HandleListPlushies(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	rows, err := a.DB.Query(`
		SELECT id, user_id, name, kind, adopted_at, image_path, conversation_history, created_at, updated_at
		FROM plushies
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, ErrFailedToListPlushies)
		return
	}
	defer rows.Close()

	var items []Plushie
	for rows.Next() {
		p, err := scanPlushieFromRow(rows)
		if err != nil {
			respondError(w, http.StatusInternalServerError, ErrDataReadFailed)
			return
		}
		items = append(items, *p)
	}

	respondJSON(w, http.StatusOK, items)
}

func (a *App) HandleGetPlushie(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	id, err := parsePlushieID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
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
			respondError(w, http.StatusNotFound, ErrPlushieNotFound)
		} else {
			respondError(w, http.StatusInternalServerError, ErrFailedToGetPlushie)
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

	userID, err := getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	// Ensure user exists in users table for foreign key constraint
	if err := a.ensureUserExistsFromRequest(r, userID); err != nil {
		log.Printf("WARNING: Failed to ensure user exists: %v", err)
	}

	if err := r.ParseMultipartForm(MaxMultipartFormSize); err != nil {
		respondError(w, http.StatusBadRequest, ErrFailedToParseForm)
		return
	}

	name := r.FormValue("name")
	kind := r.FormValue("kind")
	adoptedAt := r.FormValue("adopted_at")
	if name == "" {
		respondError(w, http.StatusBadRequest, ErrNameRequired)
		return
	}

	imagePath, err := saveUploadedFile(r, "image")
	if err != nil && !errors.Is(err, ErrNoFile) {
		respondError(w, http.StatusBadRequest, ErrFailedToSaveImage+": "+err.Error())
		return
	}

	now := time.Now().UTC()
	res, err := a.DB.Exec(`
		INSERT INTO plushies (user_id, name, kind, adopted_at, image_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, userID, name, kind, nullIfEmpty(adoptedAt), nullIfEmpty(imagePath), now, now)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			respondError(w, http.StatusBadRequest, ErrUserNotFound)
		} else {
			respondError(w, http.StatusInternalServerError, ErrFailedToCreatePlushie)
		}
		return
	}
	id, _ := res.LastInsertId()
	respondJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (a *App) HandleUpdatePlushie(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	id, err := parsePlushieID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := r.ParseMultipartForm(MaxMultipartFormSize); err != nil {
		respondError(w, http.StatusBadRequest, ErrFailedToParseForm)
		return
	}

	name := r.FormValue("name")
	kind := r.FormValue("kind")
	adoptedAt := r.FormValue("adopted_at")

	// Check ownership and get existing image
	var existingImage sql.NullString
	err = a.DB.QueryRow(`SELECT image_path FROM plushies WHERE id = ? AND user_id = ?`, id, userID).Scan(&existingImage)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, ErrPlushieNotFound)
		} else {
			respondError(w, http.StatusInternalServerError, ErrFailedToGetPlushie)
		}
		return
	}

	imagePath := ""
	filePath, err := saveUploadedFile(r, "image")
	if err != nil && !errors.Is(err, ErrNoFile) {
		respondError(w, http.StatusBadRequest, ErrFailedToSaveImage)
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
		respondError(w, http.StatusInternalServerError, ErrFailedToUpdatePlushie)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) HandleDeletePlushie(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	id, err := parsePlushieID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := a.DB.Exec(`DELETE FROM plushies WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, ErrFailedToDeletePlushie)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		respondError(w, http.StatusNotFound, ErrPlushieNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) HandleUpdateConversation(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	id, err := parsePlushieID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		ConversationHistory string `json:"conversation_history"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Check ownership
	if err := a.checkPlushieOwnership(id, userID); err != nil {
		if err.Error() == ErrPlushieNotFound {
			respondError(w, http.StatusNotFound, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	_, err = a.DB.Exec(`
		UPDATE plushies
		SET conversation_history = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`, req.ConversationHistory, time.Now().UTC(), id, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, ErrFailedToUpdateConv)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) HandleChat(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, ErrAuthRequired)
		return
	}

	id, err := parsePlushieID(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
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
			respondError(w, http.StatusNotFound, ErrPlushieNotFound)
		} else {
			respondError(w, http.StatusInternalServerError, ErrFailedToGetPlushie)
		}
		return
	}

	// Call OpenAI API
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		respondError(w, http.StatusInternalServerError, ErrOpenAIKeyNotSet)
		return
	}

	history := ""
	if conversationHistory.Valid {
		history = conversationHistory.String
	}

	prompt := buildChatPrompt(name, kind, history)
	message, err := callOpenAI(apiKey, prompt)
	if err != nil {
		respondError(w, http.StatusInternalServerError, ErrFailedToChat+": "+err.Error())
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
