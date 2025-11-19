package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SupabaseAuth struct {
	JWTSecret string
}

type SupabaseClaims struct {
	Sub      string                 `json:"sub"` // User ID (UUID)
	Email    string                 `json:"email"`
	Aud      string                 `json:"aud"`
	Role     string                 `json:"role"`
	Exp      int64                  `json:"exp"`
	Iat      int64                  `json:"iat"`
	UserMeta map[string]interface{} `json:"user_metadata,omitempty"`
	jwt.RegisteredClaims
}

// NewSupabaseAuth creates a new Supabase auth instance
func NewSupabaseAuth() *SupabaseAuth {
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	return &SupabaseAuth{
		JWTSecret: jwtSecret,
	}
}

// VerifyToken verifies a Supabase JWT token and returns the user ID
func (s *SupabaseAuth) VerifyToken(tokenString string) (string, error) {
	if s.JWTSecret == "" {
		return "", errors.New("SUPABASE_JWT_SECRET not configured")
	}

	// Parse token with verification
	token, err := jwt.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.JWTSecret), nil
	})

	if err != nil {
		log.Printf("JWT parse error: %v", err)
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*SupabaseClaims); ok && token.Valid {
		// Check expiration
		if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
			return "", errors.New("token expired")
		}
		return claims.Sub, nil
	}

	return "", errors.New("invalid token")
}

// GetUserIDFromRequest extracts user ID from Authorization header or cookie
func (s *SupabaseAuth) GetUserIDFromRequest(r *http.Request) (string, error) {
	if s.JWTSecret == "" {
		return "", errors.New("SUPABASE_JWT_SECRET not configured")
	}

	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			userID, err := s.VerifyToken(parts[1])
			if err != nil {
				return "", fmt.Errorf("token verification failed: %w", err)
			}
			return userID, nil
		}
	}

	// Try cookie (for browser-based auth)
	cookie, err := r.Cookie("sb-access-token")
	if err == nil && cookie.Value != "" {
		userID, err := s.VerifyToken(cookie.Value)
		if err != nil {
			return "", fmt.Errorf("cookie token verification failed: %w", err)
		}
		return userID, nil
	}

	// Try alternative cookie name
	cookie, err = r.Cookie("supabase-auth-token")
	if err == nil && cookie.Value != "" {
		userID, err := s.VerifyToken(cookie.Value)
		if err != nil {
			return "", fmt.Errorf("cookie token verification failed: %w", err)
		}
		return userID, nil
	}

	return "", errors.New("no valid token found: check Authorization header or login status")
}

// SupabaseAuthMiddleware verifies Supabase JWT and sets user ID in context
func (a *App) SupabaseAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		supabaseAuth := NewSupabaseAuth()
		if supabaseAuth.JWTSecret == "" {
			log.Printf("ERROR: SUPABASE_JWT_SECRET not configured")
			respondError(w, http.StatusInternalServerError, ErrServerConfigError)
			return
		}
		userID, err := supabaseAuth.GetUserIDFromRequest(r)
		if err != nil || userID == "" {
			log.Printf("Auth error: %v", err)
			errMsg := ErrAuthFailed
			if err != nil {
				if strings.Contains(err.Error(), "token expired") {
					errMsg = ErrTokenExpired
				} else if strings.Contains(err.Error(), "no valid token") {
					errMsg = ErrTokenNotFound
				} else if strings.Contains(err.Error(), "SUPABASE_JWT_SECRET") {
					errMsg = ErrServerConfigError
				}
			}
			respondError(w, http.StatusUnauthorized, errMsg)
			return
		}
		ctx := withSupabaseUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

const supabaseUserIDKey contextKey = "supabase_user_id"

func withSupabaseUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, supabaseUserIDKey, userID)
}

func supabaseUserIDFromContext(ctx context.Context) string {
	v := ctx.Value(supabaseUserIDKey)
	if v == nil {
		return ""
	}
	if id, ok := v.(string); ok {
		return id
	}
	return ""
}

// Helper function to convert UUID string to int64 for backward compatibility
// Note: This is a temporary solution. Ideally, we should migrate to UUID.
func uuidToInt64(uuidStr string) int64 {
	// Simple hash-based conversion (not perfect, but works for migration)
	hash := int64(0)
	for _, c := range uuidStr {
		hash = hash*31 + int64(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}
