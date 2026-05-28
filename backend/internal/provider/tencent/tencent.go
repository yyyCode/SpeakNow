package tencent

import (
	"context"
	"fmt"

	"speaknow/internal/domain"
)

type Provider struct {
	secretID      string
	secretKey     string
	appID         string
	costPerSecond float64
}

func New(secretID, secretKey, appID string, costPerSecond float64) *Provider {
	return &Provider{
		secretID:      secretID,
		secretKey:     secretKey,
		appID:         appID,
		costPerSecond: costPerSecond,
	}
}

func (p *Provider) Name() string {
	return "tencent"
}

func (p *Provider) CostPerSecond() float64 {
	return p.costPerSecond
}

func (p *Provider) HealthCheck(_ context.Context) error {
	if p.secretID == "" || p.secretKey == "" {
		return fmt.Errorf("tencent credentials not configured")
	}
	return nil
}

func (p *Provider) Recognize(_ context.Context, _ []byte, _ domain.RecognizeOpts) (*domain.Result, error) {
	return nil, fmt.Errorf("tencent provider not implemented yet")
}
