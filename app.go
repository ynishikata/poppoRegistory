package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
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
	ID         int64     `json:"id"`
	UserID     int64     `json:"-"`
	Name       string    `json:"name"`
	Kind       string    `json:"kind"`
	AdoptedAt  string    `json:"adopted_at"` // ISO8601 (yyyy-mm-dd)
	ImageURL   string    `json:"image_url"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

// AuthMiddleware ensures user is logged in
func (a *App) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.SessionStore.GetUserIDFromRequest(r)
		if err != nil || userID == 0 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := withUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandleRegister creates a new user (email + password)
func (a *App) HandleRegister(w http.ResponseWriter, r *http.Request) {
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
	userID := userIDFromContext(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var email string
	err := a.DB.QueryRow(`SELECT email FROM users WHERE id = ?`, userID).Scan(&email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "user not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"id":    userID,
		"email": email,
	})
}

func (a *App) HandleListPlushies(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	rows, err := a.DB.Query(`
		SELECT id, name, kind, adopted_at, image_path, created_at, updated_at
		FROM plushies
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to query plushies")
		return
	}
	defer rows.Close()

	var items []Plushie
	for rows.Next() {
		var p Plushie
		var adoptedAt sql.NullString
		var imagePath sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Kind, &adoptedAt, &imagePath, &p.CreatedAt, &p.ModifiedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to scan plushy")
			return
		}
		p.UserID = userID
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

func (a *App) HandleCreatePlushie(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		respondError(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	name := r.FormValue("name")
	kind := r.FormValue("kind")
	adoptedAt := r.FormValue("adopted_at")
	if name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	imagePath, err := saveUploadedFile(r, "image")
	if err != nil && !errors.Is(err, ErrNoFile) {
		respondError(w, http.StatusBadRequest, "failed to save image")
		return
	}

	now := time.Now().UTC()
	res, err := a.DB.Exec(`
		INSERT INTO plushies (user_id, name, kind, adopted_at, image_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, userID, name, kind, nullIfEmpty(adoptedAt), nullIfEmpty(imagePath), now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to insert plushie")
		return
	}
	id, _ := res.LastInsertId()

	respondJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (a *App) HandleUpdatePlushie(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse form")
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
		respondError(w, http.StatusBadRequest, "failed to save image")
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
		respondError(w, http.StatusInternalServerError, "failed to update plushie")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) HandleDeletePlushie(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	res, err := a.DB.Exec(`DELETE FROM plushies WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete plushie")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		respondError(w, http.StatusNotFound, "plushie not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}


