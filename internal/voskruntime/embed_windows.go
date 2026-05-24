//go:build windows && cgo

package voskruntime

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

//go:embed dll/*.dll
var embeddedDLLs embed.FS

var extractOnce sync.Once
var extractDir string
var extractErr error

// extractEmbeddedDLLs unpacks bundled DLLs next to the executable (for end users without GOMODCACHE).
func extractEmbeddedDLLs() (string, error) {
	extractOnce.Do(func() {
		exe, err := os.Executable()
		if err != nil {
			extractErr = err
			return
		}
		dir := filepath.Join(filepath.Dir(exe), "vosk-runtime")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			extractErr = err
			return
		}

		entries, err := fs.ReadDir(embeddedDLLs, "dll")
		if err != nil {
			extractErr = fmt.Errorf("read embedded dll: %w", err)
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			data, err := embeddedDLLs.ReadFile(filepath.Join("dll", e.Name()))
			if err != nil {
				extractErr = err
				return
			}
			dest := filepath.Join(dir, e.Name())
			if err := os.WriteFile(dest, data, 0o644); err != nil {
				extractErr = err
				return
			}
		}
		extractDir = dir
	})
	return extractDir, extractErr
}
