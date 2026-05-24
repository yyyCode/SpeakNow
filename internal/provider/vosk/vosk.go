package vosk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "speaknow/internal/voskruntime" // init: 解压嵌入的 DLL / 设置搜索路径（须早于下方 CGO 包）

	voskapi "github.com/alphacep/vosk-api/go"

	"speaknow/internal/domain"
)

const providerName = "vosk"

type Provider struct {
	modelPath     string
	sampleRate    float64
	costPerSecond float64

	model   *voskapi.VoskModel
	modelMu sync.RWMutex
}

func New(modelPath string, sampleRate float64, costPerSecond float64) (*Provider, error) {
	if modelPath == "" {
		return nil, fmt.Errorf("vosk model_path is required")
	}
	if sampleRate <= 0 {
		sampleRate = 16000
	}

	absPath, err := filepath.Abs(modelPath)
	if err != nil {
		return nil, fmt.Errorf("resolve model path: %w", err)
	}
	if err := verifyModelDir(absPath); err != nil {
		return nil, err
	}

	voskapi.SetLogLevel(-1)
	model, err := voskapi.NewModel(absPath)
	if err != nil {
		return nil, fmt.Errorf("load vosk model: %w", err)
	}

	return &Provider{
		modelPath:     absPath,
		sampleRate:    sampleRate,
		costPerSecond: costPerSecond,
		model:         model,
	}, nil
}

func (p *Provider) Name() string {
	return providerName
}

func (p *Provider) CostPerSecond() float64 {
	return p.costPerSecond
}

func (p *Provider) HealthCheck(_ context.Context) error {
	if err := verifyModelDir(p.modelPath); err != nil {
		return err
	}
	p.modelMu.RLock()
	defer p.modelMu.RUnlock()
	if p.model == nil {
		return fmt.Errorf("vosk model not loaded")
	}
	return nil
}

func (p *Provider) Recognize(ctx context.Context, audio []byte, opts domain.RecognizeOpts) (*domain.Result, error) {
	if err := p.HealthCheck(ctx); err != nil {
		return nil, err
	}
	if err := checkLanguage(opts.Language); err != nil {
		return nil, err
	}

	pcm, err := preparePCM(audio, opts)
	if err != nil {
		return nil, err
	}
	if len(pcm) == 0 {
		return nil, fmt.Errorf("empty audio")
	}

	text, err := p.transcribePCM(ctx, pcm)
	if err != nil {
		return nil, err
	}

	return &domain.Result{
		Text:       text,
		Confidence: 0.85,
		Duration:   estimateDuration(len(pcm)),
		Provider:   p.Name(),
	}, nil
}

func (p *Provider) transcribePCM(ctx context.Context, pcm []byte) (string, error) {
	rec, err := p.newRecognizer()
	if err != nil {
		return "", err
	}
	defer rec.Free()

	const chunkSize = 4096
	for offset := 0; offset < len(pcm); offset += chunkSize {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		end := offset + chunkSize
		if end > len(pcm) {
			end = len(pcm)
		}
		_ = rec.AcceptWaveform(pcm[offset:end])
	}

	return parseVoskText(rec.FinalResult()), nil
}

func (p *Provider) newRecognizer() (*voskapi.VoskRecognizer, error) {
	p.modelMu.RLock()
	defer p.modelMu.RUnlock()
	if p.model == nil {
		return nil, fmt.Errorf("vosk model not loaded")
	}
	rec, err := voskapi.NewRecognizer(p.model, p.sampleRate)
	if err != nil {
		return nil, fmt.Errorf("create recognizer: %w", err)
	}
	return rec, nil
}

func verifyModelDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("vosk model path %q: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("vosk model path %q is not a directory", path)
	}
	if _, err := os.Stat(filepath.Join(path, "conf", "model.conf")); err != nil {
		return fmt.Errorf("vosk model at %q missing conf/model.conf", path)
	}
	return nil
}

func checkLanguage(lang string) error {
	switch lang {
	case "", "zh-CN", "zh_CN", "zh-cn":
		return nil
	default:
		return fmt.Errorf("vosk local model only supports zh-CN, got %q", lang)
	}
}
