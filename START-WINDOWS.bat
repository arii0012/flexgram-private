@echo off
chcp 65001 >nul
title FlexGram
cd /d "%~dp0"

echo.
echo   FlexGram - zapusk
echo.

if not exist ".env" (
  copy ".env.example" ".env" >nul
  echo Sozdan .env - paroli po umolchaniyu.
  echo Esli nado - otredaktiruy .env
  echo.
)

where docker >nul 2>&1
if errorlevel 1 (
  echo Nuzhen Docker Desktop:
  echo https://www.docker.com/products/docker-desktop/
  pause
  exit /b 1
)

echo Zapuskayu server...
docker compose up -d --build

echo Zhdu server...
:wait
curl -s http://localhost:8080/ >nul 2>&1
if errorlevel 1 (
  timeout /t 2 /nobreak >nul
  goto wait
)

echo.
echo   Server rabotaet!
echo   Otkryvayu messenger v brauzere...
echo.
echo   S drugom: drug vvodit v pole "Adres servera"
echo   http://ВАШ_IP:8080
echo   (IP smotret: ipconfig)
echo.

start http://localhost:8080/app/

echo Okno mozhno ne zakryvat. Ostanovit server: docker compose down
pause
