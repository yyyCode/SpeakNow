// 纯 Go 启动器：将 CGO 主程序与 Vosk DLL 内嵌进单个 speaknow.exe，首次运行解压到 .speaknow-data/runtime。
package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed payload/*
var payload embed.FS

const coreName = "speaknow-core.exe"

func main() {
	if runtime.GOOS != "windows" {
		fmt.Fprintln(os.Stderr, "launcher: Windows only")
		os.Exit(1)
	}

	exe, err := os.Executable()
	if err != nil {
		fatal(err)
	}
	exeDir, err := filepath.Abs(filepath.Dir(exe))
	if err != nil {
		fatal(err)
	}

	dataDir := filepath.Join(exeDir, ".speaknow-data")
	runtimeDir := filepath.Join(dataDir, "runtime")
	if err := extractPayload(runtimeDir); err != nil {
		fatal(err)
	}

	core := filepath.Join(runtimeDir, coreName)
	cmd := exec.Command(core, os.Args[1:]...)
	cmd.Dir = exeDir
	cmd.Env = append(os.Environ(), "SPEAKNOW_DATA_DIR="+dataDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() != 0 {
			os.Exit(exit.ExitCode())
		}
		fatal(err)
	}
}

func extractPayload(dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	return fs.WalkDir(payload, "payload", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel("payload", path)
		if err != nil {
			return err
		}
		rel = filepath.FromSlash(rel)

		data, err := payload.ReadFile(path)
		if err != nil {
			return err
		}
		out := filepath.Join(dest, rel)
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if sameFile(out, data) {
			return nil
		}
		return os.WriteFile(out, data, 0o644)
	})
}

func sameFile(path string, want []byte) bool {
	st, err := os.Stat(path)
	if err != nil || int64(len(want)) != st.Size() {
		return false
	}
	got, err := os.ReadFile(path)
	return err == nil && string(got) == string(want)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "speaknow: %v\n", err)
	pauseIfInteractive()
	os.Exit(1)
}

func pauseIfInteractive() {
	if os.Getenv("SPEAKNOW_NO_PAUSE") != "" {
		return
	}
	// 双击 exe 时通常无参数，便于看到错误信息
	if len(os.Args) > 1 {
		return
	}
	fmt.Fprint(os.Stderr, "\n按回车键退出…")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

