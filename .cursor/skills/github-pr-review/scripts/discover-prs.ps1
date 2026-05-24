# Discover open PRs and review requests in one shot (single Shell approval).
param(
    [string]$Repo = "",
    [switch]$AllRepos
)

$ErrorActionPreference = "Continue"
. "$PSScriptRoot\_gh-helpers.ps1"

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-JsonResult @{
        ok      = $false
        error   = "gh_not_installed"
        message = "Install GitHub CLI: https://cli.github.com/ (Windows: winget install GitHub.cli)"
    }
    exit 1
}

$authRaw = Invoke-GhText auth status
if (-not $authRaw) {
    Write-JsonResult @{
        ok      = $false
        error   = "not_authenticated"
        message = "Run: gh auth login"
    }
    exit 1
}

$account = ""
if ($authRaw -match "Logged in to github\.com account (\S+)") {
    $account = $Matches[1]
}

$repoName = $Repo
if (-not $repoName) {
    $repoResult = Invoke-GhJson repo view --json nameWithOwner
    $repoItem = Get-GhSingle $repoResult
    if ($repoItem) {
        $repoName = $repoItem.nameWithOwner
    }
}

$needsReview = @()
$openPrs = @()
$prStatus = ""
$crossRepo = @()

if ($repoName) {
    $needsReview = Invoke-GhJson pr list --repo $repoName --search "review-requested:@me" --state open `
        --json number,title,author,isDraft,reviewDecision,url,updatedAt

    $openPrs = Invoke-GhJson pr list --repo $repoName --state open --limit 30 `
        --json number,title,author,isDraft,reviewDecision,url,headRefName,baseRefName,updatedAt

    $prStatus = Invoke-GhText pr status --repo $repoName
    if (-not $prStatus) { $prStatus = "" }
}
else {
    $prStatus = Invoke-GhText pr status
    if (-not $prStatus) { $prStatus = "" }
}

if ($AllRepos) {
    $crossRepo = Invoke-GhJson search prs --review-requested=@me --state=open --limit=15 `
        --json number,title,repository,updatedAt,author
}

Write-JsonResult @{
    ok              = $true
    account         = $account
    repo            = $repoName
    needsYourReview = $needsReview
    openPrs         = $openPrs
    crossRepoReview = $crossRepo
    prStatus        = $prStatus
}
