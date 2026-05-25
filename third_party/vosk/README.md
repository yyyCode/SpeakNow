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

| 脚本 | 说明 |
|------|------|
| `.\scripts\build-standalone.ps1` | **推荐**：内嵌网页 + 模型 + DLL，生成根目录 `speaknow.exe` |
| `.\scripts\build-windows.ps1` | 内嵌 DLL，模型仍用目录 `model/vosk-model-small-cn-0.22/` |

拉取源码后（需已有 `model/vosk-model-small-cn-0.22/`）：

```powershell
.\scripts\build-standalone.ps1
.\speaknow.exe
```

- **开发者**：`third_party` + CGO 编译即可，无需改本机 Go 模块缓存。
- **最终用户**：只拷贝根目录 `speaknow.exe`；首次运行解压到 `.speaknow-data/`（standalone 版已内嵌模型）。

## 换电脑 / Git

项目已允许将 `third_party/vosk`、`internal/voskruntime/dll` 等一并提交到 Git。  
克隆后执行 `.\scripts\build-standalone.ps1` 生成 exe，或从 Releases 下载 `speaknow.exe` 直接运行。详见根目录 [README.md](../../README.md#快速开始)。

## 说明

Vosk 通过 CGO 调用 **动态库**，无法在 Windows 上完全静态链接进单个 exe；当前方案是 **DLL 嵌入 + 启动时解压到 `vosk-runtime/`**。
