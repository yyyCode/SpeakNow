package cost

import (
	"sync"
	"time"

	"speaknow/internal/domain"
)

type Service struct {
	mu    sync.RWMutex
	logs  []domain.CallLog
	total float64
}

func New() *Service {
	return &Service{logs: make([]domain.CallLog, 0, 128)}
}

func (s *Service) Record(log domain.CallLog) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.ID = uint(len(s.logs) + 1)
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	s.logs = append(s.logs, log)
	if log.Status == "success" && !log.CacheHit {
		s.total += log.CostYuan
	}
}

func (s *Service) TotalCost() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.total
}

func (s *Service) Recent(limit int) []domain.CallLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.logs) {
		limit = len(s.logs)
	}
	start := len(s.logs) - limit
	if start < 0 {
		start = 0
	}
	out := make([]domain.CallLog, limit)
	copy(out, s.logs[start:])
	return out
}

func (s *Service) Summary() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var success, cacheHit int
	for _, l := range s.logs {
		if l.Status == "success" {
			success++
		}
		if l.CacheHit {
			cacheHit++
		}
	}

	hitRatio := 0.0
	if len(s.logs) > 0 {
		hitRatio = float64(cacheHit) / float64(len(s.logs))
	}

	return map[string]interface{}{
		"total_calls":     len(s.logs),
		"success_calls":   success,
		"cache_hit_ratio": hitRatio,
		"total_cost_yuan": s.total,
	}
}
