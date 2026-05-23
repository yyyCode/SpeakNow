package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"speaknow/internal/domain"
	"speaknow/internal/service/asr"
	"speaknow/internal/service/cost"
	"speaknow/internal/service/router"
	"speaknow/pkg/response"
)

type ASRHandler struct {
	asr    *asr.Service
	router *router.Service
	cost   *cost.Service
	maxSize int64
}

func NewASRHandler(asrSvc *asr.Service, routerSvc *router.Service, costSvc *cost.Service, maxSize int64) *ASRHandler {
	return &ASRHandler{
		asr:     asrSvc,
		router:  routerSvc,
		cost:    costSvc,
		maxSize: maxSize,
	}
}

func (h *ASRHandler) Recognize(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "missing audio file: "+err.Error())
		return
	}
	if file.Size > h.maxSize {
		response.BadRequest(c, "audio file too large")
		return
	}

	f, err := file.Open()
	if err != nil {
		response.InternalError(c, "open file failed")
		return
	}
	defer f.Close()

	audio, err := io.ReadAll(io.LimitReader(f, h.maxSize+1))
	if err != nil {
		response.InternalError(c, "read file failed")
		return
	}

	opts := domain.RecognizeOpts{
		Language:   c.DefaultPostForm("language", "zh-CN"),
		Format:     detectFormat(file.Filename, c.PostForm("format")),
		SampleRate: 16000,
		EnablePunc: c.DefaultPostForm("enable_punc", "true") == "true",
		Provider:   c.DefaultPostForm("provider", "auto"),
	}

	if hotWords := c.PostForm("hot_words"); hotWords != "" {
		opts.HotWords = strings.Split(hotWords, ",")
	}

	userID := c.GetString("user_id")
	result, err := h.asr.Recognize(c.Request.Context(), userID, audio, opts)
	if err != nil {
		response.ServiceUnavailable(c, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *ASRHandler) ProviderStatus(c *gin.Context) {
	response.OK(c, h.router.ProviderStatus(c.Request.Context()))
}

func (h *ASRHandler) CostStats(c *gin.Context) {
	response.OK(c, h.cost.Summary())
}

func detectFormat(filename, override string) string {
	if override != "" {
		return override
	}
	idx := strings.LastIndex(filename, ".")
	if idx < 0 {
		return "unknown"
	}
	return strings.ToLower(filename[idx+1:])
}

type HealthHandler struct {
	checkers map[string]func(*gin.Context) error
}

func NewHealthHandler(checkers map[string]func(*gin.Context) error) *HealthHandler {
	return &HealthHandler{checkers: checkers}
}

func (h *HealthHandler) Health(c *gin.Context) {
	status := make(map[string]string)
	allOK := true
	for name, check := range h.checkers {
		if err := check(c); err != nil {
			status[name] = "unhealthy: " + err.Error()
			allOK = false
		} else {
			status[name] = "healthy"
		}
	}

	code := http.StatusOK
	if !allOK {
		code = http.StatusServiceUnavailable
	}
	c.JSON(code, gin.H{
		"code":   0,
		"status": map[bool]string{true: "ok", false: "degraded"}[allOK],
		"checks": status,
	})
}
