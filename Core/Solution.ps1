function Create-Solution {
    param(
        [string]$SolutionName,
        [string]$OutputPath
    )
    Write-Step "Creating Solution: $SolutionName"
    
    Ensure-DirectoryExists $OutputPath
    Push-Location $OutputPath
    try {
        dotnet new sln -n $SolutionName
        Write-Success "Solution created."
    } catch {
        Write-Error "Failed to create solution."
        throw
    } finally {
        Pop-Location
    }
}

function Create-RootFolders {
    param([string]$OutputPath)
    Ensure-DirectoryExists (Join-Path $OutputPath $global:Config.SrcFolder)
    Ensure-DirectoryExists (Join-Path $OutputPath $global:Config.TestsFolder)
}