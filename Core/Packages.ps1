function Install-Packages {
    param(
        [string]$SolutionName,
        [string]$OutputPath,
        [string]$DatabaseType
    )
    Write-Step "Installing NuGet Packages"
    
    Push-Location $OutputPath
    try {
        $src = $global:Config.SrcFolder
        $infraProj = "$src\$SolutionName.Infrastructure\$SolutionName.Infrastructure.csproj"
        
        dotnet add $infraProj package Microsoft.EntityFrameworkCore
        
        if ($DatabaseType -eq "PostgreSQL") {
            dotnet add $infraProj package Npgsql.EntityFrameworkCore.PostgreSQL
        } else {
            dotnet add $infraProj package Microsoft.EntityFrameworkCore.SqlServer
        }

        $appProj = "$src\$SolutionName.Application\$SolutionName.Application.csproj"
        
        dotnet add $appProj package Microsoft.EntityFrameworkCore
        dotnet add $appProj package Microsoft.Extensions.Configuration
        dotnet add $appProj package Mediatr --version 12.5.0
        
        Write-Success "NuGet packages installed."
    } finally {
        Pop-Location
    }
}