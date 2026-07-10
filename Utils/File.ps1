function Ensure-DirectoryExists {
    param([string]$Path)
    if (-not (Test-Path $Path)) {
        New-Item -ItemType Directory -Path $Path | Out-Null
    }
}

function Write-File {
    param(
        [string]$Path,
        [string]$Content
    )
    $dir = Split-Path $Path
    Ensure-DirectoryExists $dir
    Set-Content -Path $Path -Value $Content -Encoding UTF8
}

function Delete-File {
    param([string]$Path)
    if (Test-Path $Path) {
        Remove-Item -Path $Path -Force
    }
}

function Replace-File {
    param(
        [string]$Path,
        [string]$Content
    )
    Write-File -Path $Path -Content $Content
}

function Resolve-Paths {
    param([string]$Path)
    return Resolve-Path $Path | Select-Object -ExpandProperty Path
}
