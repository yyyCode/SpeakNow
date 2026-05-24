package vosk

import (
	"encoding/json"
	"strings"
)

func parseVoskText(jsonStr string) string {
	if jsonStr == "" {
		return ""
	}
	var res struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &res); err != nil {
		return ""
	}
	return strings.TrimSpace(res.Text)
}
