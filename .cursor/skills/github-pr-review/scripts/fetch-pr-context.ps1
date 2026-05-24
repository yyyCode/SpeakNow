# Fetch PR overview, diff, merge status, and existing comments in one shot.
param(
    [Parameter(Mandatory = $true)]
    [int]$PrNumber,
    [string]$Repo = "",
    [int]$MaxDiffChars = 120000
)

$ErrorActionPreference = "Continue"
. "$PSScriptRoot\_gh-helpers.ps1"

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-JsonResult @{ ok = $false; error = "gh_not_installed"; message = "Install GitHub CLI: https://cli.github.com/" }
    exit 1
}

$repoArgs = @()
if ($Repo) { $repoArgs = @("--repo", $Repo) }

$viewFields = @(
    "title", "body", "author", "baseRefName", "headRefName", "commits",
    "reviewDecision", "statusCheckRollup", "url", "files", "isDraft",
    "mergeable", "mergeStateStatus", "additions", "deletions", "changedFiles",
    "createdAt", "updatedAt", "state"
) -join ","

$viewArgs = @("pr", "view", "$PrNumber") + $repoArgs + @("--json", $viewFields)
$overviewRaw = Invoke-GhText @viewArgs
if (-not $overviewRaw) {
    Write-JsonResult @{ ok = $false; error = "pr_not_found"; prNumber = $PrNumber }
    exit 1
}

$overview = $overviewRaw | ConvertFrom-Json
$commits = ConvertTo-GhArray $overview.commits
$commitSha = if ($commits.Count -gt 0) { $commits[-1].oid } else { "" }

$diffArgs = @("pr", "diff", "$PrNumber") + $repoArgs
$diff = Invoke-GhText @diffArgs
if (-not $diff) { $diff = "" }

$diffTruncated = $false
if ($diff.Length -gt $MaxDiffChars) {
    $diff = $diff.Substring(0, $MaxDiffChars) + "`n... [diff truncated at $MaxDiffChars chars]"
    $diffTruncated = $true
}

$feedbackArgs = @("pr", "view", "$PrNumber") + $repoArgs + @("--json", "reviews,comments")
$feedbackRaw = Invoke-GhText @feedbackArgs
$feedback = $null
if ($feedbackRaw) { $feedback = $feedbackRaw | ConvertFrom-Json }

$autoFindings = Get-PrAutoFindings -Overview $overview

Write-JsonResult @{
    ok            = $true
    prNumber      = $PrNumber
    repo          = $Repo
    commitSha     = $commitSha
    overview      = $overview
    diff          = $diff
    diffTruncated = $diffTruncated
    reviews       = $feedback.reviews
    comments      = $feedback.comments
    autoFindings  = $autoFindings
}
