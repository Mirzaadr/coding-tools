function Add-ProjectsToSolution {
    param([string]$SolutionName, [string]$OutputPath)
    Write-Step "Adding Projects to Solution"
    
    Push-Location $OutputPath
    try {
        $slnFile = "$SolutionName.sln"
        if (Test-Path "$SolutionName.slnx") {
            $slnFile = "$SolutionName.slnx"
        }
        
        $srcProjects = Get-ChildItem -Path $global:Config.SrcFolder -Recurse -Filter "*.csproj"
        foreach ($proj in $srcProjects) {
            dotnet sln $slnFile add $proj.FullName
        }
        
        # $testProjects = Get-ChildItem -Path $global:Config.TestsFolder -Recurse -Filter "*.csproj"
        # foreach ($proj in $testProjects) {
        #     dotnet sln $slnFile add $proj.FullName
        # }
        Write-Success "Projects added to solution."
    } finally {
        Pop-Location
    }
}

function Configure-ProjectReferences {
    param([string]$SolutionName, [string]$OutputPath)
    Write-Step "Configuring Project References"
    
    Push-Location $OutputPath
    try {
        $src = $global:Config.SrcFolder
        
        # Application -> Domain
        dotnet add "$src\$SolutionName.Application\$SolutionName.Application.csproj" reference "$src\$SolutionName.Domain\$SolutionName.Domain.csproj"
        
        # Infrastructure -> Application
        dotnet add "$src\$SolutionName.Infrastructure\$SolutionName.Infrastructure.csproj" reference "$src\$SolutionName.Application\$SolutionName.Application.csproj"
        
        # Presentation -> Application, Infrastructure
        dotnet add "$src\$SolutionName.Presentation\$SolutionName.Presentation.csproj" reference "$src\$SolutionName.Application\$SolutionName.Application.csproj"
        dotnet add "$src\$SolutionName.Presentation\$SolutionName.Presentation.csproj" reference "$src\$SolutionName.Infrastructure\$SolutionName.Infrastructure.csproj"
        
        Write-Success "Project references configured."
    } finally {
        Pop-Location
    }
}
