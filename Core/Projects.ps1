function Create-Project {
    param(
        [string]$ProjectType,
        [string]$ProjectName,
        [string]$OutputPath,
        [string]$Folder
    )
    $projectDir = Join-Path (Join-Path $OutputPath $Folder) $ProjectName
    Ensure-DirectoryExists $projectDir
    
    Write-Info "Creating $ProjectType project: $ProjectName"
    Push-Location (Join-Path $OutputPath $Folder)
    try {
        dotnet new $ProjectType -o $ProjectName -n $ProjectName
    } finally {
        Pop-Location
    }
}

function Create-CoreProjects {
    param([string]$SolutionName, [string]$OutputPath)
    Write-Step "Creating Core Projects"
    
    Create-Project -ProjectType "classlib" -ProjectName "$SolutionName.Domain" -OutputPath $OutputPath -Folder $global:Config.SrcFolder
    Create-Project -ProjectType "classlib" -ProjectName "$SolutionName.Application" -OutputPath $OutputPath -Folder $global:Config.SrcFolder
    Create-Project -ProjectType "classlib" -ProjectName "$SolutionName.Infrastructure" -OutputPath $OutputPath -Folder $global:Config.SrcFolder
}

function Create-PresentationProject {
    param(
        [string]$SolutionName,
        [string]$OutputPath,
        [string]$Type
    )
    Write-Step "Creating Presentation Project ($Type)"
    
    $template = if ($Type -eq "MVC") { "mvc" } else { "webapi" }
    Create-Project -ProjectType $template -ProjectName "$SolutionName.Presentation" -OutputPath $OutputPath -Folder $global:Config.SrcFolder
}

function Create-TestProjects {
    param([string]$SolutionName, [string]$OutputPath)
    Write-Step "Creating Test Projects"
    
    Create-Project -ProjectType "xunit" -ProjectName "$SolutionName.Domain.Tests" -OutputPath $OutputPath -Folder $global:Config.TestsFolder
    Create-Project -ProjectType "xunit" -ProjectName "$SolutionName.Application.Tests" -OutputPath $OutputPath -Folder $global:Config.TestsFolder
    Create-Project -ProjectType "xunit" -ProjectName "$SolutionName.Infrastructure.Tests" -OutputPath $OutputPath -Folder $global:Config.TestsFolder
    Create-Project -ProjectType "xunit" -ProjectName "$SolutionName.Presentation.Tests" -OutputPath $OutputPath -Folder $global:Config.TestsFolder
}