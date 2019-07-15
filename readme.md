vf1s is the tool which supports these works all on the command line.

- Build projects for the any-version's Visual Studio.
- Show the product information: version, timestamp and check-sum.
- The library to parse project-xml-files and portable-executables.

Build projects
==============

Look for devenv.com and call it to build a product.

- for 2010, see `%VS100COMNTOOLS%`
- for 2013, see `%VS120COMNTOOLS%`
- for 2015, see `%VS140COMNTOOLS%`
- for 2017, call `vswhere -version [15.0,16.0)`
- for 2019, call `vswhere -version [16.0,17.0)`


Build the release version
-------------------------

```
$ vf1s.exe -v -r WorkReport.sln
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
$ vf1s.exe -v -d
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

Show the product information 
============================

Show files in-line (separated by TAB)
-----------------------------------

```
$ vf1s.exe -ls
"bin\Debug\WorkReport.exe"      "bin\Release\WorkReport.exe"
```

These files are built by the solution files in the current directory.

Show files multi-line with some information.
--------------------------------------------

```
$ vf1s.exe -ll
bin\Debug\WorkReport.exe
        1.0.0.11          1.0.0.11          2019-07-14 17:53:21
        47104 bytes  md5sum:1a91d74d9594b2bc575c5bf5327dfa5e
bin\Release\WorkReport.exe
        1.0.0.11          1.0.0.11          2019-07-14 17:53:26
        44032 bytes  md5sum:6bfb25c0bb155e6a4b1c05e152eb9be0
```

These files are built by the solution files in the current directory.


Show files specified by path
----------------------------

```
$ vf1s.exe -showver bin\Release\WorkReport.exe
bin\Release\WorkReport.exe
        1.0.0.11          1.0.0.11          2019-07-14 17:53:26
        44032 bytes  md5sum:6bfb25c0bb155e6a4b1c05e152eb9be0
```

Help
====

```
$ vf1s.exe -h
Usage of vf1s.exe:
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
  -ll
        list products
  -ls
        list products
  -n    dry run
  -r    build configurations contains /Release/
  -re
        rebuild
  -showver string
        show version
  -v    verbose
  -w    show warnings
```

The library
===========

This product has these sub-package.

- [peinfo](https://godoc.org/github.com/zetamatta/vf1s/peinfo)
    - The library which gets version information from binary imag of executables.
- [projs](https://godoc.org/github.com/zetamatta/vf1s/projs)
    - The library which parses `*.vcxproj`, `*.vbproj` and `*.csproj`.
