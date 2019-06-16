vf1s - Visual Studio Commandline Client
=======================================

Look for devenv.com and call it to build a product.

- for 2010, see `%VS100COMNTOOLS%`
- for 2013, see `%VS120COMNTOOLS%`
- for 2015, see `%VS140COMNTOOLS%`
- for 2017, call `vswhere -version [15.0,16.0)`
- for 2019, call `vswhere -version [16.0,17.0)`

Build the release version
-------------------------

```
$ vf1s.exe -r WorkReport.sln
WorkReport.sln: word '2010' found.
%VS100COMNTOOLS% is not set.
look for other versions of Visual Studio.
found 'C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\Common7\IDE\devenv.com'
"C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\Common7\IDE\devenv.com" "WorkReport.sln" "/build" "Release|x86"

Microsoft Visual Studio 2019 RC バージョン 16.0.29009.5。
Copyright (C) Microsoft Corp. All rights reserved.
========== ビルド: 0 正常終了、0 失敗、1 更新不要、0 スキップ ==========
```

Build the debug version
-----------------------

```
$ vf1s.exe -d
WorkReport.sln: word '2010' found.
%VS100COMNTOOLS% is not set.
look for other versions of Visual Studio.
found 'C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\Common7\IDE\devenv.com'
"C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\Common7\IDE\devenv.com" "WorkReport.sln" "/build" "Debug|x86"

Microsoft Visual Studio 2019 RC バージョン 16.0.29009.5。
Copyright (C) Microsoft Corp. All rights reserved.
========== ビルド: 0 正常終了、0 失敗、1 更新不要、0 スキップ ==========
```

When the solution filename is omitted, use the solution file on the current directory.

List up products
----------------

Now, supporting .csproj and .vbproj

```
$ vf1s.exe -ls
bin\Debug\WorkReport.exe
bin\Release\WorkReport.exe
```

Help
----

```
$ vf1s.exe -h
Usage of C:\Users\hymko\go\bin\vf1s.exe:
  -2010
        use Visual Studio 2010
  -2013
        use Visual Studio 2013
  -2015
        use Visual Studio 2015
  -2017
        use Visual Studio 2017
  -2019
        use Visual Studio 2019
  -a    build all configurations
  -c string
        specify the configuraion to build
  -d    build configurations contains /Debug/
  -i    open ide
  -ls
        list products
  -n    dry run
  -r    build configurations contains /Release/
  -re
        rebuild
```
