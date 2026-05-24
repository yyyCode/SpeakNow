package cache

import (
	"context"
	"sync"
	"time"

	"speaknow/internal/domain"
)

type entry struct {
	result    domain.CachedResult
	expiresAt time.Time
}

type Service struct {
	mu    sync.RWMutex
	items map[string]entry
	ttl   time.Duration
}

func NewService(ttl time.Duration) *Service {
	return &Service{
		items: make(map[string]entry),
		ttl:   ttl,
	}
}

func (s *Service) Get(_ context.Context, key string) (*domain.CachedResult, error) {
	s.mu.RLock()
	e, ok := s.items[key]
	s.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	if time.Now().After(e.expiresAt) {
		s.mu.Lock()
		delete(s.items, key)
		s.mu.Unlock()
		return nil, nil
	}
	cached := e.result
	return &cached, nil
}

func (s *Service) Set(_ context.Context, key string, result *domain.Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = entry{
		result: domain.CachedResult{
			Text:       result.Text,
			Confidence: result.Confidence,
			Provider:   result.Provider,
			Duration:   result.Duration,
			CreatedAt:  time.Now().Unix(),
		},
		expiresAt: time.Now().Add(s.ttl),
	}
	return nil
}
