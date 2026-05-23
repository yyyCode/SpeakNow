package router

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"speaknow/internal/config"
	"speaknow/internal/domain"
	"speaknow/internal/provider"
)

type Service struct {
	registry *provider.Registry
	cfg      *config.ASRConfig
	log      *zap.Logger
}

func New(registry *provider.Registry, cfg *config.ASRConfig, log *zap.Logger) *Service {
	return &Service{registry: registry, cfg: cfg, log: log}
}

func (s *Service) Select(name string) ([]provider.Provider, error) {
	if name != "" && name != "auto" {
		p, ok := s.registry.Get(name)
		if !ok {
			return nil, fmt.Errorf("provider %q not found", name)
		}
		return []provider.Provider{p}, nil
	}

	chain := make([]provider.Provider, 0, 1+len(s.cfg.Fallback))
	if p, ok := s.registry.Get(s.cfg.Primary); ok {
		chain = append(chain, p)
	}
	for _, fb := range s.cfg.Fallback {
		if fb == s.cfg.Primary {
			continue
		}
		if p, ok := s.registry.Get(fb); ok {
			chain = append(chain, p)
		}
	}
	if len(chain) == 0 {
		return nil, fmt.Errorf("no available provider")
	}
	return chain, nil
}

func (s *Service) Recognize(ctx context.Context, audio []byte, opts domain.RecognizeOpts) (*domain.Result, error) {
	chain, err := s.Select(opts.Provider)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, p := range chain {
		if err := p.HealthCheck(ctx); err != nil {
			s.log.Warn("provider unhealthy, skip",
				zap.String("provider", p.Name()),
				zap.Error(err),
			)
			lastErr = err
			continue
		}

		result, err := p.Recognize(ctx, audio, opts)
		if err != nil {
			s.log.Warn("provider recognize failed, try fallback",
				zap.String("provider", p.Name()),
				zap.Error(err),
			)
			lastErr = err
			continue
		}
		result.Provider = p.Name()
		return result, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed: %w", lastErr)
	}
	return nil, fmt.Errorf("all providers unavailable")
}

func (s *Service) CostPerSecond(name string) float64 {
	if p, ok := s.registry.Get(name); ok {
		return p.CostPerSecond()
	}
	return 0
}

func (s *Service) StreamRecognize(ctx context.Context, audioIn <-chan []byte, opts domain.RecognizeOpts) (<-chan domain.StreamResult, error) {
	chain, err := s.Select(opts.Provider)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, p := range chain {
		if err := p.HealthCheck(ctx); err != nil {
			s.log.Warn("provider unhealthy, skip", zap.String("provider", p.Name()), zap.Error(err))
			lastErr = err
			continue
		}
		sr, ok := provider.AsStreamRecognizer(p)
		if !ok {
			lastErr = fmt.Errorf("provider %q does not support streaming", p.Name())
			continue
		}
		out, err := sr.StreamRecognize(ctx, audioIn, opts)
		if err != nil {
			s.log.Warn("stream recognize failed", zap.String("provider", p.Name()), zap.Error(err))
			lastErr = err
			continue
		}
		return out, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all providers failed: %w", lastErr)
	}
	return nil, fmt.Errorf("no streaming provider available")
}

func (s *Service) ProviderStatus(ctx context.Context) map[string]string {
	status := make(map[string]string)
	for _, p := range s.registry.List() {
		if err := p.HealthCheck(ctx); err != nil {
			status[p.Name()] = "unhealthy: " + err.Error()
		} else {
			status[p.Name()] = "healthy"
		}
	}
	return status
}
