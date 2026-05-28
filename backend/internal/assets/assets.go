package assets

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

//go:embed default.yaml
var defaultConfig []byte

// 由 scripts/prepare-bundle.ps1 在发布构建前填充；开发时可为空，走本地 web/model 目录。
//
//go:embed all:web
var webFS embed.FS

//go:embed all:model
var modelFS embed.FS

var (
	prepareOnce sync.Once
	prepareErr  error
	modelPath   string
	useEmbedded bool
)

// Prepare 解压嵌入的语音模型（若有），并记录模型目录。
func Prepare() error {
	prepareOnce.Do(func() {
		if !hasEmbeddedModel() {
			modelPath = filepath.Join("model", "vosk-model-small-cn-0.22")
			return
		}
		useEmbedded = true
		dir, err := extractModel()
		if err != nil {
			prepareErr = err
			return
		}
		modelPath = dir
	})
	return prepareErr
}

func ModelPath() string {
	if modelPath != "" {
		return modelPath
	}
	return filepath.Join("model", "vosk-model-small-cn-0.22")
}

func UseEmbeddedModel() bool { return useEmbedded }

func DefaultConfig() []byte { return defaultConfig }

func WebHandler() (http.Handler, error) {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		return nil, err
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		return http.FileServer(http.Dir("../frontend/web")), nil
	}
	return http.FileServer(http.FS(sub)), nil
}

func hasEmbeddedModel() bool {
	entries, err := fs.ReadDir(modelFS, "model")
	if err != nil || len(entries) == 0 {
		return false
	}
	_, err = fs.Stat(modelFS, "model/vosk-model-small-cn-0.22/conf/model.conf")
	return err == nil
}

func extractModel() (string, error) {
	root, err := cacheRoot()
	if err != nil {
		return "", err
	}
	dest := filepath.Join(root, "vosk-model-small-cn-0.22")
	if _, err := os.Stat(filepath.Join(dest, "conf", "model.conf")); err == nil {
		return dest, nil
	}
	if err := os.RemoveAll(dest); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", err
	}
	const prefix = "model/vosk-model-small-cn-0.22"
	err = fs.WalkDir(modelFS, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(prefix, path)
		if rel == "." {
			return nil
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		data, err := modelFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	return dest, err
}

func cacheRoot() (string, error) {
	if d := os.Getenv("SPEAKNOW_DATA_DIR"); d != "" {
		return d, os.MkdirAll(d, 0o755)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Join(filepath.Dir(exe), ".speaknow-data")
		return dir, os.MkdirAll(dir, 0o755)
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "SpeakNow")
	return dir, os.MkdirAll(dir, 0o755)
}
