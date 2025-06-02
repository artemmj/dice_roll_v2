package storage

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type GameSession struct {
	ID         string
	ServerSeed string
	ClientSeed string
	Nonce      int
	CreatedAt  time.Time
}

type SessionStorage interface {
	Save(ctx context.Context, session *GameSession) error
	Get(ctx context.Context, id string) (*GameSession, error)
	Update(ctx context.Context, session *GameSession) error
	Delete(ctx context.Context, id string) error
}

type InMemoryStorage struct {
	mu       sync.RWMutex
	sessions map[string]*GameSession
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		sessions: make(map[string]*GameSession),
	}
}

func (s *InMemoryStorage) Save(ctx context.Context, session *GameSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	return nil
}

func (s *InMemoryStorage) Get(ctx context.Context, id string) (*GameSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[id]
	if !exists {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *InMemoryStorage) Update(ctx context.Context, session *GameSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.sessions[session.ID]; !exists {
		return ErrSessionNotFound
	}
	s.sessions[session.ID] = session
	return nil
}

func (s *InMemoryStorage) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}
