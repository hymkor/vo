package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ENV_2010 = "VS100COMNTOOLS"
	ENV_2013 = "VS120COMNTOOLS"
	ENV_2015 = "VS140COMNTOOLS"
)

var envs = map[string]string{
	"2010": ENV_2010,
	"2013": ENV_2013,
	"2015": ENV_2015,
}

func findSolution(args []string) ([]string, error) {
	result := []string{}
	for _, name := range args {
		if strings.HasSuffix(strings.ToLower(name), ".sln") {
			result = append(result, name)
		}
	}
	if len(result) > 0 {
		return result, nil
	}
	fd, err := os.Open(".")
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	files, err := fd.Readdir(-1)
	if err != nil {
		return nil, err
	}
	for _, file1 := range files {
		if strings.HasSuffix(strings.ToLower(file1.Name()), ".sln") {
			result = append(result, file1.Name())
		}
	}
	return result, nil
}

func FindSolution(args []string) (string, error) {
	sln, err := findSolution(args)
	if err != nil {
		return "", err
	}
	if len(sln) < 1 {
		return "", errors.New("no solution files")
	}
	if len(sln) >= 2 {
		return "", fmt.Errorf("%s: too may solution files", strings.Join(sln, ", "))
	}
	return sln[0], nil
}

type Devenv string

func NewDevenv(ver string) (Devenv, error) {
	name := envs[ver]
	if name == "" {
		return Devenv(""), fmt.Errorf("%s not found", name)
	}
	env := os.Getenv(name)
	if env == "" {
		return Devenv(""), fmt.Errorf("%s not found", name)
	}
	env = strings.ReplaceAll(env, "Tools", "IDE")
	return Devenv(filepath.Join(env, "devenv.com")), nil
}

func LatestDevEnv() Devenv {
	if *useVs2019 {
		devenv, err := ProductPath("-version", "16")
		if err != nil {
			return Devenv("")
		}
		return Devenv(devenv)
	}
	if *useVs2015 {
		devenv, _ := NewDevenv("2015")
		return devenv
	}
	if *useVs2013 {
		devenv, _ := NewDevenv("2013")
		return devenv
	}
	if *useVs2010 {
		devenv, _ := NewDevenv("2010")
		return devenv
	}
	if devenv, err := ProductPath("-latest"); err == nil {
		return Devenv(devenv)
	}
	for _, name := range [...]string{"2015", "2013", "2010"} {
		devenv, err := NewDevenv(name)
		if err == nil {
			return devenv
		}
	}
	return ""
}

func (devenv Devenv) Run(param ...string) error {
	cmd1 := exec.Command(string(devenv), param...)
	cmd1.Stdin = os.Stdin
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	fmt.Printf("%s %s\n", devenv, strings.Join(param, " "))
	return cmd1.Run()
}

const productPath = "productPath: "

func ProductPath(args ...string) (string, error) {
	vswhere, err := exec.LookPath("vswhere")
	if err != nil {
		return "", err
	}
	cmd1 := exec.Command(vswhere, args...)
	cmd1.Stdin = os.Stdin
	cmd1.Stderr = os.Stderr
	in, err := cmd1.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer in.Close()
	cmd1.Start()
	sc := bufio.NewScanner(in)
	for sc.Scan() {
		text := sc.Text()
		if strings.HasPrefix(text, productPath) {
			exe := text[len(productPath):]
			suffix := filepath.Ext(exe)
			com := exe[:len(exe)-len(suffix)] + ".com"
			return com, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

var useVs2010 = flag.Bool("2010", false, "use Visual Studio 2010")
var useVs2013 = flag.Bool("2013", false, "use Visual Studio 2013")
var useVs2015 = flag.Bool("2015", false, "use Visual Studio 2015")
var useVs2019 = flag.Bool("2019", false, "use Visual Studio 2019")
var buildDebug = flag.Bool("d", false, "build debug")
var doRebuild = flag.Bool("r", false, "rebuld")

func _main() error {
	flag.Parse()

	devenv := LatestDevEnv()
	if devenv == "" {
		return errors.New("devenv.com not found")
	}
	args := flag.Args()
	sln, err := FindSolution(args)
	if err != nil {
		return err
	}

	action := "/build"
	if *doRebuild {
		action = "/rebuild"
	}

	if *buildDebug {
		return devenv.Run(sln, action, "Debug")
	} else {
		return devenv.Run(sln, action, "Release")
	}
	return nil
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
