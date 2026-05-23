package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"speaknow/internal/domain"
)

type Service struct {
	client *redis.Client
	ttl    time.Duration
}

func NewService(client *redis.Client, ttl time.Duration) *Service {
	return &Service{client: client, ttl: ttl}
}

func (s *Service) Get(ctx context.Context, key string) (*domain.CachedResult, error) {
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var cached domain.CachedResult
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return nil, fmt.Errorf("unmarshal cache: %w", err)
	}
	return &cached, nil
}

func (s *Service) Set(ctx context.Context, key string, result *domain.Result) error {
	cached := domain.CachedResult{
		Text:       result.Text,
		Confidence: result.Confidence,
		Provider:   result.Provider,
		Duration:   result.Duration,
		CreatedAt:  time.Now().Unix(),
	}
	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	return s.client.Set(ctx, key, data, s.ttl).Err()
}

func (s *Service) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
