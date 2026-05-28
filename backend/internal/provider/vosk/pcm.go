package vosk

import (
	"encoding/binary"
	"fmt"

	"speaknow/internal/domain"
)

func preparePCM(audio []byte, opts domain.RecognizeOpts) ([]byte, error) {
	switch opts.Format {
	case "pcm", "raw", "l16":
		return audio, nil
	case "wav":
		return extractPCMFromWAV(audio)
	case "webm", "ogg", "mp3", "unknown":
		return nil, fmt.Errorf("vosk requires pcm/wav (16kHz mono 16-bit), got %q; please upload wav/pcm", opts.Format)
	default:
		if len(audio) > 4 && string(audio[:4]) == "RIFF" {
			return extractPCMFromWAV(audio)
		}
		return nil, fmt.Errorf("unsupported audio format %q for vosk", opts.Format)
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

func estimateDuration(pcmLen int) float64 {
	if pcmLen == 0 {
		return 0
	}
	return float64(pcmLen) / 32000.0
}
