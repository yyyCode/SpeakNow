package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"speaknow/internal/domain"
)

func BuildCacheKey(audio []byte, opts domain.RecognizeOpts) string {
	h := sha256.New()
	h.Write(audio)
	h.Write([]byte(opts.Language))
	h.Write([]byte(opts.Format))
	fmt.Fprintf(h, "%d", opts.SampleRate)
	h.Write([]byte(fmt.Sprintf("%t", opts.EnablePunc)))

	if len(opts.HotWords) > 0 {
		words := append([]string(nil), opts.HotWords...)
		sort.Strings(words)
		h.Write([]byte(strings.Join(words, ",")))
	}

	return "asr:fp:" + hex.EncodeToString(h.Sum(nil))
}
