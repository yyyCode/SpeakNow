package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"speaknow/internal/domain"
	"speaknow/internal/service/router"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	router *router.Service
}

func NewWSHandler(routerSvc *router.Service) *WSHandler {
	return &WSHandler{router: routerSvc}
}

func (h *WSHandler) Stream(c *gin.Context) {
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	opts := domain.RecognizeOpts{
		Language:   c.DefaultQuery("language", "zh-CN"),
		Format:     "pcm",
		SampleRate: 16000,
		EnablePunc: c.DefaultQuery("enable_punc", "true") == "true",
		Provider:   c.DefaultQuery("provider", "auto"),
	}

	audioIn := make(chan []byte, 64)
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	var closeOnce sync.Once
	closeAudio := func() {
		closeOnce.Do(func() { close(audioIn) })
	}

	results, err := h.router.StreamRecognize(ctx, audioIn, opts)
	if err != nil {
		writeWSEvent(conn, domain.StreamResult{Type: "error", Message: err.Error()})
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for r := range results {
			if writeErr := writeWSEvent(conn, r); writeErr != nil {
				cancel()
				return
			}
			if r.Type == "final" || r.Type == "error" {
				return
			}
		}
	}()

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		switch msgType {
		case websocket.BinaryMessage:
			select {
			case audioIn <- data:
			case <-ctx.Done():
				break
			}
		case websocket.TextMessage:
			var cmd struct {
				Action string `json:"action"`
			}
			if json.Unmarshal(data, &cmd) == nil && cmd.Action == "end" {
				closeAudio()
				wg.Wait()
				return
			}
		}
	}

	closeAudio()
	wg.Wait()
}

func writeWSEvent(conn *websocket.Conn, event domain.StreamResult) error {
	return conn.WriteJSON(event)
}
