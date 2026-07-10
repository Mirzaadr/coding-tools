function Build-Solution {
    param(
        [string]$OutputPath,
        [switch]$NoBuild
    )
    if ($NoBuild) {
        Write-Info "Skipping build."
        return
    }
    
    Write-Step "Building Solution"
    
    Push-Location $OutputPath
    try {
        dotnet build
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Build succeeded."
        } else {
            Write-Error "Build failed."
        }
    } finally {
        Pop-Location
    }
}
