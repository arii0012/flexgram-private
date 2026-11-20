# FlexGram auto launcher (Windows PowerShell)
Set-Location $PSScriptRoot
$ErrorActionPreference = "Continue"

Write-Host "=== FlexGram Launcher ===" -ForegroundColor Cyan

# 1. .env
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "Created .env — edit passwords then press Enter"
    notepad .env
    Read-Host
}

# 2. Docker
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "Install Docker Desktop first: https://www.docker.com/products/docker-desktop/" -ForegroundColor Red
    exit 1
}

Write-Host "Starting Docker services..." -ForegroundColor Yellow
docker compose up -d --build

# 3. Wait for API
Write-Host "Waiting for server..." -ForegroundColor Yellow
$ok = $false
for ($i = 0; $i -lt 30; $i++) {
    try {
        $r = Invoke-WebRequest -Uri "http://localhost:8080/" -UseBasicParsing -TimeoutSec 2
        if ($r.StatusCode -eq 200) { $ok = $true; break }
    } catch {}
    Start-Sleep -Seconds 2
}
if (-not $ok) {
    Write-Host "Server did not start. Check: docker compose logs backend" -ForegroundColor Red
    exit 1
}
Write-Host "Server OK: http://localhost:8080" -ForegroundColor Green

# 4. Client
$exe1 = "mobile\build\windows\x64\runner\Release\flexgram.exe"
$exe2 = "mobile\build\windows\runner\Release\flexgram.exe"
if (Test-Path $exe1) {
    Start-Process $exe1
} elseif (Test-Path $exe2) {
    Start-Process $exe2
} elseif (Get-Command flutter -ErrorAction SilentlyContinue) {
    Set-Location mobile
    flutter pub get
    Start-Process powershell -ArgumentList "-NoExit", "-Command", "flutter run -d windows"
    Set-Location ..
} else {
    Write-Host @"

Flutter app not built yet. Run once:
  cd mobile
  flutter pub get
  flutter build windows --release

Then run this script again.
Server is already running at http://localhost:8080
"@ -ForegroundColor Yellow
}

Write-Host "`nDone. Stop server: docker compose down" -ForegroundColor Cyan
