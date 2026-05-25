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

Windows 本地 Vosk 为主；Linux 见 [Linux 部署](#linux-部署systemd)。

### 怎么选？

| 方式 | 适合谁 | 是否需要 Go |
|------|--------|-------------|
| [方式一](#方式一直接运行-exe) | 只拿到别人发的 / Releases 里的 `speaknow.exe` | 否 |
| **[方式二](#方式二拉取源码并构建-exewindows)**（**推荐**） | **`git clone` 拉下本仓库的用户** | 是 |
| [方式三](#方式三源码开发运行) | 改 Go 代码、联调 API | 是 |

> **建议**：只要你已经 `git clone` 了本仓库，**优先走 [方式二](#方式二拉取源码并构建-exewindows)**，在本机打出与当前源码一致的 `speaknow.exe`，比单独下载的旧版 Release 更可靠。  
> **若方式一的 exe 打不开、闪退、端口无响应、识别一直失败**，不要反复换下载链接，**请直接按 [方式二](#方式二拉取源码并构建-exewindows) 从源码构建**；构建过程里的报错也更容易排查。

### 环境要求

| 场景 | 要求 |
|------|------|
| **仅运行 exe（方式一）** | Windows 10/11，无需安装 Go |
| **构建 exe（方式二）** | Go **1.23+**、**CGO**、**MinGW-w64**（`gcc` 在 PATH 中）、PowerShell |
| **开发运行（方式三）** | 同方式二 |
| **可选** | Redis（未连接时自动降级无缓存） |
| **Web 录音** | 浏览器允许麦克风；建议 Chrome / Edge |

服务默认监听 **`0.0.0.0:8080`**。启动成功后控制台会出现类似 `SpeakNow server starting`、`addr` 等日志。演示页：

- **http://127.0.0.1:8080/web/**
- 或 **http://localhost:8080/web/**

根路径 `/` 会自动跳转到 `/web/`。

---

### 方式一：直接运行 exe

适合：**没有**克隆仓库、只有单个 `speaknow.exe`（例如 [GitHub Releases](docs/RELEASE.md) 附件）。

#### 1. 放置与启动

1. 将 **`speaknow.exe`** 放到任意目录（**不需要**同目录再放 DLL 或模型文件夹）。
2. 启动（任选其一）：
   - **双击** `speaknow.exe`（会弹出黑色控制台窗口，不要关）
   - 或在 PowerShell 中（便于看到报错）：

```powershell
cd D:\path\to\exe所在目录
.\speaknow.exe
```

3. 等控制台出现服务已启动的日志（首次可能要 **几十秒～数分钟** 解压内嵌资源）。
4. 浏览器打开：**http://127.0.0.1:8080/web/**

#### 2. 首次运行会发生什么

| 现象 | 说明 |
|------|------|
| 同目录出现 **`.speaknow-data`** | 正常。launcher 把内嵌的 `speaknow-core.exe`、Vosk DLL、语音模型解压到这里 |
| 目录体积很大（数百 MB） | 正常，内含中文 Vosk 模型 |
| Windows 防火墙提示 | 选择允许专用网络，否则本机浏览器可能连不上 8080 |

**不要**在程序正在运行时删除 `.speaknow-data`。若怀疑解压损坏，先 `taskkill /IM speaknow.exe /F`，再删掉整个 `.speaknow-data` 后重新运行 exe 让它重新解压。

#### 3. 可选：指定配置文件

默认使用 exe **内嵌**配置（仅 Vosk、`primary: vosk`、监听 `0.0.0.0:8080`）。若要改用讯飞等：

```powershell
.\speaknow.exe -config D:\path\to\config.yaml
```

可参考 `configs/config.example.yaml` 自行编写。

#### 4. 停止服务

- 关闭 exe 的控制台窗口，或
- `taskkill /IM speaknow.exe /F`（见 [关闭 / 重启服务](#关闭--重启服务)）

#### 5. 方式一 exe 不行？→ 请用方式二

下列情况**不要继续纠结下载的 exe**，请改做 **[方式二：拉取源码并构建 exe](#方式二拉取源码并构建-exewindows)**（你已有仓库时，直接在项目根执行 `.\scripts\build-standalone.ps1` 即可）：

- 双击后**窗口一闪就关**、没有任何日志
- 控制台有报错（缺 DLL、无法加载模型、`prepare assets` 失败等）
- 浏览器访问 8080 **连接被拒绝**（确认防火墙、且 8080 未被占用：`netstat -ano | findstr :8080`）
- 页面能开但 **Vosk 识别无结果 / 一直报错**（多半是 Release 与当前代码或模型不一致）

方式二会在你本机用**当前源码 + 当前 `model/`** 重新打包，成功率最高。若方式二构建也失败，把 PowerShell **完整报错**对照下文 [常见问题](#常见问题) 处理。

---

### 方式二：拉取源码并构建 exe（Windows）【推荐】

适合：已 `git clone` 本仓库，或 [方式一下载的 exe 无法使用](#5-方式一-exe-不行请用方式二) 时。  
产物：项目根目录 **`speaknow.exe`**（内嵌 **网页 + 中文模型 + Vosk DLL + 默认配置**），分发时只拷贝这一个文件即可。

#### 0. 构建前：安装工具链（仅首次）

1. **Go 1.23+**  
   - 安装后执行 `go version` 确认。
2. **MinGW-w64（提供 `gcc`）**  
   - 安装后执行 `gcc --version` 确认在 PATH 中。  
   - 常见安装：MSYS2 里安装 `mingw-w64-x86_64-gcc`，或将 MinGW 的 `bin` 加入系统 PATH。
3. **Git**（用于 `git clone`）

构建全程在 **PowerShell**、**项目根目录**进行。

#### 1. 克隆仓库

```powershell
git clone https://github.com/yyyCode/SpeakNow.git
cd SpeakNow
```

若你已经在仓库目录，跳过此步。

#### 2. 准备语音模型（必做，否则构建会失败）

`build-standalone.ps1` 会把模型打进 exe，必须先有目录：

```text
model/vosk-model-small-cn-0.22/
    conf/model.conf          ← 用于判断模型是否完整
    ...
```

- 若 `git clone` 后仓库里**已有** `model/`（体积大，视仓库而定），检查上述 `conf/model.conf` 是否存在即可。
- 若没有：打开 [Vosk 模型列表](https://alphacephei.com/vosk/models)，下载 **Chinese small**（`vosk-model-small-cn-0.22`），解压到 `model/`，保证路径与上表一致。

#### 3. 一键构建

在**项目根目录**执行（内部顺序：`setup-vosk.ps1` → `prepare-bundle.ps1` → 编译 `cmd/server` → 编译 `cmd/launcher`）：

```powershell
.\scripts\build-standalone.ps1
```

**脚本会做什么（了解即可，一般不用手跑子脚本）：**

| 步骤 | 脚本 / 动作 | 说明 |
|------|-------------|------|
| 1 | `setup-vosk.ps1` | 准备 `third_party/vosk/windows-amd64`；缺库时从 `third_party/vosk-win64.zip` 或 [官方 zip](https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-win64-0.3.45.zip) 下载 |
| 2 | `prepare-bundle.ps1` | 复制 `web/`、`model/vosk-model-small-cn-0.22` → `internal/assets`，并写入内嵌默认 `default.yaml` |
| 3 | `go build ./cmd/server` | 生成 `cmd/launcher/payload/speaknow-core.exe`（CGO，需 `gcc`） |
| 4 | 复制 `*.dll` 到 payload | Vosk 运行时 |
| 5 | `go build ./cmd/launcher` | 生成根目录 **`speaknow.exe`** |

成功时终端会显示类似：`Done: speaknow.exe (xxx MB)`。

| 其它脚本 | 何时用 |
|----------|--------|
| `scripts/build-standalone.ps1` | **默认用这个**（模型也打进 exe） |
| `scripts/build-windows.ps1` | 不想把模型打进 exe、愿在运行目录保留 `model/` 时用 |
| 单独跑 `setup-vosk.ps1` / `prepare-bundle.ps1` | 仅当 `build-standalone` 某步失败、需分步重试时 |

**不要**把生成的 `speaknow.exe` 提交 Git，见 [docs/RELEASE.md](docs/RELEASE.md)。

#### 4. 启动本机构建的 exe

仍在**项目根目录**（或把 `speaknow.exe` 拷到别处后进入该目录）：

```powershell
.\speaknow.exe
```

1. 保持控制台窗口打开，等待日志中出现服务已监听 `8080`。
2. 浏览器打开：**http://127.0.0.1:8080/web/**
3. 首次运行同样会生成 **`.speaknow-data`**（与方式一相同）。

可选：`.\speaknow.exe -config configs\config.yaml`（需自行准备 yaml；内嵌默认已可离线 Vosk）。

#### 5. 方式二构建常见报错

| 报错 / 现象 | 处理 |
|-------------|------|
| `missing vosk model at model\...` | 按 [§2 准备语音模型](#2-准备语音模型必做否则构建会失败) 下载并解压模型 |
| `gcc: command not found` / `cgo` 相关错误 | 安装 MinGW，确认 `gcc --version` 可用后重跑构建脚本 |
| `go: command not found` | 安装 Go 并重启终端 |
| `setup-vosk` 下载失败 | 手动下载 [vosk-win64-0.3.45.zip](https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-win64-0.3.45.zip) 放到 `third_party/`，再执行 `.\scripts\setup-vosk.ps1` 后重新 `build-standalone` |
| 构建成功但运行闪退 | 删除 exe 旁 `.speaknow-data` 后重跑；仍不行则改用 [方式三](#方式三源码开发运行) 看控制台详细日志 |

---

### 方式三：源码开发运行

适合：修改 Go 代码、调试配置或 API；**不必**每次打 120MB 的 exe。  
若 [方式二](#方式二拉取源码并构建-exewindows) 构建出的 exe 仍异常，也应用本方式启动，**错误信息会直接打在终端**，便于排查。

#### 1. 克隆与配置

```powershell
git clone https://github.com/yyyCode/SpeakNow.git
cd SpeakNow

copy configs\config.example.yaml configs\config.yaml
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

**仅本地 Vosk** 可将 `asr.primary` 设为 `vosk`，并关闭讯飞（参见 `configs/config.release.yaml`）。

> `configs/config.yaml` 已 `.gitignore`，勿提交密钥。

#### 2. 准备 Vosk 原生库

检查是否存在：`third_party\vosk\windows-amd64\bin\libvosk.dll`  
若没有，在项目根执行：

```powershell
.\scripts\setup-vosk.ps1
```

#### 3. 准备语音模型

确保存在：`model\vosk-model-small-cn-0.22\`（与 [方式二 §2](#2-准备语音模型必做否则构建会失败) 相同）。

开发模式下模型**不会**打进二进制，必须保留在 `model/` 目录。

#### 4. 启动服务

```powershell
$env:CGO_ENABLED = "1"
go run ./cmd/server -config configs/config.yaml
```

- 未传 `-config` 时使用内置默认（多用于已 `prepare-bundle` 的嵌入构建）。
- 浏览器访问：**http://localhost:8080/web/**

#### 5. 与方式二的区别

| 对比项 | 方式二（exe） | 方式三（go run） |
|--------|----------------|------------------|
| 启动命令 | `.\speaknow.exe` | `go run ./cmd/server -config ...` |
| 是否需要每次构建 | 改代码后需重新 `build-standalone` | 保存代码后重新 `go run` 即可 |
| 网页 / 模型位置 | 内嵌在 exe，解压到 `.speaknow-data` | 直接读 `web/`、`model/` |
| 配置文件 | 默认内嵌；可用 `-config` | 推荐 `configs/config.yaml` |

---

### 使用 Web 演示页

1. **识别引擎**：选择「自动 / 讯飞云端 / Vosk 本地」
2. **识别语言**：Vosk 仅支持中文普通话（`zh-CN`）
3. 点击 **开始录音** 或 **`Ctrl + Space`**
4. 停止后获得最终结果；可开启自动复制、查看历史记录
5. 支持 **上传 WAV/PCM** 离线识别

---

### 常见问题

| 问题 | 建议 |
|------|------|
| **下载的 exe 不能用** | 优先 [方式二](#方式二拉取源码并构建-exewindows) 在本机构建；不要反复换不明来源的 exe |
| **8080 端口被占用** | `netstat -ano \| findstr :8080` 查 PID 后 `taskkill /PID <pid> /F`，或改 `configs/config.yaml` 里 `server.port`（方式三）；exe 需自备 yaml 并用 `-config` |
| **页面能开，录音无识别** | Web 页引擎选「Vosk 本地」；语言选 `zh-CN`；检查麦克风权限；方式三看终端是否报 `vosk` / `model` 错误 |
| **想用讯飞云端** | 方式三 + `configs/config.yaml` 填讯飞密钥；或 exe 使用 `-config` 指向含讯飞的 yaml |
| **`.speaknow-data` 损坏** | 结束所有 `speaknow.exe` 进程后删除该目录，再重新运行 exe |
| **构建报 CGO / gcc** | 安装 MinGW-w64，新开 PowerShell 再执行 `.\scripts\build-standalone.ps1` |
| **只想最快跑起来（已 clone）** | 跳过方式一，直接 [方式二](#方式二拉取源码并构建-exewindows) 全流程 |

---

## 单文件 exe 说明（Windows 分发）

| 项目 | 说明 |
|------|------|
| **推荐构建** | 已 clone 仓库 → [方式二](#方式二拉取源码并构建-exewindows) `.\scripts\build-standalone.ps1` |
| **运行** | `.\speaknow.exe` → http://127.0.0.1:8080/web/ |
| **架构** | 外层 `speaknow.exe`（纯 Go launcher）内嵌 `speaknow-core.exe` + Vosk DLL；首次运行解压到 `.speaknow-data/runtime/` |
| **体积** | 约 120MB+（含模型）；勿提交 Git |
| **发布** | 维护者上传 GitHub Releases，见 [docs/RELEASE.md](docs/RELEASE.md) |
| **exe 异常** | 终端用户下载的 Release 若不可用，应在本机按方式二重打，而非仅重新下载 |

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
