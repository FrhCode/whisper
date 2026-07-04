$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Force -Path dist | Out-Null

go run .\tools\write-icon
if (-not (Get-Command go-winres -ErrorAction SilentlyContinue)) {
  go install github.com/tc-hib/go-winres@latest
}
Remove-Item .\cmd\whispr\rsrc_windows_*.syso -ErrorAction SilentlyContinue
Remove-Item .\rsrc_windows_*.syso -ErrorAction SilentlyContinue

go-winres make
if ($LASTEXITCODE -ne 0) { throw "go-winres failed" }
Move-Item .\rsrc_windows_*.syso .\cmd\whispr\ -Force

go build -ldflags="-H windowsgui" -o dist\whispr.exe .\cmd\whispr
if ($LASTEXITCODE -ne 0) { throw "go build failed" }
Copy-Item config.example.json dist\config.example.json -Force

Write-Host "Built dist\whispr.exe"
Write-Host "Next: copy dist\config.example.json to dist\config.json and fill API keys/model."
