@echo off
setlocal
for /F %%I in ('cd') do set "NAME=%%~nI"
call :"%1"
endlocal
exit /b

:""
    setlocal
    set GOARCH=386
    go fmt
    go build
    endlocal
    exit /b

:"install"
    for /F "skip=1" %%I in ('where "%NAME%.exe"') do call :copy "%NAME%.exe" "%%I"
    exit /b

:copy
    copy "%~1" "%~2" >nul
    if not errorlevel 1 goto copydone
    move "%~2" "%~2-%DATE:/=%_%TIME::=%
    copy "%~1" "%~2" >nul
:copydone
    echo %1 to %2
    exit /b

:"package"
    set /P "VER=Version ? "
    zip "%NAME%-%VER%-windows-386.zip" "%NAME%.exe"
    exit /b
