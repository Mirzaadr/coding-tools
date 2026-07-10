function Generate-AppDbContext {
    param([string]$SolutionName, [string]$OutputPath)
    
    $templatePath = Join-Path $global:Config.TemplateRoot "Infrastructure\AppDbContext.cs.tpl"
    if (-not (Test-Path $templatePath)) { return }
    
    $templateContent = Read-Template -TemplatePath $templatePath
    $vars = @{
        "ProjectName" = "$SolutionName.Infrastructure"
    }
    
    $outDir = Join-Path $OutputPath "$($global:Config.SrcFolder)\$SolutionName.Infrastructure\Persistence"
    Ensure-DirectoryExists $outDir
    $outPath = Join-Path $outDir "AppDbContext.cs"
    
    Write-GeneratedFile -OutputPath $outPath -TemplateContent $templateContent -Variables $vars
}

function Generate-DependencyInjection {
    param([string]$SolutionName, [string]$OutputPath)
    
    $templatePath = Join-Path $global:Config.TemplateRoot "Infrastructure\DependencyInjection.cs.tpl"
    if (-not (Test-Path $templatePath)) { return }
    
    $templateContent = Read-Template -TemplatePath $templatePath
    $vars = @{
        "ProjectName" = "$SolutionName.Infrastructure"
    }
    
    $outDir = Join-Path $OutputPath "$($global:Config.SrcFolder)\$SolutionName.Infrastructure"
    Ensure-DirectoryExists $outDir
    $outPath = Join-Path $outDir "DependencyInjection.cs"
    
    Write-GeneratedFile -OutputPath $outPath -TemplateContent $templateContent -Variables $vars
}

function Generate-Infrastructure {
    param([string]$SolutionName, [string]$OutputPath)
    Write-Step "Generating Infrastructure Layer Files"
    
    Generate-AppDbContext -SolutionName $SolutionName -OutputPath $OutputPath
    Generate-DependencyInjection -SolutionName $SolutionName -OutputPath $OutputPath
}