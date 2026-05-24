#!/usr/bin/env bash
# Discover open PRs and review requests in one shot (single Shell approval).
set -euo pipefail

REPO="${1:-}"
ALL_REPOS="${2:-}"

if ! command -v gh >/dev/null 2>&1; then
  jq -n --arg msg "Install GitHub CLI: https://cli.github.com/" \
    '{ok:false,error:"gh_not_installed",message:$msg}'
  exit 1
fi

AUTH_RAW="$(gh auth status 2>&1)" || {
  jq -n --arg detail "$AUTH_RAW" \
    '{ok:false,error:"not_authenticated",message:"Run: gh auth login",detail:$detail}'
  exit 1
}

ACCOUNT="$(echo "$AUTH_RAW" | sed -n 's/.*account \([^ ]*\).*/\1/p' | head -1)"

REPO_NAME="$REPO"
if [ -z "$REPO_NAME" ]; then
  REPO_NAME="$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || true)"
fi

NEEDS='[]'
OPEN='[]'
CROSS='[]'
STATUS=""

if [ -n "$REPO_NAME" ]; then
  NEEDS="$(gh pr list --repo "$REPO_NAME" --search "review-requested:@me" --state open \
    --json number,title,author,isDraft,reviewDecision,url,updatedAt 2>/dev/null || echo '[]')"
  OPEN="$(gh pr list --repo "$REPO_NAME" --state open --limit 30 \
    --json number,title,author,isDraft,reviewDecision,url,headRefName,baseRefName,updatedAt 2>/dev/null || echo '[]')"
  STATUS="$(gh pr status --repo "$REPO_NAME" 2>&1 || true)"
else
  STATUS="$(gh pr status 2>&1 || true)"
fi

if [ "$ALL_REPOS" = "--all-repos" ]; then
  CROSS="$(gh search prs --review-requested=@me --state=open --limit=15 \
    --json number,title,repository,updatedAt,author 2>/dev/null || echo '[]')"
fi

jq -n \
  --arg account "$ACCOUNT" \
  --arg repo "$REPO_NAME" \
  --arg status "$STATUS" \
  --argjson needs "$NEEDS" \
  --argjson open "$OPEN" \
  --argjson cross "$CROSS" \
  '{ok:true,account:$account,repo:$repo,needsYourReview:$needs,openPrs:$open,crossRepoReview:$cross,prStatus:$status}'
