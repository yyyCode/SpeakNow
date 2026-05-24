---
name: git-commit
description: >-
  Analyzes SpeakNow diffs, classifies Conventional Commits (feat/fix/docs/…),
  enforces one domain per commit, writes Chinese summaries, and commits safely.
  Use when the user asks to commit, write a commit message, or runs /git-commit.
disable-model-invocation: true
---

# SpeakNow Git Commit

仅在用户**明确要求**创建 commit 时执行。若意图不清，先询问再提交。

## 核心原则

1. **一次提交 = 同一领域**：一次 commit 只包含同一 `scope`（模块/子系统）内的改动。功能、修复、文档、脚本等不要混在同一 commit。
2. **先分类再写 message**：根据 diff **推断** Conventional Commits 的 `type`，再写标题与正文摘要。
3. **小范围**：只 `git add` 与本次任务相关的路径；遵循仓库「尽量小改动」约定。

## 提交前检查

- **禁止入库**：`speaknow.exe`、大型构建产物（见 `docs/RELEASE.md`）、`.env`、密钥、token
- **无变更则停止**：无有效变更时不创建空 commit
- **混合领域**：若工作区同时有多个 scope 的改动，**默认拆成多次 commit**（按 scope 分批 `git add` + `commit`）；若用户明确要求「一次提交全部」，需在回复中说明混入了哪些领域并征得同意

## 分析变更（并行执行）

```bash
git status
git diff
git diff --staged
git log --oneline -10
```

对 **staged**（或即将暂存）的文件逐条判断：

| 步骤 | 动作 |
|------|------|
| 1 | 列出变更文件路径，归到下方 **scope** |
| 2 | 若出现 **2 个及以上 scope** → 拆分提交或询问用户 |
| 3 | 根据 diff 性质选定 **type**（见下表） |
| 4 | 写 `type(scope): 简短标题`，正文用中文概括「改了什么、为何改」 |

用户未 `git add` 时：先说明拟暂存文件与归属 scope，再按 scope 分批 `git add`（不要 `git add .`，除非用户明确要求）。

## Type 判定（Conventional Commits）

| type | 何时使用 |
|------|----------|
| `feat` | 新功能、新 API、新配置项、用户可感知的行为增强 |
| `fix` | 修复 bug、异常、错误处理、回归问题 |
| `docs` | 仅文档、README、注释（无逻辑变更） |
| `refactor` | 结构调整、重命名、抽函数；**不改变**对外行为 |
| `perf` | 性能优化（延迟、内存、并发） |
| `test` | 测试新增或修改 |
| `build` | 构建脚本、`go.mod`、打包、嵌入资源 |
| `chore` | 杂项：`.gitignore`、`.cursor`、依赖目录整理、无用户可见影响 |
| `ci` | CI/CD、Docker、deploy 编排 |

拿不准时：修 bug → `fix`；新能力 → `feat`；只改说明 → `docs`。

## Scope 参考（SpeakNow）

按**变更文件主目录**选最贴切的一个 scope：

| scope | 典型路径 |
|-------|----------|
| `asr` | `internal/service/asr/` |
| `vosk` | `internal/provider/vosk/`, `internal/voskruntime/`, `third_party/vosk/`, `model/` |
| `xunfei` | `internal/provider/xunfei/` |
| `provider` | `internal/provider/`（factory、mock、aliyun、tencent 等，非 vosk/xunfei 专目录时） |
| `handler` | `internal/handler/`, `internal/middleware/` |
| `config` | `internal/config/`, `configs/`, `cmd/server/config_load.go` |
| `web` | `web/`, `internal/assets/web/` |
| `cmd` | `cmd/server/`（main、启动逻辑） |
| `scripts` | `scripts/` |
| `deploy` | `deployments/` |
| `deps` | `go.mod`, `go.sum`（若伴随功能改动，跟功能 scope 同 commit） |
| `cursor` | `.cursor/` |

多目录但同一功能（如 feat 同时改 `handler` + `asr`）→ 选**最核心的** scope，或在标题中体现主模块（仍只一个 scope）。

## 提交信息格式

```
<type>(<scope>): <中文简短标题，≤50 字>

<正文：2–5 行中文，总结本次改动要点，可用列表>
```

**标题**：动词开头，说清结果（如「修复 WebSocket 断线重连」「支持 Vosk 离线配置」）。

**正文**（建议 always 写）：概括改了哪些文件/行为，避免逐文件罗列，突出「做了什么、为什么」。

**示例**

```
fix(handler): 修复 WebSocket 断线后无法重连

- 在连接关闭时清理 session 状态
- 重连时复用同一 clientId，避免重复计费
```

```
feat(vosk): 支持从配置指定本地模型路径

- 读取 config 中 vosk.model_path
- 启动时校验目录存在，否则返回明确错误
```

```
docs(readme): 补充 Windows 单文件版发布说明

- 说明 speaknow.exe 通过 GitHub Release 分发，勿提交进仓库
```

```
chore(cursor): 添加 git-commit Agent Skill

- 约定 Conventional Commits 与按 scope 拆分提交
```

## 执行提交（按 scope 分批）

每个 scope 一批：

```bash
git add <本 scope 相关文件>
git commit -m "$(cat <<'EOF'
<type>(<scope>): <标题>

<正文摘要>

EOF
)"
git status
```

全部 scope 提交完成后，再向用户汇总各 commit 的 hash 与 message。

## 安全协议

- **不要**改 `git config`；**不要** `--no-verify`、`--force`、hard reset（除非用户明确要求）
- **不要**对 main/master force push
- **不要** amend，除非：用户明确要求，且 HEAD 为本会话创建、且未 push
- hook 失败：修复后**新建** commit
- **不要** push，除非用户明确要求

## 回复用户

说明：

1. 推断的 **type / scope** 及理由（一句话）
2. 若拆分：每个 commit 的文件列表与完整 message
3. 若拒绝混提：列出了哪些 scope 建议分开提交
4. 最终 `git status` 与分支状态
