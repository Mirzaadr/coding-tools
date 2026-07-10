param(
    [Parameter(Mandatory = $true)]
    [string]$SolutionName = "CleanArchitectureApp",
    
    [string]$OutputPath = "./output",
    [string]$PresentationType = "WebApi",
    
    # [Alias("db")]
    [ValidateSet("PostgreSQL", "SqlServer")]
    [string]$DatabaseType = "PostgreSQL",
    [switch]$NoBuild

    # [string]$ProjectName,

    # [ValidateSet("WebApi", "WebApp")]
    # [string]$Presentation = "WebApi",

    # [string]$OutputPath = ".",

    # [string]$Database = "PostgreSQL"

    # [switch]$NoBuild,
    # [switch]$NoProjectFolder
)

$scriptDir = $PSScriptRoot

# Load Configuration
. (Join-Path $scriptDir "config.ps1")

# Load Utils
. (Join-Path $scriptDir "Utils\Console.ps1")
. (Join-Path $scriptDir "Utils\File.ps1")
. (Join-Path $scriptDir "Utils\Template.ps1")

# Load Core
. (Join-Path $scriptDir "Core\Solution.ps1")
. (Join-Path $scriptDir "Core\Projects.ps1")
. (Join-Path $scriptDir "Core\References.ps1")
. (Join-Path $scriptDir "Core\Packages.ps1")
. (Join-Path $scriptDir "Core\Build.ps1")

# Load Generators
. (Join-Path $scriptDir "Generators\Infrastructure.ps1")
. (Join-Path $scriptDir "Generators\Application.ps1")
. (Join-Path $scriptDir "Generators\Presentation.ps1")
. (Join-Path $scriptDir "Generators\Tests.ps1")

try {
    Write-Step "Starting Clean Architecture Generator"
    
    Create-RootFolders -OutputPath $OutputPath
    Create-Solution -SolutionName $SolutionName -OutputPath $OutputPath
    
    Create-CoreProjects -SolutionName $SolutionName -OutputPath $OutputPath
    Create-PresentationProject -SolutionName $SolutionName -OutputPath $OutputPath -Type $PresentationType
    # Create-TestProjects -SolutionName $SolutionName -OutputPath $OutputPath
    
    Add-ProjectsToSolution -SolutionName $SolutionName -OutputPath $OutputPath
    Configure-ProjectReferences -SolutionName $SolutionName -OutputPath $OutputPath
    
    Install-Packages -SolutionName $SolutionName -OutputPath $OutputPath -DatabaseType $DatabaseType
    
    Generate-Infrastructure -SolutionName $SolutionName -OutputPath $OutputPath
    Generate-Application -SolutionName $SolutionName -OutputPath $OutputPath
    Generate-Presentation -SolutionName $SolutionName -OutputPath $OutputPath
    Generate-Tests -SolutionName $SolutionName -OutputPath $OutputPath
    
    Build-Solution -OutputPath $OutputPath -NoBuild:$NoBuild
    
    Write-Success "Clean Architecture setup completed successfully."
} catch {
    Write-Error "An error occurred during generation: $_"
    exit 1
}