# PowerShell Build Script for Cerberus Procure
# Usage: 
#   .\make.ps1 wasm
#   .\make.ps1 server
#   .\make.ps1 all
#   .\make.ps1 clean

param (
    [Parameter(Position=0)]
    [ValidateSet("wasm", "server", "frontend-install", "frontend-build", "clean", "all")]
    [string]$Task = "all"
)

$ErrorActionPreference = "Stop"

function Build-Wasm {
    Write-Host ">>> Building WASM..." -ForegroundColor Cyan
    if (!(Test-Path "frontend/public")) {
        New-Item -ItemType Directory -Path "frontend/public" -Force | Out-Null
    }
    
    $env:GOOS = "js"
    $env:GOARCH = "wasm"
    Write-Host "Executing: go build -o frontend/public/main.wasm cmd/wasm/main.go"
    go build -o frontend/public/main.wasm cmd/wasm/main.go
    $env:GOOS = ""
    $env:GOARCH = ""
    
    $goroot = go env GOROOT
    $wasmExecSource = Join-Path $goroot "lib/wasm/wasm_exec.js"
    Write-Host "Copying wasm_exec.js from $wasmExecSource"
    Copy-Item $wasmExecSource -Destination "frontend/public/" -Force
    Write-Host "WASM Build Success!" -ForegroundColor Green
}

function Build-Server {
    Write-Host ">>> Building Server..." -ForegroundColor Cyan
    
    Write-Host "Building Frontend..."
    Push-Location frontend
    npm run build
    Pop-Location
    
    if (!(Test-Path "cmd/server/dist")) {
        New-Item -ItemType Directory -Path "cmd/server/dist" -Force | Out-Null
    }
    
    Write-Host "Syncing Frontend Assets to Server..."
    if (Test-Path "cmd/server/dist/*") {
        Remove-Item "cmd/server/dist/*" -Recurse -Force
    }
    Copy-Item -Path "frontend/dist/*" -Destination "cmd/server/dist/" -Recurse -Force
    
    Write-Host "Compiling Go Server..."
    go build -o server_bin.exe cmd/server/main.go
    Write-Host "Server Build Success: server_bin.exe" -ForegroundColor Green
}

function Frontend-Install {
    Write-Host ">>> Installing Frontend Dependencies..." -ForegroundColor Cyan
    Push-Location frontend
    npm install
    Pop-Location
}

function Clean {
    Write-Host ">>> Cleaning Build Artifacts..." -ForegroundColor Yellow
    $targets = @("server_bin.exe", "frontend/public/main.wasm", "frontend/public/wasm_exec.js", "frontend/dist", "cmd/server/dist")
    foreach ($target in $targets) {
        if (Test-Path $target) {
            Remove-Item $target -Recurse -Force
            Write-Host "Removed: $target"
        }
    }
    Write-Host "Clean Complete." -ForegroundColor Green
}

switch ($Task) {
    "wasm" { Build-Wasm }
    "server" { Build-Server }
    "frontend-install" { Frontend-Install }
    "frontend-build" { 
        Push-Location frontend
        npm run build
        Pop-Location
    }
    "clean" { Clean }
    "all" { 
        Build-Wasm
        Build-Server
    }
}
