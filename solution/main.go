package solution

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

func Find(args []string) ([]string, error) {
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

type Solution struct {
	Path           string
	MinimumVersion string
	DefaultVersion string
	CommentVersion string
	Configuration  []string
	Project        map[string]string
}

func (s *Solution) GetMinimumVersion() string {
	if s.MinimumVersion != "" {
		return s.MinimumVersion
	}
	if s.DefaultVersion < s.CommentVersion {
		return s.DefaultVersion
	}
	return s.CommentVersion
}

func (s *Solution) GetVersion() string {
	if s.DefaultVersion != "" {
		return s.DefaultVersion
	}
	return s.CommentVersion
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

var rxCommentVersion = regexp.MustCompile(`^#\s*Visual\s+Studio\s+(\d+)`)

var rxDefaultVersion = regexp.MustCompile(`^VisualStudioVersion\s*=\s*(\d+\.\d+)`)
var rxMinimumVersion = regexp.MustCompile(`^MinimumVisualStudioVersion\s*=\s*(\d+\.\d+)`)

var rxProjectList = regexp.MustCompile(
	`^Project\([^)]+\)` +
		`\s*=\s*` +
		`"[^"]+"\s*,\s*` +
		`"([^"]+)"\s*,\s*` +
		`"([^"]+)"`)

func New(fname string) (*Solution, error) {
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
		if m := rxCommentVersion.FindStringSubmatch(line); m != nil {
			sln.CommentVersion = m[1]
		} else if m := rxDefaultVersion.FindStringSubmatch(line); m != nil {
			sln.DefaultVersion = internalVersionToProductVersion[m[1]]
		} else if m := rxMinimumVersion.FindStringSubmatch(line); m != nil {
			sln.MinimumVersion = internalVersionToProductVersion[m[1]]
		} else if m := rxProjectList.FindStringSubmatch(line); m != nil {
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
