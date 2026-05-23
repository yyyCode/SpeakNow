package domain

import "time"

type RecognizeOpts struct {
	Language   string
	Format     string
	SampleRate int
	HotWords   []string
	EnablePunc bool
	Provider   string
}

type Segment struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type Result struct {
	Text       string    `json:"text"`
	Confidence float64   `json:"confidence"`
	Duration   float64   `json:"duration"`
	Provider   string    `json:"provider"`
	Segments   []Segment `json:"segments,omitempty"`
}

type StreamResult struct {
	Type       string  `json:"type"` // partial | final | error
	Text       string  `json:"text,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
	Message    string  `json:"message,omitempty"`
}

type CallLog struct {
	ID              uint      `json:"id"`
	UserID          string    `json:"user_id"`
	Provider        string    `json:"provider"`
	AudioDurationMs int       `json:"audio_duration_ms"`
	CostYuan        float64   `json:"cost_yuan"`
	LatencyMs       int       `json:"latency_ms"`
	CacheHit        bool      `json:"cache_hit"`
	TextLength      int       `json:"text_length"`
	Status          string    `json:"status"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type CachedResult struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Provider   string  `json:"provider"`
	Duration   float64 `json:"duration"`
	CreatedAt  int64   `json:"created_at"`
}
