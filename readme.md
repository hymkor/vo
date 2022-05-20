`vo.exe` reads `*.sln` and `*.*proj` files in the current directory, finds the appropriate devenv.com's fullpath and calls it for these purpose.

- Start Visual Studio (`vo ide`)
- Build the application (`vo build`)
- Show the executables' information. (`vo ls` / `vo list`)

```
$ vo help
NAME:
   vo.exe - Visual studio solution commandline Operator

USAGE:
   vo.exe [global options] command [command options] [arguments...]

COMMANDS:
   ide      start visual-studio associated the solution with no options
   build    call devenv.com associated the solution with /build option
   rebuild  call devenv.com associated the solution with /rebuild option
   ls       list up executables inline
   list     list up executables and thier version-information with long format
   showver  Show the version information for executables given by parameters
   eval     eval the equation given by parameter
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --2010      use Visual Studio 2010 (default: false)
   --2013      use Visual Studio 2013 (default: false)
   --2015      use Visual Studio 2015 (default: false)
   --2017      use Visual Studio 2017 (default: false)
   --2019      use Visual Studio 2019 (default: false)
   -w          show warnings (default: false)
   -v          verbose (default: false)
   --help, -h  show help (default: false)
```

Build projects
==============

Look for devenv.com and call it to build a product.

- for 2010, see `%VS100COMNTOOLS%`
- for 2013, see `%VS120COMNTOOLS%`
- for 2015, see `%VS140COMNTOOLS%`
- for 2017, call `vswhere -version [15.0,16.0)`
- for 2019, call `vswhere -version [16.0,17.0)`


Build with the default configuration
------------------------------------

```
$ vo -v build
WorkReport.sln: comment version: 2010
WorkReport.sln: default version:
WorkReport.sln: minimum version:
WorkReport.sln: required ToolsVersion is '4.0'.
WorkReport.sln: try to use Visual Studio 2010.
"C:\Program Files (x86)\Microsoft Visual Studio 10.0\Common7\IDE\devenv.com" "WorkReport.sln" "/build"

Microsoft(R) Visual Studio Version 10.0.40219.1.
Copyright (C) Microsoft Corp. All rights reserved.
------ ビルド開始: プロジェクト: WorkReport, 構成: Release x86 ------
  WorkReport -> Z:\Share\Src\github.com\xxxxxxxx\workreport\bin\Release\WorkReport.exe
========== ビルド: 正常終了または最新の状態 1、失敗 0、スキップ 0 ==========
```

Build the debug version
-----------------------

```
$ vo -v build -d
WorkReport.sln: comment version: 2010
WorkReport.sln: default version:
WorkReport.sln: minimum version:
WorkReport.sln: required ToolsVersion is '4.0'.
WorkReport.sln: try to use Visual Studio 2010.
"C:\Program Files (x86)\Microsoft Visual Studio 10.0\Common7\IDE\devenv.com" "WorkReport.sln" "/build" "Debug|x86"

Microsoft(R) Visual Studio Version 10.0.40219.1.
Copyright (C) Microsoft Corp. All rights reserved.
------ ビルド開始: プロジェクト: WorkReport, 構成: Debug x86 ------
  WorkReport -> Z:\Share\Src\github.com\xxxxxxxx\workreport\bin\Debug\WorkReport.exe
========== ビルド: 正常終了または最新の状態 1、失敗 0、スキップ 0 ==========
```

When the solution filename is omitted, use the solution file on the current directory.

Show the product information 
============================

Show files in-line (separated by TAB)
-----------------------------------

```
$ vo ls
"bin\Debug\WorkReport.exe"      "bin\Release\WorkReport.exe"
```

These files are built by the solution files in the current directory.

Show files multi-line with some information.
--------------------------------------------

```
$ vo list
WorkReport.sln:
  WorkReport.csproj:
    Debug|x86:
      bin\Debug\WorkReport.exe
        3.0.0.1           3.0.0.1           2022-05-16 18:42:26
        66560 bytes  md5sum:9c4d0d5b243e60ff51784174117dcca5
    Release|x86:
      bin\Release\WorkReport.exe
        3.0.1.0           3.0.1.0           2022-05-16 18:47:24
        61952 bytes  md5sum:3443f47a1ba2c66bcc85116a8f8a3009
```

These files are built by the solution files in the current directory.


Show files specified by path
----------------------------

```
$ vo showver bin\Release\WorkReport.exe
bin\Release\WorkReport.exe
        1.0.0.16          1.0.0.16          2020-03-16 11:42:44
        50688 bytes  md5sum:1fcbf90db2a4824cac4aa8e936f94ce6
```
