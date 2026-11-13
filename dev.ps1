#Requires -Version 5.1
<#
.SYNOPSIS
    Runs the API locally with environment variables loaded from an env file.
.DESCRIPTION
    Windows/PowerShell equivalent of `make dev`.
.EXAMPLE
    .\dev.ps1
.EXAMPLE
    .\dev.ps1 -EnvFile .env.production
#>
[CmdletBinding()]
param(
    [string]$EnvFile = ".env.dev"
)

$ErrorActionPreference = "Stop"
Set-Location -LiteralPath $PSScriptRoot

if (Test-Path -LiteralPath $EnvFile) {
    Get-Content -LiteralPath $EnvFile | ForEach-Object {
        $line = $_.Trim()
        if ($line -eq "" -or $line.StartsWith("#")) { return }

        $separator = $line.IndexOf("=")
        if ($separator -lt 1) { return }

        $name = $line.Substring(0, $separator).Trim()
        $value = $line.Substring($separator + 1).Trim().Trim('"').Trim("'")
        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}
else {
    Write-Host "Env file not found: $EnvFile (using defaults)"
}

go run .
