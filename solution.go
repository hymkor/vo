package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

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

var visualStudioPattern = regexp.MustCompile(`^#\s*Visual\s*Studio\s*(2\d{3})\s*$`)

func findVersionInSolution(fname string) (string, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer fd.Close()
	sc := bufio.NewScanner(fd)
	for sc.Scan() {
		text := sc.Text()
		m := visualStudioPattern.FindStringSubmatch(text)
		if m != nil {
			return m[1], nil
		}
	}
	return "", io.EOF
}
