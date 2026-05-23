# SpeakNow 语音输入法

SpeakNow 是一个基于 Go 的**语音输入法后端服务**，配套 Web 演示页，支持**实时流式语音识别**与离线文件识别。项目采用多厂商 ASR 抽象、智能路由降级、Redis 缓存与限流成本控制，在准确度、易用性、响应速度和费用之间取得平衡。

---

## 项目亮点

| 亮点 | 说明 |
|------|------|
| **实时流式识别** | 浏览器麦克风 → Web Audio 转 PCM → WebSocket 推流 → 讯飞 ASR，**边说边出字** |
| **多厂商抽象** | 统一 `Provider` 接口，已接入讯飞，预留阿里云 / 腾讯云，Mock 模式免密钥开发 |
| **智能路由降级** | 主 Provider 不可用时自动 fallback，保障可用性 |
| **双层限流** | 本地 Token Bucket + Redis 分布式计数，防止 QPS 爆炸和费用失控 |
| **结果缓存** | 相同音频指纹命中 Redis 缓存，零成本快速返回 |
| **成本统计** | 每次调用记录时长、延迟、缓存命中与费用 |
| **开箱即用 UI** | 参考 it365 风格的 Web 演示页，支持 `Ctrl+Space` 快捷键、自动复制、历史记录 |
| **容器化部署** | 提供 Dockerfile 与 docker-compose，一键启动 |

---

## 功能概览

```
浏览器 / 客户端
    │
    ├─ WebSocket  /api/v1/asr/stream   →  实时流式识别（主流程）
    │
    └─ REST POST  /api/v1/asr/recognize →  上传 WAV/PCM 离线识别
                │
                ▼
         [Gin 中间件]  限流 · 日志 · CORS
                │
                ▼
         [ASR 服务层]  缓存 · 路由 · 成本统计
                │
                ▼
    ┌───────────┼───────────┐
    ▼           ▼           ▼
  讯飞        阿里云       Mock
 (已接入)    (预留)      (开发调试)
```

---

## 快速开始

### 环境要求

- **Go 1.23+**
- **Redis**（可选，未连接时自动降级为无缓存模式）
- 麦克风权限（Web 实时识别）

### 1. 克隆与配置

```bash
git clone https://github.com/yyyCode/SpeakNow.git
cd SpeakNow

# 复制配置模板并填入密钥
copy configs\config.example.yaml configs\config.yaml   # Windows
# cp configs/config.example.yaml configs/config.yaml   # Linux/macOS
```

编辑 `configs/config.yaml`，填入讯飞密钥（或其他厂商配置）：

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
```

> `configs/config.yaml` 已加入 `.gitignore`，请勿将密钥提交到 Git。

也可通过环境变量注入（前缀 `SPEAKNOW_`）：

```bash
set SPEAKNOW_PROVIDERS_XUNFEI_ENABLED=true
set SPEAKNOW_PROVIDERS_XUNFEI_APP_ID=xxx
set SPEAKNOW_PROVIDERS_XUNFEI_API_KEY=xxx
set SPEAKNOW_PROVIDERS_XUNFEI_API_SECRET=xxx
```

### 2. 启动服务

```bash
go mod tidy
go run ./cmd/server -config configs/config.yaml
```

看到 `Listening and serving HTTP on 0.0.0.0:8080` 即启动成功。

### 3. 打开前端

**前端无需单独启动**，由 Go 后端静态托管。

浏览器访问：

| 地址 | 说明 |
|------|------|
| http://localhost:8080/web/ | Web 演示页（推荐） |
| http://localhost:8080/ | 自动跳转到 `/web/` |

### 4. 使用方式

1. 点击 **「开始录音」** 或按 **`Ctrl + Space`**
2. 对着麦克风说话，文字会**实时**显示在输入框
3. 再次点击按钮停止录音，获得最终结果
4. 开启「录音结束自动复制」可将结果写入剪贴板
5. 也可点击 **「上传 WAV/PCM 文件」** 进行离线识别

---

## 关闭 / 重启服务

### 查找占用 8080 的进程

```powershell
netstat -ano | findstr :8080
```

### 结束进程

```powershell
taskkill /PID <PID> /F
# 或
taskkill /IM speaknow.exe /F
```

---

## API 文档

### 实时流式识别（WebSocket）

```
GET /api/v1/asr/stream?language=zh-CN&enable_punc=true&provider=auto
```

| 方向 | 格式 | 说明 |
|------|------|------|
| 客户端 → 服务端 | 二进制 | PCM 音频分片（16kHz, 16-bit, mono） |
| 客户端 → 服务端 | JSON | `{"action":"end"}` 结束录音 |
| 服务端 → 客户端 | JSON | `{"type":"partial","text":"..."}` 中间结果 |
| 服务端 → 客户端 | JSON | `{"type":"final","text":"..."}` 最终结果 |
| 服务端 → 客户端 | JSON | `{"type":"error","message":"..."}` 错误 |

### 离线文件识别（REST）

```bash
curl -X POST http://localhost:8080/api/v1/asr/recognize \
  -F "file=@test.wav" \
  -F "language=zh-CN" \
  -F "enable_punc=true" \
  -F "provider=auto"
```

### 其他接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查 |
| GET | `/api/v1/providers/status` | 各 ASR 厂商状态 |
| GET | `/api/v1/stats/cost` | 成本统计 |

---

## Docker 部署

```bash
# 可选：启动 Redis
docker run -d --name speaknow-redis -p 6379:6379 redis:7-alpine

# 使用 docker-compose
cd deployments
docker compose up --build
```

---

## 项目结构

```
SpeakNow/
├── cmd/server/              # 服务入口
├── internal/
│   ├── config/              # Viper 配置
│   ├── handler/             # HTTP / WebSocket 处理器
│   ├── middleware/          # 限流、日志、CORS
│   ├── provider/            # ASR 厂商抽象（mock / xunfei / aliyun / tencent）
│   ├── service/             # ASR、缓存、路由、成本统计
│   └── queue/               # 任务队列（预留）
├── pkg/                     # fingerprint、logger、response
├── web/                     # 前端演示页（静态 HTML）
├── configs/
│   ├── config.example.yaml  # 配置模板（可提交 Git）
│   └── config.yaml          # 本地配置（含密钥，已 gitignore）
├── deployments/             # Dockerfile、docker-compose
└── iat_ws_go_demo/          # 讯飞官方 WebSocket Demo 参考
```

---

## 讯飞 ASR 配置

1. 前往 [讯飞开放平台](https://www.xfyun.cn/) 注册并创建应用
2. 开通 **语音听写（流式版）** 服务
3. 获取 APPID、APIKey、APISecret 填入 `configs/config.yaml`

**音频格式要求**：

- 实时识别：前端 Web Audio 自动转为 16kHz PCM
- 文件上传：支持 `.wav` / `.pcm`（16kHz, 16-bit, mono）

参考实现：`iat_ws_go_demo/iat_ws_go_demo.go`

---

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.23 |
| Web 框架 | Gin |
| WebSocket | gorilla/websocket |
| 配置 | Viper |
| 日志 | Zap |
| 缓存 / 限流 | Redis |
| 前端 | 原生 HTML + Web Audio API |

---

## 开发路线

- [x] Phase 1：项目脚手架、Mock ASR、REST 识别、缓存、限流、Web 演示
- [x] Phase 2：讯飞 WebSocket 流式识别、实时前端、智能路由
- [ ] Phase 3：阿里云 / 腾讯云接入、API Key 鉴权、Prometheus 监控、PostgreSQL 持久化

---

## License

MIT
