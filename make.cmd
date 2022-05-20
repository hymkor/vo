@setlocal
@set "PROMPT=$G "
@for /F %%I in ('cd') do set "NAME=%%~nI"
@call :"%1"
@endlocal
@exit /b

:""
    set GOARCH=386
    @for /D %%I in (internal\*) do pushd "%%~I" & go fmt & popd "%%~I"
    pushd cmd\vo & go fmt && go build & popd
    @exit /b

:"install"
    @for /F "skip=1" %%I in ('where "%NAME%.exe"') do call :copy "%NAME%.exe" "%%I"
    @exit /b

:copy
    copy "%~1" "%~2" >nul
    if not errorlevel 1 goto copydone
    move "%~2" "%~2-%DATE:/=%_%TIME::=%
    copy "%~1" "%~2" >nul
:copydone
    echo %1 to %2
    @exit /b

:"package"
    @for /F %%I in ('git describe --tags') do set "VER=%%~I"
    zip -j "%NAME%-%VER%-windows-386.zip" "cmd\vo\vo.exe"
    @exit /b
