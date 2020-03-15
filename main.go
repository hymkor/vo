package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "github.com/mattn/getwild"
	"github.com/zetamatta/go-numeric-compare"

	"github.com/zetamatta/vo/solution"
	"github.com/zetamatta/vo/vswhere"
)

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
	return vswhere.ProductPath("-version", "[15.0,16.0)")
}

func seek2019() (string, error) {
	return vswhere.ProductPath("-version", "[16.0,17.0)")
}

func seekLatest() (string, error) {
	return vswhere.ProductPath("-latest")
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

var flag2010 = flag.Bool("2010", false, "use Visual Studio 2010")
var flag2013 = flag.Bool("2013", false, "use Visual Studio 2013")
var flag2015 = flag.Bool("2015", false, "use Visual Studio 2015")
var flag2017 = flag.Bool("2017", false, "use Visual Studio 2017")
var flag2019 = flag.Bool("2019", false, "use Visual Studio 2019")

type xmlProjectT struct {
	XMLName         xml.Name `xml:"Project"`
	ToolsVersion    string   `xml:"ToolsVersion,attr"`
	PlatformToolset []string `xml:"PropertyGroup>PlatformToolset"`
}

func maxToolsVersion(sln *solution.Solution) (toolsVersion, platformToolset string) {
	for projPath := range sln.Project {
		xmlBin, err := ioutil.ReadFile(projPath)
		if err == nil {
			var xmlProject xmlProjectT
			if xml.Unmarshal(xmlBin, &xmlProject) == nil {
				v := xmlProject.ToolsVersion
				if numeric.Compare(toolsVersion, v) <= 0 {
					toolsVersion = v
				}
				for _, v := range xmlProject.PlatformToolset {
					if numeric.Compare(platformToolset, v) < 0 {
						platformToolset = v
					}
				}
			}
		}
	}
	return
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

func seekDevenv(sln *solution.Solution, log io.Writer) (compath string, err error) {
	// option to force
	if *flag2019 {
		compath, err = seek2019()
	}
	if *flag2017 {
		compath, err = seek2017()
	}
	if *flag2015 {
		compath, err = seek2015()
	}
	if *flag2013 {
		compath, err = seek2013()
	}
	if *flag2010 {
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
	toolsVersion, platformToolSet := maxToolsVersion(sln)
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

	// latest version
	for _, f := range searchList {
		compath, err = f()
		if compath != "" && err == nil {
			fmt.Fprintf(log, "found '%s'\n", compath)
			return
		}
		if err != nil {
			fmt.Fprintln(log, err)
		}
	}
	return "", io.EOF
}

func run(devenvPath string, param ...string) error {
	cmd1 := exec.Command(devenvPath, param...)
	cmd1.Stdin = os.Stdin
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	fmt.Printf("\"%s\" \"%s\"\n", devenvPath, strings.Join(param, "\" \""))
	if *flagDryRun {
		return nil
	}
	return cmd1.Run()
}

var (
	flagShowVer           = flag.String("showver", "", "show version")
	flagListProductInline = flag.Bool("ls", false, "list products")
	flagListProductLong   = flag.Bool("ll", false, "list products")
	flagDryRun            = flag.Bool("n", false, "dry run")
	flagDebug             = flag.Bool("d", false, "build configurations contains /Debug/")
	flagRelease           = flag.Bool("r", false, "build configurations contains /Release/")
	flagAll               = flag.Bool("a", false, "build all configurations")
	flagRebuild           = flag.Bool("re", false, "rebuild")
	flagIde               = flag.Bool("i", false, "open ide")
	flagConfig            = flag.String("c", "", "specify the configuraion to build")
	flagWarning           = flag.Bool("w", false, "show warnings")
	flagVerbose           = flag.Bool("v", false, "verbose")
	flagEval              = flag.String("e", "", "eval variable")
)

func _main() error {
	flag.Parse()

	args := flag.Args()

	warning := ioutil.Discard
	if *flagWarning {
		warning = os.Stderr
	}
	verbose := ioutil.Discard
	if *flagVerbose {
		verbose = os.Stderr
	}

	if *flagShowVer != "" {
		showVer(*flagShowVer, os.Stdout)
		return nil
	}

	slnPaths, err := solution.Find(args)
	if err != nil {
		return err
	}
	for slnCount, slnPath := range slnPaths {
		sln, err := solution.New(slnPath)
		if err != nil {
			return fmt.Errorf("%s: %w", slnPath, err)
		}

		devenvPath, err := seekDevenv(sln, verbose)
		if err != nil {
			return fmt.Errorf("%s: devenv.com not found", slnPath)
		}
		if *flagEval != "" {
			if err := eval(sln, devenvPath, *flagEval); err != nil {
				return fmt.Errorf("%s: %w", slnPath, err)
			}
			continue
		}
		if *flagListProductInline {
			if err := listProductInline(sln, devenvPath, warning); err != nil {
				return fmt.Errorf("%s: %w", slnPath, err)
			}
			if slnCount == len(slnPaths)-1 {
				fmt.Println()
			} else {
				fmt.Print(" ")
			}
			continue
		}
		if *flagListProductLong {
			if err := listProductLong(sln, devenvPath, warning); err != nil {
				return fmt.Errorf("%s: %w", slnPath, err)
			}
			continue
		}

		// Below code are executed when only one solution file exists.
		if len(slnPaths) >= 2 {
			return fmt.Errorf("%s: too many solution files", strings.Join(slnPaths, " "))
		}
		if *flagIde {
			if err := run(devenvPath, slnPath); err != nil {
				return fmt.Errorf("%s: %w", slnPath, err)
			}
			continue
		}
		action := "/build"
		if *flagRebuild {
			action = "/rebuild"
		}
		if *flagConfig != "" {
			return run(devenvPath, slnPath, action, *flagConfig)
		}

		var filter func(string) bool
		if *flagAll {
			filter = func(c string) bool { return true }
		} else if *flagDebug {
			filter = func(c string) bool { return strings.Contains(c, "debug") }
		} else if *flagRelease {
			filter = func(c string) bool { return strings.Contains(c, "release") }
		} else {
			flag.PrintDefaults()
			return nil
		}

		for _, conf := range sln.Configuration {
			if filter(strings.ToLower(conf)) {
				if err := run(devenvPath, slnPath, action, conf); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
