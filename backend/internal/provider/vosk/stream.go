package vosk

import (
	"context"

	"speaknow/internal/domain"
)

func (p *Provider) StreamRecognize(ctx context.Context, audioIn <-chan []byte, opts domain.RecognizeOpts) (<-chan domain.StreamResult, error) {
	if err := p.HealthCheck(ctx); err != nil {
		return nil, err
	}
	if err := checkLanguage(opts.Language); err != nil {
		return nil, err
	}

	rec, err := p.newRecognizer()
	if err != nil {
		return nil, err
	}

	out := make(chan domain.StreamResult, 16)

	go func() {
		defer close(out)
		defer rec.Free()

		lastPartial := ""
		emitPartial := func(text string) {
			if text == "" || text == lastPartial {
				return
			}
			lastPartial = text
			out <- domain.StreamResult{
				Type:     "partial",
				Text:     text,
				Provider: providerName,
			}
		}

		for {
			select {
			case <-ctx.Done():
				out <- domain.StreamResult{Type: "error", Message: ctx.Err().Error(), Provider: providerName}
				return
			case chunk, ok := <-audioIn:
				if !ok {
					finalText := parseVoskText(rec.FinalResult())
					if finalText == "" {
						finalText = lastPartial
					}
					out <- domain.StreamResult{
						Type:       "final",
						Text:       finalText,
						Confidence: 0.85,
						Provider:   providerName,
					}
					return
				}
				if len(chunk) == 0 {
					continue
				}
				if rec.AcceptWaveform(chunk) != 0 {
					emitPartial(parseVoskText(rec.Result()))
				} else {
					emitPartial(parseVoskText(rec.PartialResult()))
				}
			}
		}
	}()

	return out, nil
}
