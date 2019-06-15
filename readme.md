vf1s - Visual Studio Commandline Client
=======================================

Search devenv.com's fullpath and call it.

- for 2010, see `%VS100COMNTOOLS%`
- for 2013, see `%VS120COMNTOOLS%`
- for 2015, see `%VS140COMNTOOLS%`
- for 2017, call `vswhere -version [15.0,16.0)`
- for 2019, call `vswhere -version [16.0,17.0)`

```
Usage:
    vf1s {options} [solutionFile]
```

- On default, vf1s builds configurations containing /Release/.
- If `solutionFile` is not given, seek one on the current directory.
- If the version is not given, the version written in the solution file or the latest version of Visual Studio is used.

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
  -d    build configurations containing /Debug/.
  -i    open ide
  -r    rebuild
```
