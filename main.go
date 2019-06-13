package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
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
	if len(sln) >= 2 {
		return "", fmt.Errorf("which solution: %s", strings.Join(sln, ", "))
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

func vswhere() (map[string]string, error) {
	cmd1 := exec.Command("vswhere")
	cmd1.Stdin = os.Stdin
	cmd1.Stderr = os.Stderr
	in, err := cmd1.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer in.Close()
	cmd1.Start()
	result := map[string]string{}
	for sc := bufio.NewScanner(in); sc.Scan(); {
		text := sc.Text()
		field := strings.Split(text, ": ")
		if len(field) >= 2 {
			result[field[0]] = field[1]
		}
	}
	return result, nil
}

var useVs2010 = flag.Bool("2010", false, "use Visual Studio 2010")
var useVs2013 = flag.Bool("2013", false, "use Visual Studio 2013")
var useVs2015 = flag.Bool("2015", false, "use Visual Studio 2015")
var buildDebug = flag.Bool("d", false, "build debug")
var buildAll = flag.Bool("a", false, "build all(debug and release)")

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
	if *buildAll {
		if err := devenv.Run(sln, "/rebuild", "Debug"); err != nil {
			return err
		}
		return devenv.Run(sln, "/rebuild", "Release")
	} else if *buildDebug {
		return devenv.Run(sln, "/rebuild", "Debug")
	} else {
		return devenv.Run(sln, "/rebuild", "Release")
	}
	return nil
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
