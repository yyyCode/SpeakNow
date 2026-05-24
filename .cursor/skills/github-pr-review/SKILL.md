---
name: github-pr-review
description: >-
  Discover, list, and review GitHub pull requests with gh CLI. Auto-fetches open
  PRs and review requests when no PR number is given. Generates HTML analysis cards
  with highlights, weaknesses, suggestions, conflict/CI/merge status. Uses pending
  reviews, batched comments, and code suggestions. Use when reviewing PRs, listing
  PRs, analyzing PR code, generating PR report cards, checking what needs review,
  posting review comments, or adding ```suggestion blocks.
---

# GitHub PR Review

Professional PR reviews via `gh api`: discovery, **HTML 分析卡片**, pending reviews, batched comments, code suggestions, and explicit user approval before posting.

## Required workflow

Do not skip or reorder these steps:

1. **Verify gh CLI** — use the discovery script below (includes auth check); do not run separate `gh` commands
2. **Resolve target PR(s)** — one script call for discovery when user did not give a PR number or URL
3. **Gather PR context** — one script call after PR is selected (see below)
4. **Analyze & draft** — read `diff` + `autoFindings`; write analysis JSON; optionally draft GitHub review comments
5. **Render HTML card** — one script call to generate the analysis report (when user wants visual summary)
6. **Show and get approval** — use AskQuestion before posting anything to GitHub
7. **Post via pending review** — create PENDING review, then submit event (only when posting to GitHub)

**Shell approval rule:** Each script step = **one** Shell invocation. Never split `gh auth status`, `gh pr status`, `gh pr list` into separate Shell calls.

## Script location (important)

```
.cursor/skills/github-pr-review/
  SKILL.md
  templates/
    pr-report.html          # HTML 分析卡片模板
    analysis.example.json   # Agent 写入的分析 JSON 示例
  scripts/
    discover-prs.ps1
    fetch-pr-context.ps1
    render-pr-report.ps1
    discover-prs.sh
    _gh-helpers.ps1
```

Reports output to: `reports/pr/pr-<number>-<timestamp>.html`

Do **not** use nested submodule path `cursor-git-pr-skill/.cursor/skills/...`.

---

## Workflow A — 列出 PR

Run **exactly one** Shell from repo root:

```powershell
. ".cursor/skills/github-pr-review/scripts/discover-prs.ps1"
```

Present tables grouped by: **待你审查** → **Open PR** → **你的 PR**. Use AskQuestion to pick a PR.

---

## Workflow B — PR 代码分析 + HTML 卡片（推荐）

When user asks to analyze a PR, review code, or generate a report card:

### Step 1: Fetch context (one Shell)

```powershell
. ".cursor/skills/github-pr-review/scripts/fetch-pr-context.ps1" -PrNumber <PR_NUMBER>
```

Parse JSON fields:
| Field | Purpose |
|-------|---------|
| `overview` | title, author, branches, files, mergeable, CI |
| `diff` | full diff for code analysis |
| `autoFindings` | auto-detected blockers/warnings (conflicts, CI, binary files, draft) |
| `reviews`, `comments` | existing feedback |
| `commitSha` | for posting GitHub review |

**Agent analysis duties** (no extra Shell): read diff and produce structured judgment:

- **summary** — 1–3 sentences on purpose & impact
- **highlights** — what is good (cleanup, feature, tests, docs)
- **weaknesses** — bugs, missing tests, binary bloat, breaking changes
- **suggestions** — concrete next steps
- **mergeRecommendation** — `{ canMerge, verdict: READY|CAUTION|BLOCKED, summary }`

Combine `autoFindings.blockers` / `warnings` with your code review. Do not ignore auto-detected conflicts or CI failures.

### Step 2: Write analysis JSON (Write tool, not Shell)

Save to `reports/pr/pr-<PR_NUMBER>-analysis.json`. Schema:

```json
{
  "summary": "总体摘要",
  "highlights": ["亮点..."],
  "weaknesses": ["缺点..."],
  "suggestions": ["建议..."],
  "mergeNote": "可选补充",
  "mergeRecommendation": {
    "canMerge": true,
    "verdict": "READY",
    "summary": "合并结论说明"
  }
}
```

See `templates/analysis.example.json` for a full example.

### Step 3: Render HTML card (one Shell)

```powershell
. ".cursor/skills/github-pr-review/scripts/render-pr-report.ps1" -PrNumber <PR_NUMBER> -AnalysisPath "reports/pr/pr-<PR_NUMBER>-analysis.json" -Open
```

Returns JSON with `outputPath`, `verdict`, `canMerge`, `blockers`. Tell user to open the HTML file in browser.

**HTML card sections:** 总体摘要 · 合并评估(可否 merge/冲突/CI) · 亮点 · 缺点 · 建议 · CI 状态 · 变更文件

---

## Workflow C — Post GitHub review (optional)

Only when user explicitly wants to post review comments to GitHub.

Show draft comments + event type (`APPROVE` / `REQUEST_CHANGES` / `COMMENT`). Use AskQuestion:

```
Question: "Ready to post this review on PR #123?"
Options: Yes, post it | No, let me revise
```

Pending review pattern (two API calls):

```bash
gh api repos/:owner/:repo/pulls/<PR_NUMBER>/reviews \
  -X POST -f commit_id="<COMMIT_SHA>" \
  -f 'comments[][path]=path/to/file' -F 'comments[][line]=10' \
  -f 'comments[][side]=RIGHT' -f 'comments[][body]=Comment' --jq '{id, state}'

gh api repos/:owner/:repo/pulls/<PR_NUMBER>/reviews/<REVIEW_ID>/events \
  -X POST -f event="REQUEST_CHANGES" -f body="Overall message"
```

---

## Auto-detection (in `autoFindings`)

Scripts automatically flag:
- Merge conflicts (`mergeable: CONFLICTING`)
- Draft PR, branch behind base, blocked merge state
- CI failures / pending checks
- Binary/large files (.dll, .zip, .pcm, etc.)
- Large single-file diffs (>500 lines)

Agent must add semantic analysis from reading `diff`.

---

## Red flags

- Splitting gh discovery into parallel Shell calls — use `discover-prs.ps1`
- Skipping analysis JSON and only showing chat summary when user asked for report card
- Posting GitHub review without AskQuestion approval
- Using nested `cursor-git-pr-skill/...` path
- Using bare `powershell -File` on Windows — dot-source instead

## Additional resources

- `templates/analysis.example.json` — analysis JSON schema
- `templates/pr-report.html` — HTML template
- `reference.md` — GitHub review API syntax
