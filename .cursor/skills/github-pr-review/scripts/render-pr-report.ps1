# Render PR analysis card (HTML) from agent-written analysis JSON + live gh metadata.
param(
    [Parameter(Mandatory = $true)]
    [int]$PrNumber,
    [Parameter(Mandatory = $true)]
    [string]$AnalysisPath,
    [string]$Repo = "",
    [string]$OutputPath = "",
    [switch]$Open
)

$ErrorActionPreference = "Stop"
. "$PSScriptRoot\_gh-helpers.ps1"

$skillRoot = Split-Path $PSScriptRoot -Parent
$templatePath = Join-Path $skillRoot "templates\pr-report.html"

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-JsonResult @{ ok = $false; error = "gh_not_installed"; message = "Install GitHub CLI" }
    exit 1
}

if (-not (Test-Path $AnalysisPath)) {
    Write-JsonResult @{ ok = $false; error = "analysis_not_found"; path = $AnalysisPath }
    exit 1
}

if (-not (Test-Path $templatePath)) {
    Write-JsonResult @{ ok = $false; error = "template_not_found"; path = $templatePath }
    exit 1
}

$analysisRaw = [System.IO.File]::ReadAllText((Resolve-Path $AnalysisPath), [System.Text.Encoding]::UTF8)
$analysis = $analysisRaw | ConvertFrom-Json

$repoArgs = @()
if ($Repo) { $repoArgs = @("--repo", $Repo) }

$viewFields = @(
    "title", "body", "author", "baseRefName", "headRefName", "commits",
    "reviewDecision", "statusCheckRollup", "url", "files", "isDraft",
    "mergeable", "mergeStateStatus", "additions", "deletions", "changedFiles", "state"
) -join ","

$viewArgs = @("pr", "view", "$PrNumber") + $repoArgs + @("--json", $viewFields)
$overviewRaw = Invoke-GhText @viewArgs
if (-not $overviewRaw) {
    Write-JsonResult @{ ok = $false; error = "pr_not_found"; prNumber = $PrNumber }
    exit 1
}

$overview = $overviewRaw | ConvertFrom-Json
$auto = Get-PrAutoFindings -Overview $overview

$highlights = @()
$weaknesses = @()
$suggestions = @()
if ($analysis.highlights) { $highlights = @(ConvertTo-GhArray $analysis.highlights) }
if ($analysis.weaknesses) { $weaknesses = @(ConvertTo-GhArray $analysis.weaknesses) }
if ($analysis.suggestions) { $suggestions = @(ConvertTo-GhArray $analysis.suggestions) }

$summary = if ($analysis.summary) { [string]$analysis.summary } else { "No summary provided." }
$mergeNote = if ($analysis.mergeNote) { [string]$analysis.mergeNote } else { "" }

$verdict = $auto.verdict
$canMerge = $auto.canMerge
if ($analysis.mergeRecommendation) {
    $rec = $analysis.mergeRecommendation
    if ($null -ne $rec.canMerge) { $canMerge = [bool]$rec.canMerge }
    if ($rec.verdict) { $verdict = [string]$rec.verdict }
    if ($rec.summary -and -not $mergeNote) { $mergeNote = [string]$rec.summary }
}

$verdictText = if ($mergeNote) {
    $mergeNote
} else {
    switch ($verdict) {
        "READY" { if ($canMerge) { "Ready to merge" } else { "Merge with caution" } }
        "CAUTION" { "Caution: review warnings before merge" }
        "BLOCKED" { "Blocked: resolve blockers before merge" }
        default { "Manual review required" }
    }
}

$verdictClass = $verdict.ToLowerInvariant()
if ($verdictClass -notin @("ready", "caution", "blocked")) {
    $verdictClass = if ($canMerge) { "ready" } else { "blocked" }
}

$verdictBadge = switch ($verdictClass) {
    "ready" { '<span class="badge ready">Merge OK</span>' }
    "caution" { '<span class="badge caution">Caution</span>' }
    default { '<span class="badge blocked">Blocked</span>' }
}

$mergeableBadge = switch ($overview.mergeable) {
    "MERGEABLE" { '<span class="badge ready">No conflicts</span>' }
    "CONFLICTING" { '<span class="badge blocked">Conflicts</span>' }
    default { '<span class="badge">Mergeable: ' + (Escape-Html $overview.mergeable) + '</span>' }
}

$conflictBadge = if ($overview.mergeStateStatus) {
    '<span class="badge">' + (Escape-Html $overview.mergeStateStatus) + '</span>'
} else { "" }

$commits = ConvertTo-GhArray $overview.commits
$commitCount = $commits.Count

$files = ConvertTo-GhArray $overview.files
$filesList = ""
foreach ($f in $files) {
    $change = "+$($f.additions)/-$($f.deletions)"
    $filesList += "<li><code>$(Escape-Html $f.path)</code> <span style=""color:var(--muted)"">($change)</span></li>"
}
if (-not $filesList) { $filesList = '<li class="empty">No file info</li>' }

$ciTable = ""
$checks = ConvertTo-GhArray $overview.statusCheckRollup
if ($checks.Count -eq 0) {
    $ciTable = '<p style="color:var(--muted);font-size:0.9rem;">No CI checks returned</p>'
}
else {
    $ciTable = "<table class=""ci-table""><thead><tr><th>Check</th><th>State</th></tr></thead><tbody>"
    foreach ($check in $checks) {
        $name = if ($check.context) { $check.context } elseif ($check.name) { $check.name } else { "check" }
        $state = if ($check.state) { $check.state } else { $check.conclusion }
        $ciTable += "<tr><td>$(Escape-Html $name)</td><td>$(Escape-Html $state)</td></tr>"
    }
    $ciTable += "</tbody></table>"
}

$blockers = @($auto.blockers | ForEach-Object { [string]$_ })
$warnings = @($auto.warnings | ForEach-Object { [string]$_ })

$html = [System.IO.File]::ReadAllText($templatePath, [System.Text.Encoding]::UTF8)
$replacements = @{
    "{{PAGE_TITLE}}"       = Escape-Html ("PR #" + $PrNumber + " Report")
    "{{PR_TITLE}}"         = Escape-Html $overview.title
    "{{PR_NUMBER}}"        = [string]$PrNumber
    "{{AUTHOR}}"           = Escape-Html $overview.author.login
    "{{HEAD_BRANCH}}"      = Escape-Html $overview.headRefName
    "{{BASE_BRANCH}}"      = Escape-Html $overview.baseRefName
    "{{PR_URL}}"           = Escape-Html $overview.url
    "{{VERDICT_BADGE}}"    = $verdictBadge
    "{{MERGEABLE_BADGE}}"  = $mergeableBadge
    "{{CONFLICT_BADGE}}"   = $conflictBadge
    "{{ADDITIONS}}"        = [string]$overview.additions
    "{{DELETIONS}}"        = [string]$overview.deletions
    "{{CHANGED_FILES}}"    = [string]$overview.changedFiles
    "{{COMMITS}}"          = [string]$commitCount
    "{{SUMMARY}}"          = Escape-Html $summary
    "{{VERDICT_CLASS}}"    = $verdictClass
    "{{MERGE_VERDICT_TEXT}}" = Escape-Html $verdictText
    "{{MERGE_NOTE}}"       = ""
    "{{BLOCKERS_LIST}}"    = ConvertTo-HtmlList -Items $blockers -EmptyText "None"
    "{{WARNINGS_LIST}}"    = ConvertTo-HtmlList -Items $warnings -EmptyText "None"
    "{{HIGHLIGHTS_LIST}}"  = ConvertTo-HtmlList -Items $highlights -EmptyText "None"
    "{{WEAKNESSES_LIST}}"  = ConvertTo-HtmlList -Items $weaknesses -EmptyText "None"
    "{{SUGGESTIONS_LIST}}" = ConvertTo-HtmlList -Items $suggestions -EmptyText "None"
    "{{CI_TABLE}}"         = $ciTable
    "{{FILES_LIST}}"       = $filesList
    "{{GENERATED_AT}}"     = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
}

foreach ($key in $replacements.Keys) {
    $html = $html.Replace($key, $replacements[$key])
}

$root = git rev-parse --show-toplevel 2>$null
if ($root) {
    $reportDir = Join-Path $root "reports\pr"
}
else {
    $reportDir = Join-Path (Get-Location) "reports\pr"
}

if (-not (Test-Path $reportDir)) {
    New-Item -ItemType Directory -Path $reportDir -Force | Out-Null
}

if (-not $OutputPath) {
    $stamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $OutputPath = Join-Path $reportDir ("pr-" + $PrNumber + "-" + $stamp + ".html")
}

$OutputPath = $ExecutionContext.SessionState.Path.GetUnresolvedProviderPathFromPSPath($OutputPath)
[System.IO.File]::WriteAllText($OutputPath, $html, [System.Text.UTF8Encoding]::new($false))

if ($Open) {
    Start-Process $OutputPath | Out-Null
}

Write-JsonResult @{
    ok           = $true
    prNumber     = $PrNumber
    outputPath   = $OutputPath
    verdict      = $verdict
    canMerge     = $canMerge
    mergeable    = $overview.mergeable
    mergeState   = $overview.mergeStateStatus
    blockers     = $blockers
    warnings     = $warnings
    analysisPath = (Resolve-Path $AnalysisPath).Path
}
