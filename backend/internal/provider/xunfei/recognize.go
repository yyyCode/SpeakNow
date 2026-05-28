package xunfei

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"

	"speaknow/internal/domain"
)

const (
	statusFirstFrame    = 0
	statusContinueFrame = 1
	statusLastFrame     = 2

	frameSize     = 1280
	frameInterval = 40 * time.Millisecond
)

func recognizePCM(ctx context.Context, appID, apiKey, apiSecret, hostURL string, pcm []byte, opts domain.RecognizeOpts) (*domain.Result, error) {
	if len(pcm) == 0 {
		return nil, fmt.Errorf("empty pcm audio")
	}

	authURL, err := assembleAuthURL(hostURL, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, resp, err := dialer.DialContext(ctx, authURL, nil)
	if err != nil {
		if resp != nil && resp.Body != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("websocket dial: %s (%s)", err, string(body))
		}
		return nil, fmt.Errorf("websocket dial: %w", err)
	}
	defer conn.Close()

	if resp != nil && resp.StatusCode != 101 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("websocket handshake failed: code=%d body=%s", resp.StatusCode, string(body))
	}

	sendDone := make(chan error, 1)
	go func() {
		sendDone <- sendAudioFrames(conn, appID, pcm, opts)
	}()

	decoder := &Decoder{}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			return nil, err
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf("read message: %w", err)
		}

		var wsResp wsResp
		if err := json.Unmarshal(msg, &wsResp); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}

		if wsResp.Code != 0 {
			return nil, fmt.Errorf("xunfei error %d: %s", wsResp.Code, wsResp.Message)
		}

		decoder.Decode(&wsResp.Data.Result)

		if wsResp.Data.Status == statusLastFrame {
			break
		}
	}

	if err := <-sendDone; err != nil {
		return nil, err
	}

	text := decoder.String()
	duration := float64(len(pcm)) / 32000.0

	return &domain.Result{
		Text:       text,
		Confidence: 0.9,
		Duration:   duration,
		Provider:   "xunfei",
		Segments: []domain.Segment{{
			Text: text,
			End:  duration,
		}},
	}, nil
}

func sendAudioFrames(conn *websocket.Conn, appID string, pcm []byte, opts domain.RecognizeOpts) error {
	lang, accent := mapLanguage(opts.Language)
	ptt := 0
	if opts.EnablePunc {
		ptt = 1
	}

	status := statusFirstFrame
	offset := 0

	for {
		end := offset + frameSize
		if end > len(pcm) {
			end = len(pcm)
		}
		chunk := pcm[offset:end]

		atEOF := end >= len(pcm)
		frameStatus := status
		if atEOF && status != statusFirstFrame {
			frameStatus = statusLastFrame
		}

		var frame map[string]interface{}
		switch status {
		case statusFirstFrame:
			frame = map[string]interface{}{
				"common": map[string]interface{}{
					"app_id": appID,
				},
				"business": map[string]interface{}{
					"language": lang,
					"domain":   "iat",
					"accent":   accent,
					"vad_eos":  3000,
					"dwa":      "wpgs",
					"ptt":      ptt,
				},
				"data": audioDataFrame(statusFirstFrame, chunk),
			}
			if atEOF {
				// 短音频：第一帧后立即发结束帧
				if err := conn.WriteJSON(frame); err != nil {
					return fmt.Errorf("write first frame: %w", err)
				}
				lastFrame := map[string]interface{}{
					"data": audioDataFrame(statusLastFrame, []byte{}),
				}
				return conn.WriteJSON(lastFrame)
			}
			status = statusContinueFrame
		case statusContinueFrame:
			if atEOF {
				frameStatus = statusLastFrame
			}
			frame = map[string]interface{}{
				"data": audioDataFrame(frameStatus, chunk),
			}
		default:
			frame = map[string]interface{}{
				"data": audioDataFrame(statusLastFrame, chunk),
			}
		}

		if err := conn.WriteJSON(frame); err != nil {
			return fmt.Errorf("write frame: %w", err)
		}

		if frameStatus == statusLastFrame || status == statusLastFrame {
			return nil
		}

		offset = end
		time.Sleep(frameInterval)
	}
}

func audioDataFrame(status int, chunk []byte) map[string]interface{} {
	return map[string]interface{}{
		"status":   status,
		"format":   "audio/L16;rate=16000",
		"audio":    base64.StdEncoding.EncodeToString(chunk),
		"encoding": "raw",
	}
}

func mapLanguage(lang string) (language, accent string) {
	switch lang {
	case "en-US", "en_us":
		return "en_us", "mandarin"
	case "ja-JP", "ja_jp":
		return "ja_jp", "mandarin"
	case "ko-KR", "ko_kr":
		return "ko_kr", "mandarin"
	case "zh-TW", "zh_tw":
		return "zh_cn", "cantonese"
	default:
		return "zh_cn", "mandarin"
	}
}

func preparePCM(audio []byte, opts domain.RecognizeOpts) ([]byte, error) {
	switch opts.Format {
	case "pcm", "raw", "l16":
		return audio, nil
	case "wav":
		return extractPCMFromWAV(audio)
	case "webm", "ogg", "mp3", "unknown":
		return nil, fmt.Errorf("xunfei requires pcm/wav (16kHz mono 16-bit), got %q; please upload wav/pcm", opts.Format)
	default:
		if len(audio) > 4 && string(audio[:4]) == "RIFF" {
			return extractPCMFromWAV(audio)
		}
		return nil, fmt.Errorf("unsupported audio format %q for xunfei", opts.Format)
	}
}

func extractPCMFromWAV(data []byte) ([]byte, error) {
	if len(data) < 12 || string(data[:4]) != "RIFF" {
		return nil, fmt.Errorf("invalid wav file")
	}

	offset := 12
	for offset+8 <= len(data) {
		chunkID := string(data[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		offset += 8
		if offset+chunkSize > len(data) {
			break
		}
		if chunkID == "data" {
			return data[offset : offset+chunkSize], nil
		}
		offset += chunkSize
	}
	return nil, fmt.Errorf("wav data chunk not found")
}
