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
    for /F %%I in ('where %EXE%') do copy "%EXE%" "%%I" >nul && echo %EXE% to %%I
    exit /b
