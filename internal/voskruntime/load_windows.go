//go:build windows && cgo

package voskruntime

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

var modKernel32 = syscall.NewLazyDLL("kernel32.dll")
var procSetDllDirectoryW = modKernel32.NewProc("SetDllDirectoryW")

// Ensure must run before the vosk CGO package loads (import voskruntime first in provider/vosk).
func Ensure() error {
	dir, err := resolveBinDir()
	if err != nil {
		return err
	}
	if dir == "" {
		return nil
	}
	path, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return err
	}
	r, _, e := procSetDllDirectoryW.Call(uintptr(unsafe.Pointer(path)))
	if r == 0 {
		return e
	}
	return nil
}

func resolveBinDir() (string, error) {
	// 1) 嵌入 DLL 解压到 exe 旁 vosk-runtime（分发单 exe 时使用）
	if dir, err := extractEmbeddedDLLs(); err == nil && dir != "" {
		if _, err := os.Stat(filepath.Join(dir, "libvosk.dll")); err == nil {
			return dir, nil
		}
	}

	// 2) 开发/源码：项目 third_party；dist 打包：DLL 与 exe 同目录
	rel := filepath.Join("third_party", "vosk", "windows-amd64", "bin")
	candidates := []string{rel}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			exeDir,
			filepath.Join(exeDir, rel),
			filepath.Join(exeDir, "vosk-runtime"),
			filepath.Join(exeDir, "vosk", "bin"),
		)
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, rel))
	}

	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(filepath.Join(abs, "libvosk.dll")); err == nil {
			return abs, nil
		}
	}
	return "", nil
}
