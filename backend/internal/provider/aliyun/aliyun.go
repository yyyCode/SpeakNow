package aliyun

import (
	"context"
	"fmt"

	"speaknow/internal/domain"
)

// Provider 阿里云 ASR 占位实现，Phase 2 接入真实 SDK。
type Provider struct {
	appKey          string
	accessKeyID     string
	accessKeySecret string
	costPerSecond   float64
}

func New(appKey, accessKeyID, accessKeySecret string, costPerSecond float64) *Provider {
	return &Provider{
		appKey:          appKey,
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
		costPerSecond:   costPerSecond,
	}
}

func (p *Provider) Name() string {
	return "aliyun"
}

func (p *Provider) CostPerSecond() float64 {
	return p.costPerSecond
}

func (p *Provider) HealthCheck(_ context.Context) error {
	if p.accessKeyID == "" || p.accessKeySecret == "" {
		return fmt.Errorf("aliyun credentials not configured")
	}
	return nil
}

func (p *Provider) Recognize(_ context.Context, _ []byte, _ domain.RecognizeOpts) (*domain.Result, error) {
	return nil, fmt.Errorf("aliyun provider not implemented yet")
}
