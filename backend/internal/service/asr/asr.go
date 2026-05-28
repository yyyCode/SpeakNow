package asr

import (
	"context"
	"time"

	"go.uber.org/zap"

	"speaknow/internal/domain"
	"speaknow/internal/service/cache"
	"speaknow/internal/service/cost"
	"speaknow/internal/service/router"
	"speaknow/pkg/fingerprint"
)

type Service struct {
	router *router.Service
	cache  *cache.Service
	cost   *cost.Service
	log    *zap.Logger
}

func New(routerSvc *router.Service, cacheSvc *cache.Service, costSvc *cost.Service, log *zap.Logger) *Service {
	return &Service{
		router: routerSvc,
		cache:  cacheSvc,
		cost:   costSvc,
		log:    log,
	}
}

type RecognizeResponse struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Provider   string  `json:"provider"`
	DurationMs int     `json:"duration_ms"`
	LatencyMs  int     `json:"latency_ms"`
	CacheHit   bool    `json:"cache_hit"`
}

func (s *Service) Recognize(ctx context.Context, userID string, audio []byte, opts domain.RecognizeOpts) (*RecognizeResponse, error) {
	start := time.Now()
	cacheKey := fingerprint.BuildCacheKey(audio, opts)

	if cached, err := s.cache.Get(ctx, cacheKey); err != nil {
		s.log.Warn("cache get failed", zap.Error(err))
	} else if cached != nil {
		latency := int(time.Since(start).Milliseconds())
		s.cost.Record(domain.CallLog{
			UserID:          userID,
			Provider:        cached.Provider,
			AudioDurationMs: int(cached.Duration * 1000),
			LatencyMs:       latency,
			CacheHit:        true,
			TextLength:      len(cached.Text),
			Status:          "success",
		})
		return &RecognizeResponse{
			Text:       cached.Text,
			Confidence: cached.Confidence,
			Provider:   cached.Provider,
			DurationMs: int(cached.Duration * 1000),
			LatencyMs:  latency,
			CacheHit:   true,
		}, nil
	}

	result, err := s.router.Recognize(ctx, audio, opts)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		s.cost.Record(domain.CallLog{
			UserID:       userID,
			Provider:     opts.Provider,
			LatencyMs:    latency,
			Status:       "error",
			ErrorMessage: err.Error(),
		})
		return nil, err
	}

	if err := s.cache.Set(ctx, cacheKey, result); err != nil {
		s.log.Warn("cache set failed", zap.Error(err))
	}

	costYuan := result.Duration * s.router.CostPerSecond(result.Provider)
	s.cost.Record(domain.CallLog{
		UserID:          userID,
		Provider:        result.Provider,
		AudioDurationMs: int(result.Duration * 1000),
		CostYuan:        costYuan,
		LatencyMs:       latency,
		CacheHit:        false,
		TextLength:      len(result.Text),
		Status:          "success",
	})

	return &RecognizeResponse{
		Text:       result.Text,
		Confidence: result.Confidence,
		Provider:   result.Provider,
		DurationMs: int(result.Duration * 1000),
		LatencyMs:  latency,
		CacheHit:   false,
	}, nil
}
