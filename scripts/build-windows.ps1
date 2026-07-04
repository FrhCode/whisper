$ErrorActionPreference = "Stop"

New-Item -ItemType Directory -Force -Path dist | Out-Null

go run .\tools\write-icon
if (-not (Get-Command go-winres -ErrorAction SilentlyContinue)) {
  go install github.com/tc-hib/go-winres@latest
}
go-winres make

go build -ldflags="-H windowsgui" -o dist\whispr.exe .\cmd\whispr
Copy-Item config.example.json dist\config.example.json -Force

Write-Host "Built dist\whispr.exe"
Write-Host "Next: copy dist\config.example.json to dist\config.json and fill API keys/model."
