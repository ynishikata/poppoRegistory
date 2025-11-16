package main

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

const sessionCookieName = "poppo_session"

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]sessionEntry
}

type sessionEntry struct {
	UserID int64
	Expiry time.Time
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]sessionEntry),
	}
}

func (s *SessionStore) Set(token string, userID int64, expiry time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = sessionEntry{UserID: userID, Expiry: expiry}
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func (s *SessionStore) GetUserID(token string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.sessions[token]
	if !ok {
		return 0, errors.New("not found")
	}
	if time.Now().After(entry.Expiry) {
		return 0, errors.New("expired")
	}
	return entry.UserID, nil
}

func (s *SessionStore) GetUserIDFromRequest(r *http.Request) (int64, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return 0, err
	}
	return s.GetUserID(c.Value)
}

type contextKey string

const userIDKey contextKey = "user_id"

func withUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func userIDFromContext(ctx context.Context) int64 {
	v := ctx.Value(userIDKey)
	if v == nil {
		return 0
	}
	if id, ok := v.(int64); ok {
		return id
	}
	return 0
}


