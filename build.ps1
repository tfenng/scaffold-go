#Requires -Version 5.1
<#
.SYNOPSIS
    Builds the frontend static export and produces the single self-contained binary.
.DESCRIPTION
    Windows/PowerShell equivalent of `make web` + `make release`.
    Requires Node.js; pnpm is preferred, npm is used as a fallback.

    `next build` is invoked directly instead of `pnpm build` / `npm run build`
    because the package script calls a POSIX shell script
    (scripts/postbuild-static-export.sh) that cannot run under PowerShell.
    The root redirect page is generated here instead.
.EXAMPLE
    .\build.ps1
.EXAMPLE
    .\build.ps1 -SkipInstall
#>
[CmdletBinding()]
param(
    [string]$WebDir = "scaffold-web",
    [string]$DistDir = "internal/webui/dist",
    [string]$Output = "bin/scaffold-api.exe",
    [switch]$SkipInstall
)

$ErrorActionPreference = "Stop"
Set-Location -LiteralPath $PSScriptRoot

$pnpm = Get-Command pnpm.cmd -ErrorAction SilentlyContinue
if (-not $pnpm) { $pnpm = Get-Command pnpm -ErrorAction SilentlyContinue }

$npm = Get-Command npm.cmd -ErrorAction SilentlyContinue
if (-not $npm) { $npm = Get-Command npm -ErrorAction SilentlyContinue }

if (-not $pnpm -and -not $npm) {
    throw @"
Node.js is required to build the frontend.
Install it, then re-run this script:

    winget install OpenJS.NodeJS.LTS

After opening a new shell, either install pnpm or let this script use npm:

    npm install -g pnpm
"@
}

Push-Location -LiteralPath $WebDir
try {
    if ($pnpm) {
        if (-not $SkipInstall) { & $pnpm.Source install --frozen-lockfile }
        & $pnpm.Source exec next build
    }
    else {
        if (-not $SkipInstall) { & $npm.Source install --no-package-lock }

        $npx = Get-Command npx.cmd -ErrorAction SilentlyContinue
        if (-not $npx) { $npx = Get-Command npx -ErrorAction SilentlyContinue }
        if (-not $npx) { throw "npx was not found; install Node.js first." }
        & $npx.Source next build
    }
}
finally {
    Pop-Location
}

$outDir = Join-Path $WebDir "out"
if (-not (Test-Path -LiteralPath $outDir)) {
    throw "Frontend build output not found: $outDir"
}

# Mirror scripts/postbuild-static-export.sh so the root path redirects to /users/.
$redirectHtml = @'
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>Redirecting...</title>
    <meta http-equiv="refresh" content="0; url=/users/" />
    <meta name="robots" content="noindex" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="canonical" href="/users/" />
  </head>
  <body>
    <main>
      <p>Redirecting to <a href="/users/">/users/</a>...</p>
    </main>
  </body>
</html>
'@
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText((Join-Path $outDir "index.html"), $redirectHtml, $utf8NoBom)

if (Test-Path -LiteralPath $DistDir) {
    Remove-Item -LiteralPath $DistDir -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $DistDir | Out-Null
Copy-Item -Path (Join-Path $outDir "*") -Destination $DistDir -Recurse -Force
New-Item -ItemType File -Force -Path (Join-Path $DistDir ".gitkeep") | Out-Null

$outputDir = Split-Path -Parent $Output
if ($outputDir) {
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
}
go build -trimpath -o $Output .

Write-Host "Built $Output"
