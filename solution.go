package main

import (
	"bufio"
	"errors"
	"fmt"
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

type Solution struct {
	Path          string
	Version       string
	MinVersion    string
	Configuration []string
	Project       map[string]string
}

var internalVersionToProductVersion = map[string]string{
	"8.0":  "2005",
	"9.0":  "2008",
	"10.0": "2010",
	"11.0": "2012",
	"12.0": "2013",
	"14.0": "2015",
	"15.0": "2017",
}

var rxDefaultVersion = regexp.MustCompile(`^VisualStudioVersion\s*=\s*(\d+\.\d+)`)
var rxMinimumVersion = regexp.MustCompile(`^MinimumVisualStudioVersion\s*=\s*(\d+\.\d+)`)

var rxProjectList = regexp.MustCompile(
	`^Project\([^)]+\)` +
		`\s*=\s*` +
		`"[^"]+"\s*,\s*` +
		`"([^"]+)"\s*,\s*` +
		`"([^"]+)"`)

func NewSolution(fname string) (*Solution, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	sln := &Solution{
		Path:    fname,
		Project: make(map[string]string),
	}

	var block func(string, []string)
	block = func(line string, f []string) {
		if f[0] == "#" && f[1] == "Visual" && f[2] == "Studio" && len(f) >= 4 && sln.Version == "" {
			sln.Version = f[3]
			// println("CommentVersion:", sln.Version)
		} else if m := rxDefaultVersion.FindStringSubmatch(line); m != nil {
			sln.Version = internalVersionToProductVersion[m[1]]
			// println("DefaultVersion:", sln.Version)
		} else if m := rxMinimumVersion.FindStringSubmatch(line); m != nil {
			sln.MinVersion = internalVersionToProductVersion[m[1]]
			// println("MinumumVersion:", sln.MinVersion)
		} else if m := rxProjectList.FindStringSubmatch(line); m != nil {
			//println("Found: ", m[1], " ", m[2])
			sln.Project[m[1]] = m[2]
			save := block
			block = func(line string, f []string) {
				if f[0] == "EndProject" {
					block = save
				}
			}
		} else if f[0] == "GlobalSection(SolutionConfigurationPlatforms)" {
			save := block
			block = func(line string, f []string) {
				if f[0] == "EndGlobalSection" {
					block = save
				} else {
					piece := strings.Split(line, "=")
					sln.Configuration = append(sln.Configuration,
						strings.TrimSpace(piece[0]))
				}
			}
		}
	}

	sc := bufio.NewScanner(fd)
	for sc.Scan() {
		text := sc.Text()
		block(text, strings.Fields(text))
	}
	return sln, nil
}
