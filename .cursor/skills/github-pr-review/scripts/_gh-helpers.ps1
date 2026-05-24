# Shared helpers for gh CLI JSON output (UTF-8 safe on Windows PowerShell 5.1).

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

function Invoke-GhJson {
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$GhArgs
    )

    $output = & gh @GhArgs 2>$null
    if ($LASTEXITCODE -ne 0) { return @() }

    $raw = ($output | Out-String).Trim()
    if (-not $raw -or $raw -eq "[]") { return @() }

    $parsed = $raw | ConvertFrom-Json
    if ($null -eq $parsed) { return @() }

    if ($parsed -is [System.Array]) {
        return $parsed
    }
    return @([object]$parsed)
}

function Invoke-GhText {
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$GhArgs
    )

    $output = & gh @GhArgs 2>&1
    if ($LASTEXITCODE -ne 0) { return $null }
    return ($output | Out-String).Trim()
}

function ConvertTo-GhArray {
    param($Value)

    if ($null -eq $Value) { return @() }
    if ($Value -is [System.Array]) { return $Value }
    return @([object]$Value)
}

function Get-GhSingle {
    param($Value)

    if ($null -eq $Value) { return $null }
    if ($Value -is [System.Array]) {
        if ($Value.Count -gt 0) { return $Value[0] }
        return $null
    }
    return $Value
}

function Normalize-JsonArray {
    param($Value)

    $list = New-Object System.Collections.ArrayList
    foreach ($item in (ConvertTo-GhArray $Value)) {
        [void]$list.Add($item)
    }
    return $list
}

function Write-JsonResult {
    param([hashtable]$Obj)

    foreach ($key in @("needsYourReview", "openPrs", "crossRepoReview")) {
        if ($Obj.ContainsKey($key)) {
            $Obj[$key] = (Normalize-JsonArray $Obj[$key])
        }
    }

    $Obj | ConvertTo-Json -Depth 15 -Compress:$false
}

function Escape-Html {
    param([string]$Text)
    if (-not $Text) { return "" }
    return [System.Net.WebUtility]::HtmlEncode($Text)
}

function Get-PrAutoFindings {
    param(
        [Parameter(Mandatory = $true)]
        $Overview
    )

    $findings = @{
        blockers   = [System.Collections.ArrayList]@()
        warnings   = [System.Collections.ArrayList]@()
        ciChecks   = @()
        mergeable  = $Overview.mergeable
        mergeState = $Overview.mergeStateStatus
        canMerge   = $true
        verdict    = "READY"
    }

    if ($Overview.isDraft) {
        [void]$findings.blockers.Add("Draft PR")
        $findings.canMerge = $false
    }

    if ($Overview.mergeable -eq "CONFLICTING") {
        [void]$findings.blockers.Add("Merge conflicts with base branch")
        $findings.canMerge = $false
    }
    elseif ($Overview.mergeable -eq "UNKNOWN") {
        [void]$findings.warnings.Add("Mergeability not yet computed by GitHub")
    }

    switch ($Overview.mergeStateStatus) {
        "BEHIND" { [void]$findings.warnings.Add("Head branch is behind base; rebase or merge base first") }
        "BLOCKED" { [void]$findings.blockers.Add("Blocked by branch protection or required checks"); $findings.canMerge = $false }
        "DIRTY" { [void]$findings.blockers.Add("Unresolved conflicts or dirty merge state"); $findings.canMerge = $false }
        "UNSTABLE" { [void]$findings.warnings.Add("Unstable merge state; watch CI/checks") }
    }

    if ($Overview.reviewDecision -eq "CHANGES_REQUESTED") {
        [void]$findings.blockers.Add("Review requested changes (CHANGES_REQUESTED)")
        $findings.canMerge = $false
    }

    $checks = ConvertTo-GhArray $Overview.statusCheckRollup
    foreach ($check in $checks) {
        $state = $check.state
        $name = if ($check.context) { $check.context } elseif ($check.name) { $check.name } else { "check" }
        $findings.ciChecks += @{
            name       = $name
            state      = $state
            conclusion = $check.conclusion
            url        = $check.targetUrl
        }
        if ($state -in @("FAILURE", "ERROR", "TIMED_OUT", "ACTION_REQUIRED")) {
            [void]$findings.blockers.Add("CI failed: $name ($state)")
            $findings.canMerge = $false
        }
        elseif ($state -in @("PENDING", "IN_PROGRESS", "QUEUED", "WAITING", "REQUESTED")) {
            [void]$findings.warnings.Add("CI pending: $name")
        }
    }

    $binaryExt = @(".dll", ".exe", ".zip", ".tar", ".gz", ".7z", ".so", ".dylib", ".bin", ".pcm", ".wav", ".mp3", ".png", ".jpg", ".jpeg", ".gif", ".pdf")
    $files = ConvertTo-GhArray $Overview.files
    $binaryFiles = @()
    foreach ($file in $files) {
        $path = $file.path
        $ext = [System.IO.Path]::GetExtension($path).ToLowerInvariant()
        if ($binaryExt -contains $ext) {
            $binaryFiles += $path
        }
        if ($file.additions -gt 500 -or $file.deletions -gt 500) {
            [void]$findings.warnings.Add("Large diff: $path (+$($file.additions)/-$($file.deletions))")
        }
    }
    if ($binaryFiles.Count -gt 0) {
        $sample = ($binaryFiles | Select-Object -First 5) -join ", "
        $suffix = if ($binaryFiles.Count -gt 5) { " (+$($binaryFiles.Count) total)" } else { "" }
        [void]$findings.warnings.Add("Binary/large assets: $sample$suffix")
    }

    if ($findings.blockers.Count -gt 0) {
        $findings.verdict = "BLOCKED"
    }
    elseif ($findings.warnings.Count -gt 0) {
        $findings.verdict = "CAUTION"
    }
    else {
        $findings.verdict = "READY"
    }

    return $findings
}

function ConvertTo-HtmlList {
    param(
        [string[]]$Items,
        [string]$EmptyText = "N/A"
    )

    if (-not $Items -or $Items.Count -eq 0) {
        return "<li class=""empty"">$(Escape-Html $EmptyText)</li>"
    }

    $html = ""
    foreach ($item in $Items) {
        if ($item) {
            $html += "<li>$(Escape-Html $item)</li>"
        }
    }
    return $html
}
