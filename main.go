package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	_ "github.com/mattn/getwild"

	"github.com/zetamatta/vo/solution"
	"github.com/zetamatta/vo/vswhere"
)

var flag2010 = flag.Bool("2010", false, "use Visual Studio 2010")
var flag2013 = flag.Bool("2013", false, "use Visual Studio 2013")
var flag2015 = flag.Bool("2015", false, "use Visual Studio 2015")
var flag2017 = flag.Bool("2017", false, "use Visual Studio 2017")
var flag2019 = flag.Bool("2019", false, "use Visual Studio 2019")

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

		devenvPath, err := vswhere.Flag{
			V2010: *flag2010,
			V2013: *flag2013,
			V2015: *flag2015,
			V2017: *flag2017,
			V2019: *flag2019,
		}.SeekDevenv(sln, verbose)
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
