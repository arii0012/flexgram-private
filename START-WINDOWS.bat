@echo off
chcp 65001 >nul
title FlexGram Launcher
cd /d "%~dp0"

echo ========================================
echo   FlexGram - avtozapusk
echo ========================================
echo.

REM --- .env ---
if not exist ".env" (
    echo [1/4] Sozdayu .env ...
    copy ".env.example" ".env" >nul
    echo.
    echo  !!! Otkroy .env i postav paroli !!!
    echo  POSTGRES_PASSWORD=mypass123
    echo  REDIS_PASSWORD=redis123
    echo  JWT_SECRET=supersecretkeyatleast32characterslong
    echo.
    notepad .env
    echo Nazhmi lyubuyu klavishu posle sohraneniya .env...
    pause >nul
)

REM --- Docker ---
echo [2/4] Zapuskayu Docker (PostgreSQL + Redis + Backend)...
where docker >nul 2>&1
if errorlevel 1 (
    echo.
    echo  OSHIBKA: Docker ne ustanovlen.
    echo  Skachay Docker Desktop: https://www.docker.com/products/docker-desktop/
    echo  Posle ustanovki perezagruzi PC i zapusti etot bat snova.
    pause
    exit /b 1
)

docker info >nul 2>&1
if errorlevel 1 (
    echo  Docker ne zapushen. Zapuskayu Docker Desktop...
    start "" "C:\Program Files\Docker\Docker\Docker Desktop.exe"
    echo  Zhdu 30 sekund...
    timeout /t 30 /nobreak >nul
)

docker compose up -d --build
if errorlevel 1 (
    echo.
    echo  Oshibka docker compose. Prover Docker Desktop.
    pause
    exit /b 1
)

echo.
echo [3/4] Zhdu poka server podnimetsya...
set /a tries=0
:waitloop
set /a tries+=1
curl -s http://localhost:8080/ >nul 2>&1
if not errorlevel 1 goto server_ok
if %tries% geq 30 (
    echo  Server ne otvetil za 60 sek. Prover: docker compose logs backend
    pause
    exit /b 1
)
timeout /t 2 /nobreak >nul
goto waitloop

:server_ok
echo  Server OK: http://localhost:8080
echo.

REM --- Client ---
echo [4/4] Zapuskayu prilozhenie...

if exist "mobile\build\windows\x64\runner\Release\flexgram.exe" (
    start "" "mobile\build\windows\x64\runner\Release\flexgram.exe"
    goto done
)

if exist "mobile\build\windows\runner\Release\flexgram.exe" (
    start "" "mobile\build\windows\runner\Release\flexgram.exe"
    goto done
)

where flutter >nul 2>&1
if not errorlevel 1 (
    echo  Sobirayu i zapuskayu Flutter Windows...
    cd mobile
    flutter pub get
    start cmd /c "flutter run -d windows"
    cd ..
    goto done
)

echo.
echo  Flutter-klient eshche ne sobran.
echo  Sdelay odin raz:
echo    cd mobile
echo    flutter pub get
echo    flutter build windows --release
echo  Potom snova zapusti START-WINDOWS.bat
echo.
echo  Server uzhe rabotaet na http://localhost:8080
echo  Mozhesh proverit: curl http://localhost:8080/
echo.

:done
echo.
echo ========================================
echo  Server:  http://localhost:8080
echo  Ostanovit: docker compose down
echo ========================================
echo.
pause
