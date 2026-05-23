package mock

import (
	"context"
	"fmt"
	"time"

	"speaknow/internal/domain"
)

func (p *Provider) StreamRecognize(ctx context.Context, audioIn <-chan []byte, opts domain.RecognizeOpts) (<-chan domain.StreamResult, error) {
	out := make(chan domain.StreamResult, 8)

	go func() {
		defer close(out)

		var total int
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				out <- domain.StreamResult{Type: "error", Message: ctx.Err().Error()}
				return
			case chunk, ok := <-audioIn:
				if !ok {
					text := buildMockText(total, opts)
					out <- domain.StreamResult{Type: "final", Text: text, Confidence: 0.95}
					return
				}
				total += len(chunk)
			case <-ticker.C:
				if total > 0 {
					partial := buildMockPartial(total, opts)
					out <- domain.StreamResult{Type: "partial", Text: partial}
				}
			}
		}
	}()

	return out, nil
}

func buildMockPartial(total int, opts domain.RecognizeOpts) string {
	sec := float64(total) / 32000.0
	return fmt.Sprintf("[实时识别中] 已接收 %.1f 秒音频…", sec)
}

func buildMockText(total int, opts domain.RecognizeOpts) string {
	text := fmt.Sprintf("[Mock实时识别] 共 %d 字节 PCM，语言=%s", total, opts.Language)
	if opts.EnablePunc {
		text += "。"
	}
	return text
}
