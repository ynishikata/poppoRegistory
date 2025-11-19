package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUnauthorized = errors.New(ErrAuthRequired)
)

// getUserIDFromRequest extracts user ID from request context
func getUserIDFromRequest(r *http.Request) (string, error) {
	userID := supabaseUserIDFromContext(r.Context())
	if userID == "" {
		return "", ErrUnauthorized
	}
	return userID, nil
}

// parsePlushieID parses plushie ID from URL parameter
func parsePlushieID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.New(ErrInvalidID)
	}
	return id, nil
}

// ensureUserExists ensures that a Supabase user exists in the users table
func (a *App) ensureUserExists(userID, email string) error {
	// Insert or ignore if user already exists
	_, err := a.DB.Exec(`
		INSERT OR IGNORE INTO users (supabase_user_id, email, password_hash, created_at)
		VALUES (?, ?, '', datetime('now'))
	`, userID, email)
	if err != nil {
		return err
	}
	// Update email if user exists
	_, err = a.DB.Exec(`
		UPDATE users SET email = ? WHERE supabase_user_id = ?
	`, email, userID)
	return err
}

// getEmailFromJWT extracts email from JWT token in Authorization header
func getEmailFromJWT(r *http.Request) (string, error) {
	supabaseAuth := NewSupabaseAuth()
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", nil
	}

	token, err := jwt.ParseWithClaims(parts[1], &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(supabaseAuth.JWTSecret), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*SupabaseClaims); ok && token.Valid {
		return claims.Email, nil
	}

	return "", nil
}

// ensureUserExistsFromRequest ensures user exists by extracting email from JWT
func (a *App) ensureUserExistsFromRequest(r *http.Request, userID string) error {
	email, err := getEmailFromJWT(r)
	if err != nil || email == "" {
		// If we can't get email, still try to ensure user exists with empty email
		_, _ = a.DB.Exec(`
			INSERT OR IGNORE INTO users (supabase_user_id, email, password_hash, created_at)
			VALUES (?, '', '', datetime('now'))
		`, userID)
		return nil
	}
	return a.ensureUserExists(userID, email)
}

// scanPlushieFromRow scans a plushie from database row
func scanPlushieFromRow(rows *sql.Rows) (*Plushie, error) {
	var p Plushie
	var adoptedAt sql.NullString
	var imagePath sql.NullString
	var conversationHistory sql.NullString

	err := rows.Scan(
		&p.ID, &p.UserID, &p.Name, &p.Kind,
		&adoptedAt, &imagePath, &conversationHistory,
		&p.CreatedAt, &p.ModifiedAt,
	)
	if err != nil {
		return nil, err
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

	return &p, nil
}

// checkPlushieOwnership checks if a plushie belongs to a user
func (a *App) checkPlushieOwnership(plushieID int64, userID string) error {
	var exists int
	err := a.DB.QueryRow(
		`SELECT 1 FROM plushies WHERE id = ? AND user_id = ?`,
		plushieID, userID,
	).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New(ErrPlushieNotFound)
		}
		return errors.New(ErrFailedToGetPlushie)
	}
	return nil
}

