package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func envToCom(envname string) (string, error) {
	env := os.Getenv(envname)
	if env == "" {
		return "", fmt.Errorf("%%%s%% not set", envname)
	}
	env = strings.ReplaceAll(env, "Tools", "IDE")
	com := filepath.Join(env, "devenv.com")
	if fd, err := os.Open(com); err == nil {
		fd.Close()
		return com, nil
	}
	return "", fmt.Errorf("%s not found", com)
}

func seek2010() (string, error) {
	return envToCom("VS100COMNTOOLS")
}

func seek2013() (string, error) {
	return envToCom("VS120COMNTOOLS")
}

func seek2015() (string, error) {
	return envToCom("VS140COMNTOOLS")
}

func seek2017() (string, error) {
	return ProductPath("-version", "[15.0,16.0)")
}

func seek2019() (string, error) {
	return ProductPath("-version", "[16.0,17.0)")
}

func seekLatest() (string, error) {
	return ProductPath("-latest")
}

var versionToSeekfunc = map[string]func() (string, error){
	"2010": seek2010,
	"2013": seek2013,
	"2015": seek2015,
	"2017": seek2017,
	"2019": seek2019,
}

var searchList = []func() (string, error){
	seekLatest,
	seek2015,
	seek2013,
	seek2010,
}

var useVs2010 = flag.Bool("2010", false, "use Visual Studio 2010")
var useVs2013 = flag.Bool("2013", false, "use Visual Studio 2013")
var useVs2015 = flag.Bool("2015", false, "use Visual Studio 2015")
var useVs2017 = flag.Bool("2017", false, "use Visual Studio 2017")
var useVs2019 = flag.Bool("2019", false, "use Visual Studio 2019")

func seekDevenv(solutionFile string, log io.Writer) (compath string, err error) {
	// option to force
	if *useVs2019 {
		compath, err = seek2019()
	}
	if *useVs2017 {
		compath, err = seek2017()
	}
	if *useVs2015 {
		compath, err = seek2015()
	}
	if *useVs2013 {
		compath, err = seek2013()
	}
	if *useVs2010 {
		compath, err = seek2010()
	}
	if err == nil && compath != "" {
		return
	}
	if err != nil {
		fmt.Fprintln(log, err)
	}

	// solution files
	var ver string
	ver, err = findVersionInSolution(solutionFile)
	if err == nil && ver != "" {
		if f := versionToSeekfunc[ver]; f != nil {
			fmt.Fprintf(log, "%s: word '%s' found.\n", solutionFile, ver)
			compath, err = f()
			if compath != "" && err == nil {
				return
			}
			if err != nil {
				fmt.Fprintln(log, err)
			}
		}
	}

	// latest version
	for _, f := range searchList {
		compath, err = f()
		if compath != "" && err == nil {
			return
		}
		if err != nil {
			fmt.Fprintln(log, err)
		}
	}
	return "", io.EOF
}

func run(devenv string, param ...string) error {
	cmd1 := exec.Command(string(devenv), param...)
	cmd1.Stdin = os.Stdin
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	fmt.Printf("\"%s\" \"%s\"\n", devenv, strings.Join(param, "\" \""))
	return cmd1.Run()
}

var buildDebug = flag.Bool("d", false, "build debug")
var buildAll = flag.Bool("a", false, "build all(debug and release)")
var doRebuild = flag.Bool("r", false, "rebuld")
var openIde = flag.Bool("i", false, "open ide")
var flagConfig = flag.String("c", "", "configuration(Release,Debug..)")

func _main() error {
	flag.Parse()

	args := flag.Args()
	sln, err := FindSolution(args)
	if err != nil {
		return err
	}

	devenv, err := seekDevenv(sln, os.Stderr)
	if devenv == "" || err != nil {
		return errors.New("devenv.com not found")
	}

	if *openIde {
		return run(devenv, sln)
	}
	action := "/build"
	if *doRebuild {
		action = "/rebuild"
	}
	if *flagConfig != "" {
		return run(devenv, sln, action, *flagConfig)
	} else if *buildAll {
		if err := run(devenv, sln, action, "Debug"); err != nil {
			return err
		}
		return run(devenv, sln, action, "Release")
	} else if *buildDebug {
		return run(devenv, sln, action, "Debug")
	} else {
		return run(devenv, sln, action, "Release")
	}
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
