# 发布单文件 exe（GitHub Releases）

`speaknow.exe` 约 120MB+，**不要** `git commit` 进仓库（体积大且会拖慢克隆）。

改为在 **GitHub Release** 里上传安装包，用户下载即用。

## 最终用户：下载后怎么用

1. 从仓库 **Releases** 下载 `speaknow.exe`（仅需这一个文件）。
2. 放到任意文件夹，**双击**运行；或在 PowerShell 中：

```powershell
cd 到 exe 所在目录
.\speaknow.exe
```

3. 浏览器打开：**http://127.0.0.1:8080/web/**
4. 首次运行会在 exe 旁生成 **`.speaknow-data`**（解压内嵌模型与 Vosk 运行时），请保留该目录。
5. 停止：关闭窗口，或 `taskkill /IM speaknow.exe /F`。

可选自定义配置：`.\speaknow.exe -config D:\path\to\config.yaml`

**若下载的 exe 无法运行**（闪退、8080 连不上、识别失败等），不要反复换链接，请 `git clone` 后按 README **[方式二](../README.md#方式二拉取源码并构建-exewindows)** 在本机执行 `.\scripts\build-standalone.ps1` 重新打包。

更完整的三种用法见 [README.md#快速开始](../README.md#快速开始)。

---

## 维护者：本地构建

```powershell
git clone https://github.com/yyyCode/SpeakNow.git
cd SpeakNow
# 确保 backend/model/vosk-model-small-cn-0.22/ 存在
.\backend\scripts\build-standalone.ps1
```

生成项目根目录 `speaknow.exe`。本地验证：`.\speaknow.exe` → 打开 http://127.0.0.1:8080/web/

## 维护者：提交代码（不含 exe）

```powershell
git add .
git commit -m "your message"
git push
```

确保 `git status` 里**没有** `speaknow.exe`。

## 维护者：创建 Release（网页）

1. 打开仓库 → **Releases** → **Draft a new release**
2. Tag：`v1.0.0`（自定）
3. Title：例如 `SpeakNow v1.0.0 Windows 单文件版`
4. 上传附件：`speaknow.exe`
5. 说明：解压/下载后双击运行，浏览器打开 `http://127.0.0.1:8080/web/`

## 维护者：创建 Release（命令行，需安装 [gh](https://cli.github.com/)）

```powershell
gh release create v1.0.0 speaknow.exe `
  --title "SpeakNow v1.0.0" `
  --notes "Windows 单文件版。下载 speaknow.exe 后直接运行，无需另装模型或 Vosk。"
```

## 换一台电脑怎么用

| 方式 | 操作 |
|------|------|
| **只用程序** | Releases 下载 `speaknow.exe` → 双击或 `.\speaknow.exe` → http://127.0.0.1:8080/web/ |
| **拉源码再构建** | `git clone` → 准备 `backend/model/vosk-model-small-cn-0.22/` → `.\backend\scripts\build-standalone.ps1` → `.\speaknow.exe` |
| **改代码调试** | `git clone` → 配置 `backend/configs/config.yaml` → `.\backend\scripts\setup-vosk.ps1` → `go run -C backend ./cmd/server -config configs/config.yaml` |

## 若坚持要把大文件放进 Git

可使用 [Git LFS](https://git-lfs.github.com/) 跟踪 `speaknow.exe`，但 Releases 更简单，也适合最终用户下载。
