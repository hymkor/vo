package vswhere

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zetamatta/vo/internal/solution"
)

type Flag struct {
	V2017      bool
	V2019      bool
	V2015      bool
	V2013      bool
	V2010      bool
	SearchDesc bool
}

func envToCom(envname string) (string, error) {
	env := os.Getenv(envname)
	if env == "" {
		return "", fmt.Errorf("%%%s%% is not set.", envname)
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
	return envToCom("VS100COMNTOOLS") // [10.0,11.0)
}

func seek2013() (string, error) {
	return envToCom("VS120COMNTOOLS") // [12.0,13.0)
}

func seek2015() (string, error) {
	return envToCom("VS140COMNTOOLS") // [14.0,15.0)
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

var searchListAsc = []string{
	"2010",
	"2013",
	"2015",
	"2017",
	"2019",
}

var searchListDesc = []string{
	"2019",
	"2017",
	"2015",
	"2013",
	"2010",
}

var toolsVersionToRequiredVisualStudio = map[string]string{
	"4.0":  "2010",
	"12.0": "2013",
	"14.0": "2015",
	"15.0": "2017",
}

var platformToolSetToRequiredVisualStudio = map[string]string{
	"v90":     "2008",
	"v100":    "2010",
	"v120":    "2013",
	"v120_xp": "2013",
	"v140":    "2017",
	"v141":    "2017",
	"v141_xp": "2017",
}

// https://docs.microsoft.com/ja-jp/visualstudio/msbuild/msbuild-toolset-toolsversion?view=vs-2019

func (flg Flag) SeekDevenv(sln *solution.Solution, log io.Writer) (compath string, err error) {
	// option to force
	if flg.V2019 {
		compath, err = seek2019()
	}
	if flg.V2017 {
		compath, err = seek2017()
	}
	if flg.V2015 {
		compath, err = seek2015()
	}
	if flg.V2013 {
		compath, err = seek2013()
	}
	if flg.V2010 {
		compath, err = seek2010()
	}
	if err == nil && compath != "" {
		return
	}
	if err != nil {
		fmt.Fprintln(log, err)
		fmt.Fprintln(log, "look for the other Visual Studio.")
	}

	// see project-files
	toolsVersion, platformToolSet := sln.MaxToolsVersion()
	requiredVisualStudio := toolsVersionToRequiredVisualStudio[toolsVersion]

	if t, ok := platformToolSetToRequiredVisualStudio[platformToolSet]; ok {
		if t > requiredVisualStudio {
			requiredVisualStudio = t
		}
	}

	if v := sln.GetMinimumVersion(); v > requiredVisualStudio {
		requiredVisualStudio = v
	}

	if f := versionToSeekfunc[requiredVisualStudio]; f != nil {
		fmt.Fprintf(log, "%s: comment version: %s\n", sln.Path, sln.CommentVersion)
		fmt.Fprintf(log, "%s: default version: %s\n", sln.Path, sln.DefaultVersion)
		fmt.Fprintf(log, "%s: minimum version: %s\n", sln.Path, sln.MinimumVersion)
		fmt.Fprintf(log, "%s: required ToolsVersion is '%s'.\n", sln.Path, toolsVersion)
		if platformToolSet != "" {
			fmt.Fprintf(log, "Required PlatformToolSet is '%s'.\n", platformToolSet)
		}
		fmt.Fprintf(log, "%s: try to use Visual Studio %s.\n", sln.Path, requiredVisualStudio)
		compath, err = f()
		if compath != "" && err == nil {
			return
		}
		if err != nil {
			fmt.Fprintln(log, err)
			fmt.Fprintln(log, "look for other versions of Visual Studio.")
		}
	}

	var searchList []string
	if flg.SearchDesc {
		searchList = searchListDesc
	} else {
		searchList = searchListAsc
	}
	// latest version
	for _, v := range searchList {
		if requiredVisualStudio != "" && v < requiredVisualStudio {
			continue
		}
		f, ok := versionToSeekfunc[v]
		if !ok {
			continue
		}
		compath, err = f()
		if compath != "" && err == nil {
			fmt.Fprintf(log, "found '%s'\n", compath)
			return
		}
		if err != nil {
			fmt.Fprintln(log, err)
		}
	}

	// Use the version specified by the solution file and ignore project files
	if f := versionToSeekfunc[sln.DefaultVersion]; f != nil {
		fmt.Fprintf(log, "%s: use default version: %s\n", sln.Path, sln.DefaultVersion)
		compath, err = f()
		if compath != "" && err == nil {
			return
		}
	}
	if f := versionToSeekfunc[sln.MinimumVersion]; f != nil {
		fmt.Fprintf(log, "%s: use minimum version: %s\n", sln.Path, sln.MinimumVersion)
		compath, err = f()
		if compath != "" && err == nil {
			return
		}
	}
	return "", io.EOF
}
