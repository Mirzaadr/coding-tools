function Read-Template {
    param([string]$TemplatePath)
    if (-not (Test-Path $TemplatePath)) {
        Write-Error "Template not found: $TemplatePath"
        throw "TemplateNotFoundException"
    }
    return Get-Content -Path $TemplatePath -Raw
}

function Replace-TemplateVariables {
    param(
        [string]$TemplateContent,
        [hashtable]$Variables
    )
    $content = $TemplateContent
    foreach ($key in $Variables.Keys) {
        $content = $content -replace "\{\{$key\}\}", $Variables[$key]
    }
    return $content
}

function Write-GeneratedFile {
    param(
        [string]$OutputPath,
        [string]$TemplateContent,
        [hashtable]$Variables
    )
    $content = Replace-TemplateVariables -TemplateContent $TemplateContent -Variables $Variables
    Write-File -Path $OutputPath -Content $content
}