package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const productPath = "productPath: "

func programFiles86() string {
	if val, ok := os.LookupEnv("ProgramFiles(x86)"); ok {
		return val
	}
	return os.Getenv("ProgramFiles")
}

func vswherePath() (string, error) {
	vswhere := filepath.Join(programFiles86(), `Microsoft Visual Studio\Installer\vswhere.exe`)
	if fd, err := os.Open(vswhere); err == nil {
		fd.Close()
		return vswhere, nil
	} else {
		return "", err
	}
}

func ProductPath(args ...string) (string, error) {
	vswhere, err := vswherePath()
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
