@echo off
call :"%1"
exit /b

:""
    setlocal
    set GOARCH=386
    go fmt
    go build
    endlocal
    exit /b

:"install"
    for /F %%I in ('cd') do set "EXE=%%~nI.exe"
    for /F "skip=1" %%I in ('where %EXE%') do call :copy "%EXE%" "%%I"
    exit /b

:copy
    copy "%~1" "%~2" >nul
    if not errorlevel 1 goto copydone
    move "%~2" "%~2-%DATE:/=%_%TIME::=%
    copy "%~1" "%~2" >nul
:copydone
    echo %1 to %2
    exit /b
