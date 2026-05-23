package provider

import (
	"context"

	"speaknow/internal/domain"
)

type Provider interface {
	Name() string
	Recognize(ctx context.Context, audio []byte, opts domain.RecognizeOpts) (*domain.Result, error)
	HealthCheck(ctx context.Context) error
	CostPerSecond() float64
}

// StreamRecognizer 支持边录边传的流式识别。
type StreamRecognizer interface {
	Provider
	StreamRecognize(ctx context.Context, audioIn <-chan []byte, opts domain.RecognizeOpts) (<-chan domain.StreamResult, error)
}

func AsStreamRecognizer(p Provider) (StreamRecognizer, bool) {
	sr, ok := p.(StreamRecognizer)
	return sr, ok
}

type Registry struct {
	providers map[string]Provider
	order     []string
}

func NewRegistry(providers ...Provider) *Registry {
	r := &Registry{
		providers: make(map[string]Provider, len(providers)),
		order:     make([]string, 0, len(providers)),
	}
	for _, p := range providers {
		r.Register(p)
	}
	return r
}

func (r *Registry) Register(p Provider) {
	name := p.Name()
	if _, exists := r.providers[name]; !exists {
		r.order = append(r.order, name)
	}
	r.providers[name] = p
}

func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

func (r *Registry) List() []Provider {
	out := make([]Provider, 0, len(r.order))
	for _, name := range r.order {
		if p, ok := r.providers[name]; ok {
			out = append(out, p)
		}
	}
	return out
}

func (r *Registry) Names() []string {
	return append([]string(nil), r.order...)
}
