package xunfei

import (
	"context"
	"fmt"
	"time"

	"speaknow/internal/domain"
)

type Provider struct {
	appID         string
	apiKey        string
	apiSecret     string
	hostURL       string
	costPerSecond float64
}

func New(appID, apiKey, apiSecret, hostURL string, costPerSecond float64) *Provider {
	if hostURL == "" {
		hostURL = defaultHostURL
	}
	return &Provider{
		appID:         appID,
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		hostURL:       hostURL,
		costPerSecond: costPerSecond,
	}
}

func (p *Provider) Name() string {
	return "xunfei"
}

func (p *Provider) CostPerSecond() float64 {
	return p.costPerSecond
}

func (p *Provider) HealthCheck(_ context.Context) error {
	if p.appID == "" || p.apiKey == "" || p.apiSecret == "" {
		return fmt.Errorf("xunfei credentials not configured")
	}
	return nil
}

func (p *Provider) Recognize(ctx context.Context, audio []byte, opts domain.RecognizeOpts) (*domain.Result, error) {
	pcm, err := preparePCM(audio, opts)
	if err != nil {
		return nil, err
	}

	recognizeCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	result, err := recognizePCM(recognizeCtx, p.appID, p.apiKey, p.apiSecret, p.hostURL, pcm, opts)
	if err != nil {
		return nil, err
	}
	result.Provider = p.Name()
	return result, nil
}
