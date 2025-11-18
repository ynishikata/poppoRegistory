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

	// Debug: Log token info (first 50 chars only)
	if len(tokenString) > 50 {
		log.Printf("DEBUG: Verifying token, length: %d, first 50 chars: %s", len(tokenString), tokenString[:50])
	} else {
		log.Printf("DEBUG: Verifying token: %s", tokenString)
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
		// More detailed error logging
		if strings.Contains(err.Error(), "signature is invalid") {
			log.Printf("ERROR: JWT signature verification failed. Check SUPABASE_JWT_SECRET.")
			log.Printf("DEBUG: JWT Secret length: %d", len(s.JWTSecret))
		}
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*SupabaseClaims); ok && token.Valid {
		// Check expiration
		if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
			log.Printf("DEBUG: Token expired. Exp: %d, Now: %d", claims.Exp, time.Now().Unix())
			return "", errors.New("token expired")
		}
		log.Printf("DEBUG: Token verified successfully, userID: %s", claims.Sub)
		return claims.Sub, nil
	}

	log.Printf("DEBUG: Token validation failed. Valid: %v", token.Valid)
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
			respondError(w, http.StatusInternalServerError, "サーバー設定エラー: SUPABASE_JWT_SECRETが設定されていません")
			return
		}
		
		// Debug: Log Authorization header (first 50 chars only)
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			if len(authHeader) > 50 {
				log.Printf("DEBUG: Authorization header present, length: %d, first 50 chars: %s", len(authHeader), authHeader[:50])
			} else {
				log.Printf("DEBUG: Authorization header: %s", authHeader)
			}
		} else {
			log.Printf("DEBUG: No Authorization header found")
		}
		
		userID, err := supabaseAuth.GetUserIDFromRequest(r)
		if err != nil || userID == "" {
			log.Printf("Auth error: %v", err)
			// More detailed error logging
			if err != nil {
				log.Printf("Auth error details: %T, %v", err, err)
			}
			errMsg := "認証に失敗しました"
			if err != nil {
				if strings.Contains(err.Error(), "token expired") {
					errMsg = "トークンの有効期限が切れています。再度ログインしてください。"
				} else if strings.Contains(err.Error(), "no valid token") {
					errMsg = "認証トークンが見つかりません。ログインしてください。"
				} else if strings.Contains(err.Error(), "SUPABASE_JWT_SECRET") {
					errMsg = "サーバー設定エラーが発生しました。"
				} else if strings.Contains(err.Error(), "signature is invalid") {
					errMsg = "トークンの検証に失敗しました。SUPABASE_JWT_SECRETが正しく設定されているか確認してください。"
				} else if strings.Contains(err.Error(), "failed to parse token") {
					errMsg = "トークンの解析に失敗しました。"
				}
			}
			respondError(w, http.StatusUnauthorized, errMsg)
			return
		}
		log.Printf("DEBUG: Auth successful, userID: %s", userID)
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
