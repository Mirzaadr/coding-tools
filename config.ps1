$global:Config = @{
    TemplateRoot = Join-Path $PSScriptRoot "Templates"
    DefaultProjectName = "CleanArchitectureApp"
    DefaultOutputFolder = ".\output"
    SrcFolder = "src"
    TestsFolder = "tests"
}
