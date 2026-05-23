package mock

import (
	"context"
	"fmt"
	"time"

	"speaknow/internal/domain"
)

type Provider struct {
	costPerSecond float64
}

func New(costPerSecond float64) *Provider {
	return &Provider{costPerSecond: costPerSecond}
}

func (p *Provider) Name() string {
	return "mock"
}

func (p *Provider) CostPerSecond() float64 {
	return p.costPerSecond
}

func (p *Provider) HealthCheck(_ context.Context) error {
	return nil
}

func (p *Provider) Recognize(_ context.Context, audio []byte, opts domain.RecognizeOpts) (*domain.Result, error) {
	if len(audio) == 0 {
		return nil, fmt.Errorf("empty audio")
	}

	duration := estimateDuration(audio, opts.Format)
	text := fmt.Sprintf("[Mock识别] 收到 %d 字节音频，语言=%s", len(audio), opts.Language)
	if opts.EnablePunc {
		text += "。"
	}

	time.Sleep(80 * time.Millisecond)

	return &domain.Result{
		Text:       text,
		Confidence: 0.95,
		Duration:   duration,
		Provider:   p.Name(),
	}, nil
}

func estimateDuration(audio []byte, format string) float64 {
	switch format {
	case "wav":
		if len(audio) > 44 {
			// 16kHz mono 16-bit ≈ 32000 bytes/s
			return float64(len(audio)-44) / 32000.0
		}
	case "pcm":
		return float64(len(audio)) / 32000.0
	}
	// 其他格式按 16KB/s 估算
	if len(audio) == 0 {
		return 0
	}
	return float64(len(audio)) / 16000.0
}
