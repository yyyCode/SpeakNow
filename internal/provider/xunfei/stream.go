package xunfei

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"speaknow/internal/domain"
)

// StreamRecognize 实时流式识别：audioIn 收到 PCM 分片，resultOut 推送 partial/final。
func (p *Provider) StreamRecognize(ctx context.Context, audioIn <-chan []byte, opts domain.RecognizeOpts) (<-chan domain.StreamResult, error) {
	if err := p.HealthCheck(ctx); err != nil {
		return nil, err
	}

	conn, err := dialXunfei(ctx, p.apiKey, p.apiSecret, p.hostURL)
	if err != nil {
		return nil, err
	}

	out := make(chan domain.StreamResult, 16)
	lang, accent, ptt := businessParams(opts)

	go func() {
		defer close(out)
		defer conn.Close()

		decoder := &Decoder{}
		var totalBytes int
		var sendMu sync.Mutex

		// 读取讯飞返回
		readDone := make(chan error, 1)
		go func() {
			for {
				select {
				case <-ctx.Done():
					readDone <- ctx.Err()
					return
				default:
				}

				if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
					readDone <- err
					return
				}

				_, msg, err := conn.ReadMessage()
				if err != nil {
					readDone <- err
					return
				}

				var resp wsResp
				if err := json.Unmarshal(msg, &resp); err != nil {
					readDone <- fmt.Errorf("unmarshal: %w", err)
					return
				}
				if resp.Code != 0 {
					readDone <- fmt.Errorf("xunfei error %d: %s", resp.Code, resp.Message)
					return
				}

				decoder.Decode(&resp.Data.Result)
				text := decoder.String()
				if text != "" {
					out <- domain.StreamResult{
						Type:     "partial",
						Text:     text,
						Provider: p.Name(),
					}
				}

				if resp.Data.Status == statusLastFrame {
					out <- domain.StreamResult{
						Type:       "final",
						Text:       text,
						Confidence: 0.9,
						Provider:   p.Name(),
					}
					readDone <- nil
					return
				}
			}
		}()

		// 发送音频分片
		sendErr := p.streamSendAudio(ctx, conn, &sendMu, p.appID, lang, accent, ptt, audioIn, &totalBytes)
		if sendErr != nil {
			out <- domain.StreamResult{Type: "error", Message: sendErr.Error()}
			conn.Close()
			<-readDone
			return
		}

		if err := <-readDone; err != nil && err != context.Canceled {
			out <- domain.StreamResult{Type: "error", Message: err.Error()}
		}
	}()

	return out, nil
}

func (p *Provider) streamSendAudio(
	ctx context.Context,
	conn *websocket.Conn,
	mu *sync.Mutex,
	appID, lang, accent string,
	ptt int,
	audioIn <-chan []byte,
	totalBytes *int,
) error {
	first := true
	buf := make([]byte, 0, frameSize*2)

	flush := func(chunk []byte, status int) error {
		mu.Lock()
		defer mu.Unlock()
		var frame map[string]interface{}
		switch status {
		case statusFirstFrame:
			frame = firstFrame(appID, lang, accent, ptt, chunk)
		case statusContinueFrame:
			frame = continueFrame(chunk)
		default:
			frame = lastFrame(chunk)
		}
		return conn.WriteJSON(frame)
	}

	sendBuffer := func(forceLast bool) error {
		for len(buf) >= frameSize {
			chunk := buf[:frameSize]
			buf = buf[frameSize:]
			st := statusContinueFrame
			if first {
				st = statusFirstFrame
				first = false
			}
			if err := flush(chunk, st); err != nil {
				return err
			}
			time.Sleep(frameInterval)
		}
		if forceLast {
			st := statusContinueFrame
			if first {
				st = statusFirstFrame
				first = false
				if err := flush(buf, st); err != nil {
					return err
				}
				buf = nil
				return flush(nil, statusLastFrame)
			}
			if len(buf) > 0 {
				if err := flush(buf, statusLastFrame); err != nil {
					return err
				}
			} else {
				if err := flush(nil, statusLastFrame); err != nil {
					return err
				}
			}
			buf = nil
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case chunk, ok := <-audioIn:
			if !ok {
				return sendBuffer(true)
			}
			if len(chunk) == 0 {
				continue
			}
			*totalBytes += len(chunk)
			buf = append(buf, chunk...)
			if err := sendBuffer(false); err != nil {
				return err
			}
		}
	}
}
