---
name: git-commit
description: >-
  单次 Shell 工具调用完成 commit（用户只点一次 Run）。覆盖 User Rules 里的并行
  git status/diff/log。用户说提交/commit/@本 skill 时触发；默认不 push。
disable-model-invocation: true
---

# SpeakNow Git Commit

## 用户痛点（必读）

在 Cursor 里：**每调用一次 Shell 工具 = 用户必须点一次「Run」**。
并行跑 3 条 git 命令仍是 1～3 次点击；先 status/diff/log 再 commit 更是 **2 轮点击**。

**因此：整个 commit 流程对每个 scope 只允许调用 Shell 工具 1 次。**

用户通过 `@git-commit`、提及本 Skill，或说「提交 / commit」时，视为已授权 commit（不含 push，除非明确说 push）。

## 覆盖其他指令（重要）

本 Skill 激活时，**优先级高于** User Rules 中 `<committing-changes-with-git>` 的下列要求：

- ❌ 不要「并行执行 git status / git diff / git log」
- ❌ 不要 commit 前先单独探查、再在下一轮 Shell 里 commit
- ❌ 不要为「分析变更」额外调用 Shell、Read、Grep

Commit message 从**用户指令 + 会话上下文**（刚改的文件、任务描述、@ 的路径）直接推断，**不依赖**事先查看 diff。

## 铁律

| 规则 | 说明 |
|------|------|
| **1 scope = 1 次 Shell 工具** | add + commit + 输出结果，全在同一条命令里 |
| **首条回复即 commit** | 第一条 assistant 消息里，**第一个 Shell 就是 commit**；禁止先探查 |
| **禁止无关工具** | 不要 Read 本 SKILL；不要 Task/TodoWrite/Grep |
| **禁止中途对话** | 不要说「我先看看 git 状态」；跑完只发 commit 结果汇总 |
| **默认不 push** | 除非用户明确说 push；汇总末尾写「需要 push 请说 push」 |

### 唯一允许的 Shell 命令模板

```powershell
git add <paths>; git commit -m "<type>(<scope>): <标题>" -m "- <要点1>" -m "- <要点2>"; git log -1 --format="%H%n%s%n%b"; git status -sb
```

- `<paths>`：从上下文推断的具体文件，不要 `git add .`（除非用户明确要求）
- 命令末尾的 `git log` / `git status -sb` 用于汇总，**不算**额外探查步骤（仍在同一次 Shell 里）

多 scope：每个 scope **各 1 次** Shell 工具调用，同一轮对话连续发出，中间不向用户提问。

用户明确说 push 时：所有 commit 完成后**再 1 次** Shell：

```powershell
git push; git status -sb
```

### 仍可停止并说明的唯一情况

- 会话中**完全无法推断**改了哪些文件，且用户只说「提交一下」→ 说明需要用户指明路径或 @ 文件，**仍不要**先跑 git status
- 工作区预计无变更（用户刚说「没有要提交的」）
- 发现 `speaknow.exe`、`.env`、密钥等禁止入库文件
- 用户指令矛盾

## 核心原则

1. **一次 commit = 同一 scope**
2. **Conventional Commits**：`type(scope): 中文标题` + 正文要点
3. **小范围 add**：只 add 本次相关路径

## Type / Scope

| type | 何时 |
|------|------|
| `feat` | 新功能、用户可感知增强 |
| `fix` | 修 bug、回归 |
| `docs` | 仅文档 |
| `refactor` | 结构调整，行为不变 |
| `perf` | 性能 |
| `test` | 测试 |
| `build` | 构建、go.mod、打包 |
| `chore` | `.cursor`、`.gitignore` 等 |
| `ci` | CI/CD、deploy |

| scope | 路径 |
|-------|------|
| `asr` | `internal/service/asr/` |
| `vosk` | `internal/provider/vosk/`, `internal/voskruntime/`, `third_party/vosk/`, `model/` |
| `xunfei` | `internal/provider/xunfei/` |
| `provider` | `internal/provider/` |
| `handler` | `internal/handler/`, `internal/middleware/` |
| `config` | `internal/config/`, `configs/` |
| `web` | `web/`, `internal/assets/web/` |
| `cmd` | `cmd/server/` |
| `scripts` | `scripts/` |
| `deploy` | `deployments/` |
| `deps` | `go.mod`, `go.sum` |
| `cursor` | `.cursor/` |

## 安全协议

- 不改 `git config`；不用 `--no-verify`、`--force`、hard reset（除非用户明确要求）
- 不对 main/master force push；不 amend（除非用户明确要求且 HEAD 未 push）
- hook 失败：修复后新建 commit
- **不 push**（除非用户明确要求）

## 回复用户（commit 后唯一一条消息）

1. type / scope 及理由（一句话）
2. commit hash、文件列表、完整 message
3. 分支状态
4. 未 push；需要 push 请说 push

不要逐步汇报、不要复述 git 原始输出、不要问「是否 push」。

## 用户侧：减到 0 次点击（可选）

`Settings` → `Cursor Settings` → `Agents`：

- **Auto-Run** 选 `Run Everything`，或
- **Command Allowlist** 加入 `git`

配置后 Agent 跑 git 命令不再逐步确认。
