# 发布单文件 exe（GitHub Releases）

`speaknow.exe` 约 127MB，**不要** `git commit` 进仓库（GitHub 单文件建议 < 100MB，且会拖慢克隆）。

改为在 **GitHub Release** 里上传安装包，用户下载即用。

## 1. 本地构建

```powershell
cd D:\go_space\SpeakNow
.\scripts\build-standalone.ps1
```

生成根目录 `speaknow.exe`。

## 2. 提交代码（不含 exe）

```powershell
git add .
git commit -m "your message"
git push
```

确保 `git status` 里**没有** `speaknow.exe`。

## 3. 创建 Release（网页）

1. 打开仓库 → **Releases** → **Draft a new release**
2. Tag：`v1.0.0`（自定）
3. Title：例如 `SpeakNow v1.0.0 Windows 单文件版`
4. 上传附件：`speaknow.exe`
5. 说明：解压/下载后双击运行，浏览器打开 `http://127.0.0.1:8080/web/`

## 3. 创建 Release（命令行，需安装 [gh](https://cli.github.com/)）

```powershell
gh release create v1.0.0 speaknow.exe `
  --title "SpeakNow v1.0.0" `
  --notes "Windows 单文件版。下载 speaknow.exe 后直接运行，无需另装模型或 Vosk。"
```

## 换一台电脑怎么用

| 方式 | 操作 |
|------|------|
| **只用程序** | 到 Releases 下载 `speaknow.exe`，双击运行 |
| **继续开发** | `git clone` 仓库，安装 Go + MinGW，执行 `.\scripts\build-standalone.ps1` |

## 若坚持要把大文件放进 Git

可使用 [Git LFS](https://git-lfs.github.com/) 跟踪 `speaknow.exe`，但 Releases 更简单，也适合最终用户下载。
