从 https://github.com/alphacep/vosk-api/releases 下载 Linux 包后，解压为：

- `include/vosk_api.h`
- `lib/libvosk.so`（及链接所需文件）
- `bin/` 下运行时 `.so`（若发行包单独提供）

然后 `CGO_ENABLED=1 go build ./cmd/server`。
