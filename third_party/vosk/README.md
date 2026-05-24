# Vosk 原生库（项目内，不依赖 GOMODCACHE）

**不再使用** `D:\go\pkg\mod\github.com\alphacep\vosk-api\src`。

## 一键准备依赖

项目根目录执行（优先使用 `third_party/vosk-win64.zip`，否则从官方下载）：

```powershell
.\scripts\setup-vosk.ps1
```

官方包地址：

https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-win64-0.3.45.zip

解压结果：

| 路径 | 用途 |
|------|------|
| `windows-amd64/include/` | 编译：头文件 |
| `windows-amd64/lib/` | 编译：libvosk.lib |
| `windows-amd64/bin/` | 开发运行：DLL |
| `internal/voskruntime/dll/` | 编译进 exe（go:embed） |

## 构建与分发

```powershell
.\scripts\build-windows.ps1
```

- **开发者**：`third_party` + `go build` 即可，无需改本机 Go 模块缓存。
- **最终用户**：
  - 方式 A：只拷贝 `dist/speaknow/` 整个文件夹（exe + dll，最简单可靠）
  - 方式 B：只拷贝单个 `speaknow.exe`（内含嵌入 DLL，首次运行解压到同目录 `vosk-runtime/`）
  - 语音模型 `model/vosk-model-small-cn-0.22` 需单独放置（体积大，不打进 exe）

## 换电脑 / Git

项目已允许将 `third_party/vosk`、`internal/voskruntime/dll`、`speaknow.exe` 等一并提交到 Git。  
克隆到另一台电脑后可直接运行根目录的 `speaknow.exe`，或执行 `.\scripts\build-standalone.ps1` 重新构建。

## 说明

Vosk 通过 CGO 调用 **动态库**，无法在 Windows 上完全静态链接进单个 exe；当前方案是 **DLL 嵌入 + 启动时解压**，或 **dist 目录随 exe 附带 DLL**。
