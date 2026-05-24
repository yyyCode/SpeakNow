<div align="center">

# 🎤 SpeakNow

**免费在线语音转文字 — 讯飞云端 + Vosk 本地离线 · 实时流式 · 单文件 exe 分发**

*A Go voice input backend with streaming ASR, dual engines (iFlytek cloud + Vosk offline), and optional single-exe distribution.*

<br/>

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![ASR](https://img.shields.io/badge/ASR-讯飞%20%7C%20Vosk-purple?style=flat)](#识别引擎)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-lightgrey?style=flat)](#快速开始)
[![Release](https://img.shields.io/badge/Download-Release-blue?style=flat)](docs/RELEASE.md)

<br/>

[快速开始](#快速开始) · [识别引擎](#识别引擎) · [单文件 exe](#单文件-exewindows-分发) · [API](#api-文档) · [English](#english) · [License](#开源协议)

</div>

---

SpeakNow 是一个基于 Go 的**语音输入法后端服务**，配套 Web 演示页，支持**实时流式语音识别**与离线文件识别。项目采用多厂商 ASR 抽象、智能路由降级、Redis 缓存与限流成本控制；并支持 **Vosk 本地离线识别**、前端切换讯飞 / 本地引擎，以及 **单文件 exe** 分发。

---

## 项目亮点

| 亮点 | 说明 |
|------|------|
| **双引擎识别** | **讯飞云端**（高精度、需密钥）+ **Vosk 本地**（离线、免 API、中文小模型） |
| **前端一键切换** | Web 页可选择「自动 / 讯飞 / Vosk」，偏好写入 `localStorage` |
| **实时流式识别** | 麦克风 → Web Audio 16kHz PCM → WebSocket → 边说边出字 |
| **离线文件识别** | 上传 WAV/PCM，`POST /api/v1/asr/recognize` |
| **Vosk 依赖项目内嵌** | `third_party/vosk`，不依赖 `GOMODCACHE` 手工拷 DLL |
| **单文件 exe 分发** | 内嵌网页、中文模型、Vosk 库；用户仅下载 exe 即可运行（见 [发布说明](docs/RELEASE.md)） |
| **多厂商抽象** | 统一 `Provider` / `StreamRecognizer` 接口，预留阿里云 / 腾讯云 |
| **智能路由降级** | `provider=auto` 时主 Provider 不可用自动 fallback |
| **双层限流** | 本地 Token Bucket + Redis 分布式计数 |
| **结果缓存** | 相同音频指纹命中 Redis（流式路径不经缓存） |
| **成本统计** | 记录时长、延迟、缓存命中与费用 |
| **开箱即用 UI** | `Ctrl+Space` 快捷键、自动复制、历史记录 |
| **容器化 / systemd** | Docker、Linux 一键安装脚本 |

---

## 识别引擎

| 引擎 | 标识 | 流式 | 离线 | 说明 |
|------|------|------|------|------|
| 讯飞 | `xunfei` | ✅ | ✅ | 需配置 APPID / APIKey / APISecret |
| Vosk 本地 | `vosk` | ✅ | ✅ | 使用 `model/vosk-model-small-cn-0.22`，仅推荐 `zh-CN` |
| Mock | `mock` | ✅ | ✅ | 无密钥开发调试 |
| 自动 | `auto` | ✅ | ✅ | 按 `asr.primary` + `fallback` 选择 |

Web 演示页与 API 均支持 `provider` 参数（见下方 API 文档）。

---

## 功能概览

```
浏览器 / 客户端
    │
    ├─ WebSocket  /api/v1/asr/stream?provider=vosk|xunfei|auto
    │
    └─ REST POST  /api/v1/asr/recognize
                │
                ▼
         [Gin] 限流 · 日志 · CORS · 静态 Web（可内嵌）
                │
                ▼
         [ASR] 缓存 · 路由 · 成本统计
                │
                ▼
    ┌───────────┼───────────┬───────────┐
    ▼           ▼           ▼           ▼
  讯飞        Vosk 本地     Mock      阿里云/腾讯
 (云端)      (离线)       (调试)      (预留)
```

---

## 快速开始

### 环境要求

| 场景 | 要求 |
|------|------|
| **开发 / 源码运行** | Go 1.23+、**CGO**（启用 Vosk 时）、Windows 需 MinGW-w64 |
| **仅运行 exe** | 无（下载 [Release](docs/RELEASE.md) 中的 `speaknow.exe` 即可） |
| **可选** | Redis（未连接时自动降级无缓存） |
| **Web 录音** | 浏览器麦克风权限 |

### 1. 克隆与配置

```bash
git clone https://github.com/yyyCode/SpeakNow.git
cd SpeakNow

copy configs\config.example.yaml configs\config.yaml   # Windows
# cp configs/config.example.yaml configs/config.yaml   # Linux/macOS
```

**使用讯飞时**，编辑 `configs/config.yaml` 填入密钥：

```yaml
asr:
  primary: "xunfei"
  fallback: ["mock"]

providers:
  xunfei:
    enabled: true
    app_id: "your-app-id"
    api_key: "your-api-key"
    api_secret: "your-api-secret"
  vosk:
    enabled: true
    model_path: "model/vosk-model-small-cn-0.22"
```

**仅本地 Vosk** 可将 `asr.primary` 设为 `vosk`，并关闭讯飞（参见 `internal/assets/default.yaml` 单文件版默认配置）。

> `configs/config.yaml` 已 `.gitignore`，勿提交密钥。

### 2. 准备 Vosk 原生库（源码编译时）

依赖已放在 **`third_party/vosk/`**，无需再改 `GOMODCACHE`。

若目录不完整，在项目根执行：

```powershell
# 优先使用 third_party/vosk-win64.zip，否则从 GitHub 下载
.\scripts\setup-vosk.ps1
```

官方包：[vosk-win64-0.3.45.zip](https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-win64-0.3.45.zip)

### 3. 准备语音模型

中文小模型目录（约 65MB）：

```text
model/vosk-model-small-cn-0.22/
```

可从 [Vosk 模型列表](https://alphacephei.com/vosk/models) 下载解压到 `model/`。

### 4. 启动服务（开发模式）

```powershell
$env:CGO_ENABLED=1
go run ./cmd/server -config configs/config.yaml
```

未传 `-config` 时使用内置默认配置（适合单文件构建）。

浏览器访问：**http://localhost:8080/web/**

### 5. 使用 Web 演示页

1. **识别引擎**：选择「自动 / 讯飞云端 / Vosk 本地」
2. **识别语言**：Vosk 仅支持中文普通话（`zh-CN`）
3. 点击 **开始录音** 或 **`Ctrl + Space`**
4. 停止后获得最终结果；可开启自动复制、查看历史记录
5. 支持 **上传 WAV/PCM** 离线识别

---

## 单文件 exe（Windows 分发）

将 **网页 + 中文模型 + Vosk DLL** 打进一个 exe，最终用户**无需**再下载模型或配置 Go 模块缓存。

```powershell
.\scripts\build-standalone.ps1
```

生成根目录 **`speaknow.exe`**（约 120MB+）。首次运行会在 exe 旁生成 `.speaknow-data`（解压内嵌模型），属正常现象。

| 事项 | 说明 |
|------|------|
| **不要**把 exe 提交 Git | 体积过大，见 [docs/RELEASE.md](docs/RELEASE.md) |
| **推荐** | 在 GitHub **Releases** 上传 `speaknow.exe` 供下载 |
| **换机开发** | `git clone` 后执行 `build-standalone.ps1` 或 `go build`（需 CGO） |

内置默认：仅启用 Vosk、`primary: vosk`、监听 `0.0.0.0:8080`。

---

## 脚本说明

| 脚本 | 作用 |
|------|------|
| `scripts/setup-vosk.ps1` | 解压/下载 Vosk Win64 库到 `third_party/vosk` 与 `internal/voskruntime/dll` |
| `scripts/prepare-bundle.ps1` | 将 `web/`、`model/` 复制到 `internal/assets` 供 `go:embed` |
| `scripts/build-standalone.ps1` | 一键构建单文件 `speaknow.exe` |
| `scripts/build-windows.ps1` | 构建 exe 并组装 `dist/speaknow/` 目录（exe + dll + web） |

---

## 关闭 / 重启服务

### Windows

```powershell
netstat -ano | findstr :8080
taskkill /PID <PID> /F
# 或
taskkill /IM speaknow.exe /F
```

### Linux（systemd）

```bash
sudo systemctl stop speaknow
sudo systemctl restart speaknow
journalctl -u speaknow -f
```

---

## Linux 部署（systemd）

> **注意**：Linux 安装脚本默认 `CGO_ENABLED=0`，**不包含 Vosk**。若需在 Linux 使用 Vosk，需自行准备 `third_party/vosk/linux-amd64` 并 `CGO_ENABLED=1` 编译。

```bash
cp configs/config.example.yaml configs/config.yaml
sudo bash deployments/install-linux.sh
sudo systemctl start speaknow
```

详见上文「安装路径」与 `deployments/install-linux.sh`。

---

## API 文档

### 实时流式识别（WebSocket）

```
GET /api/v1/asr/stream?language=zh-CN&enable_punc=true&provider=auto
```

| Query | 说明 |
|-------|------|
| `language` | `zh-CN` / `en-US` 等（Vosk 仅 `zh-CN`） |
| `enable_punc` | `true` / `false` |
| `provider` | `auto` / `xunfei` / `vosk` / `mock` |

| 方向 | 格式 | 说明 |
|------|------|------|
| 客户端 → 服务端 | 二进制 | PCM（16kHz, 16-bit, mono） |
| 客户端 → 服务端 | JSON | `{"action":"end"}` 结束录音 |
| 服务端 → 客户端 | JSON | `{"type":"partial","text":"...","provider":"vosk"}` |
| 服务端 → 客户端 | JSON | `{"type":"final","text":"...","provider":"vosk"}` |
| 服务端 → 客户端 | JSON | `{"type":"error","message":"..."}` |

### 离线文件识别（REST）

```bash
curl -X POST http://localhost:8080/api/v1/asr/recognize \
  -F "file=@test.wav" \
  -F "language=zh-CN" \
  -F "enable_punc=true" \
  -F "provider=vosk"
```

### 其他接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查（Redis 可选） |
| GET | `/api/v1/providers/status` | 各引擎状态（`healthy` / `unhealthy`） |
| GET | `/api/v1/stats/cost` | 成本统计 |

---

## Docker 部署

```bash
docker run -d --name speaknow-redis -p 6379:6379 redis:7-alpine
cd deployments && docker compose up -d --build
```

Docker 镜像默认面向讯飞云端场景；本地 Vosk 建议使用 Windows 单文件 exe 或自行扩展镜像。

---

## 项目结构

```
SpeakNow/
├── cmd/server/                 # 服务入口（支持内嵌配置 / 静态资源）
├── internal/
│   ├── assets/                 # go:embed 默认配置、web、模型（prepare-bundle 生成）
│   ├── voskruntime/            # Windows DLL 搜索路径 / 内嵌解压
│   ├── config/                 # Viper 配置
│   ├── handler/                # HTTP / WebSocket
│   ├── middleware/
│   ├── provider/
│   │   ├── xunfei/             # 讯飞流式 + 离线
│   │   ├── vosk/               # Vosk 本地流式 + 离线
│   │   ├── mock/
│   │   └── factory/
│   └── service/                # ASR、缓存、路由、成本
├── third_party/
│   └── vosk/
│       ├── bindings/           # 本地 CGO 绑定（replace 覆盖上游 go 模块）
│       └── windows-amd64/      # include / lib / bin
├── pkg/
├── web/                        # 前端演示页（含识别引擎切换）
├── model/                      # Vosk 中文模型（需自行下载或随仓库提供）
├── scripts/                    # setup-vosk、prepare-bundle、build-standalone
├── configs/
├── docs/
│   └── RELEASE.md              # GitHub Release 发布指南
└── deployments/
```

---

## 讯飞 ASR 配置

1. 前往 [讯飞开放平台](https://www.xfyun.cn/) 注册并创建应用  
2. 开通 **语音听写（流式版）**  
3. 将 APPID、APIKey、APISecret 写入 `configs/config.yaml`

**音频格式**：16kHz、16-bit、单声道 PCM；上传支持 `.wav` / `.pcm`。

参考：`iat_ws_go_demo/`

---

## Vosk 本地识别配置

```yaml
providers:
  vosk:
    enabled: true
    model_path: "model/vosk-model-small-cn-0.22"
    sample_rate: 16000
    cost_per_second: 0
```

| 项 | 说明 |
|----|------|
| `model_path` | Kaldi 模型目录，需含 `conf/model.conf` |
| 语言 | 当前小模型仅 **中文普通话** |
| 编译 | 必须 `CGO_ENABLED=1`，Windows 需 MinGW |
| 依赖 | `third_party/vosk/windows-amd64`，见 [third_party/vosk/README.md](third_party/vosk/README.md) |

---

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.23 |
| Web | Gin |
| WebSocket | gorilla/websocket |
| 本地 ASR | Vosk（CGO，`third_party/vosk/bindings`） |
| 云端 ASR | 讯飞 IAT WebSocket |
| 配置 | Viper |
| 日志 | Zap |
| 缓存 / 限流 | Redis（可选） |
| 前端 | 原生 HTML + Web Audio API |
| 分发 | go:embed + GitHub Releases |

---

## 开发路线

- [x] Phase 1：脚手架、Mock、REST、缓存、限流、Web 演示  
- [x] Phase 2：讯飞流式、实时前端、智能路由  
- [x] Phase 2.5：**Vosk 本地识别**、前端引擎切换、项目内嵌原生库  
- [x] Phase 2.6：**单文件 exe**（内嵌 web / 模型 / 默认配置）、Release 发布文档  
- [ ] Phase 3：阿里云 / 腾讯云、API Key 鉴权、Prometheus、PostgreSQL  

---

## English

**SpeakNow** is a Go backend for voice-to-text with a built-in web UI. It supports:

- **Streaming ASR** over WebSocket (16 kHz PCM)
- **Dual engines**: iFlytek cloud (`xunfei`) and local **Vosk** (`vosk`)
- **Offline upload** via `POST /api/v1/asr/recognize`
- Optional **standalone Windows exe** (embedded web + model + runtime libs) — see [docs/RELEASE.md](docs/RELEASE.md)

```bash
git clone https://github.com/yyyCode/SpeakNow.git
cd SpeakNow
go run ./cmd/server -config configs/config.yaml
# Open http://localhost:8080/web/
```

---

## 开源协议

| 范围 | 协议 |
|------|------|
| **本仓库代码**（SpeakNow） | [MIT License](LICENSE) |
| **Vosk 库 / API**（`third_party/vosk`） | [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)（第三方） |
| **讯飞服务**（可选） | 须遵守[讯飞开放平台](https://www.xfyun.cn/)用户协议 |

- 根目录 **[LICENSE](LICENSE)**：MIT 全文。提交到 GitHub 后，仓库顶部会出现 **「MIT license」** 标签页。  
- **[NOTICE](NOTICE)**：第三方组件声明（分发单文件 exe 时建议一并保留）。  
- README 徽章：`[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)`（由 [Shields.io](https://shields.io/) 生成）。

若希望**整个项目**改用 Apache 2.0，可将 `LICENSE` 替换为 Apache 2.0 全文，并更新 README 徽章为 `License-Apache%202.0-blue.svg`；MIT 与 Apache 2.0 **不要混写在同一 LICENSE 文件里**，选其一即可。
