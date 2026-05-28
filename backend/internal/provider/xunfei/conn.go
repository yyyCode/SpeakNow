package xunfei

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"

	"speaknow/internal/domain"
)

func dialXunfei(ctx context.Context, apiKey, apiSecret, hostURL string) (*websocket.Conn, error) {
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
	if resp != nil && resp.StatusCode != 101 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		conn.Close()
		return nil, fmt.Errorf("websocket handshake failed: code=%d body=%s", resp.StatusCode, string(body))
	}
	return conn, nil
}

func businessParams(opts domain.RecognizeOpts) (lang, accent string, ptt int) {
	lang, accent = mapLanguage(opts.Language)
	ptt = 0
	if opts.EnablePunc {
		ptt = 1
	}
	return lang, accent, ptt
}

func firstFrame(appID, lang, accent string, ptt int, chunk []byte) map[string]interface{} {
	return map[string]interface{}{
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
}

func continueFrame(chunk []byte) map[string]interface{} {
	return map[string]interface{}{
		"data": audioDataFrame(statusContinueFrame, chunk),
	}
}

func lastFrame(chunk []byte) map[string]interface{} {
	return map[string]interface{}{
		"data": audioDataFrame(statusLastFrame, chunk),
	}
}
