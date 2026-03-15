# Agent Envs - Windows 安装脚本
# 用法: powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"

$AppName = "agent-envs"
$Binary = "$AppName.exe"
$InstallDir = "$env:LOCALAPPDATA\Programs\$AppName"

Write-Host "🔨 编译 $AppName..." -ForegroundColor Cyan

# 获取版本信息
$Version = git describe --tags --always --dirty 2>$null
if (-not $Version) { $Version = "dev" }
$BuildTime = (Get-Date -Format "yyyy-MM-dd_HH:mm:ss")

$LdFlags = "-s -w -X main.version=$Version -X main.buildTime=$BuildTime"

go build -trimpath -ldflags $LdFlags -o $Binary ./cmd/agent-envs
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 编译失败" -ForegroundColor Red
    exit 1
}
Write-Host "✅ 编译完成" -ForegroundColor Green

# 创建安装目录
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# 复制二进制文件
Write-Host "📦 安装到 $InstallDir..." -ForegroundColor Cyan
Copy-Item -Path $Binary -Destination "$InstallDir\$Binary" -Force
Remove-Item -Path $Binary -Force

# 检查 PATH 是否已包含安装目录
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "🔧 添加 $InstallDir 到用户 PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
    Write-Host "✅ PATH 已更新（新终端窗口生效）" -ForegroundColor Green
} else {
    Write-Host "✅ PATH 中已包含安装目录" -ForegroundColor Green
}

Write-Host ""
Write-Host "🎉 安装完成！运行 '$AppName --version' 验证" -ForegroundColor Green
