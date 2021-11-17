package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	_ "github.com/mattn/getwild"
	"github.com/urfave/cli/v2"

	"github.com/zetamatta/vo/solution"
	"github.com/zetamatta/vo/vswhere"
)

func run(dryrun bool, devenvPath string, param ...string) error {
	cmd1 := exec.Command(devenvPath, param...)
	cmd1.Stdin = os.Stdin
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	fmt.Printf("\"%s\" \"%s\"\n", devenvPath, strings.Join(param, "\" \""))
	if dryrun {
		return nil
	}
	return cmd1.Run()
}

type TargetSolution struct {
	*solution.Solution
	DevenvPath string
}

func seekSolutions(flags *vswhere.Flag, args []string, verbose io.Writer, mustHaveDevenv bool) ([]*TargetSolution, error) {
	slnPaths, err := solution.Find(args)
	if err != nil {
		return nil, err
	}
	targets := make([]*TargetSolution, 0, len(slnPaths))
	for _, slnPath := range slnPaths {
		sln, err := solution.New(slnPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", slnPath, err)
		}

		devenvPath, err := flags.SeekDevenv(sln, verbose)
		if err != nil && mustHaveDevenv {
			return nil, fmt.Errorf("%s: devenv.com not found", slnPath)
		}

		targets = append(targets, &TargetSolution{
			DevenvPath: devenvPath,
			Solution:   sln,
		})
	}
	return targets, nil
}

func seekOneSolution(flags *vswhere.Flag, args []string, verbose io.Writer) (*TargetSolution, error) {
	slns, err := seekSolutions(flags, args, verbose, true)
	if err != nil {
		return nil, err
	}
	if len(slns) <= 0 {
		return nil, errors.New("no solution files")
	}
	if len(slns) >= 2 {
		var buffer strings.Builder
		for _, s := range slns {
			if buffer.Len() > 0 {
				buffer.WriteByte(' ')
			}
			buffer.WriteString(s.Path)
		}
		return nil, fmt.Errorf("%s: too many solution files", buffer.String())
	}
	return slns[0], nil
}

func seekConfig(c *cli.Context, sln *solution.Solution) []string {
	if conf := c.String("config"); conf != "" {
		return []string{conf}
	}

	var filter func(string) bool
	if c.Bool("a") {
		filter = func(c string) bool { return true }
	} else if c.Bool("d") {
		filter = func(c string) bool { return strings.Contains(c, "debug") }
	} else if c.Bool("r") {
		filter = func(c string) bool { return strings.Contains(c, "release") }
	} else {
		return []string{}
	}

	confs := []string{}
	for _, conf := range sln.Configuration {
		if filter(strings.ToLower(conf)) {
			confs = append(confs, conf)
		}
	}
	return confs
}

func getWarningOut(c *cli.Context) io.Writer {
	w := ioutil.Discard
	if c.Bool("w") || globalFlagWarning {
		w = os.Stderr
	}
	return w
}

func getVerboseOut(c *cli.Context) io.Writer {
	v := ioutil.Discard
	if c.Bool("v") || globalFlagVerbose {
		v = os.Stderr
	}
	return v
}

func context2flag(c *cli.Context) *vswhere.Flag {
	return &vswhere.Flag{
		V2019:      c.Bool("2019") || globalFlag2019,
		V2017:      c.Bool("2017") || globalFlag2017,
		V2015:      c.Bool("2015") || globalFlag2015,
		V2013:      c.Bool("2013") || globalFlag2013,
		V2010:      c.Bool("2010") || globalFlag2010,
		SearchDesc: c.Bool("latest") || globalFlagLatest,
	}
}

func build(c *cli.Context, action string) error {
	sln, err := seekOneSolution(context2flag(c), c.Args().Slice(), getVerboseOut(c))
	if err != nil {
		return err
	}
	confs := seekConfig(c, sln.Solution)
	if len(confs) <= 0 {
		return run(c.Bool("n"), sln.DevenvPath, sln.Path, action)
	}
	for _, conf1 := range confs {
		err = run(c.Bool("n"), sln.DevenvPath, sln.Path, action, conf1)
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	globalFlag2010    = false
	globalFlag2013    = false
	globalFlag2015    = false
	globalFlag2017    = false
	globalFlag2019    = false
	globalFlagLatest  = false
	globalFlagWarning = false
	globalFlagVerbose = false
)

func mains() error {
	globalFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:        "2010",
			Usage:       "use Visual Studio 2010",
			Destination: &globalFlag2010,
		},
		&cli.BoolFlag{
			Name:        "2013",
			Usage:       "use Visual Studio 2013",
			Destination: &globalFlag2013,
		},
		&cli.BoolFlag{
			Name:        "2015",
			Usage:       "use Visual Studio 2015",
			Destination: &globalFlag2015,
		},
		&cli.BoolFlag{
			Name:        "2017",
			Usage:       "use Visual Studio 2017",
			Destination: &globalFlag2017,
		},
		&cli.BoolFlag{
			Name:        "2019",
			Usage:       "use Visual Studio 2019",
			Destination: &globalFlag2019,
		},
		&cli.BoolFlag{
			Name:        "latest",
			Usage:       "search Visual Studio order by the version descending",
			Destination: &globalFlagLatest,
		},
		&cli.BoolFlag{
			Name:        "w",
			Usage:       "show warnings",
			Destination: &globalFlagWarning,
		},
		&cli.BoolFlag{
			Name:        "v",
			Usage:       "verbose",
			Destination: &globalFlagVerbose,
		},
	}

	buildOptions := []cli.Flag{
		&cli.BoolFlag{
			Name:  "n",
			Usage: "dry run",
		},
		&cli.BoolFlag{
			Name:  "d",
			Usage: "build configurations contains /Debug/",
		},
		&cli.BoolFlag{
			Name:  "r",
			Usage: "build configurations contains /Release/",
		},
		&cli.BoolFlag{
			Name:  "a",
			Usage: "build all configurations",
		},
		&cli.BoolFlag{
			Name:  "re",
			Usage: "rebuild",
		},
		&cli.StringFlag{
			Name:  "c",
			Usage: "specify the configuraion to build",
		},
	}

	for _, f := range globalFlags {
		if bf, ok := f.(*cli.BoolFlag); ok {
			buildOptions = append(buildOptions, &cli.BoolFlag{
				Name:  bf.Name,
				Usage: bf.Usage,
			})
		}
	}

	app := &cli.App{
		Usage: "Visual studio solution commandline Operator",
		Flags: globalFlags,
		Commands: []*cli.Command{
			{
				Name:  "ide",
				Usage: "start visual-studio associated the solution with no options",
				Action: func(c *cli.Context) error {
					sln, err := seekOneSolution(context2flag(c), c.Args().Slice(), getVerboseOut(c))
					if err != nil {
						return err
					}
					err = run(c.Bool("n"), sln.DevenvPath, sln.Path)
					if err != nil {
						return fmt.Errorf("%s: %w", sln.Path, err)
					}
					return nil
				},
			},
			{
				Name:  "build",
				Usage: "call devenv.com associated the solution with /build option",
				Flags: buildOptions,
				Action: func(c *cli.Context) error {
					return build(c, "/build")
				},
			},
			{
				Name:  "rebuild",
				Usage: "call devenv.com associated the solution with /rebuild option",
				Flags: buildOptions,
				Action: func(c *cli.Context) error {
					return build(c, "/rebuild")
				},
			},
			{
				Name:  "ls",
				Usage: "list up expected executables inline",
				Action: func(c *cli.Context) error {
					slns, err := seekSolutions(context2flag(c), c.Args().Slice(), getVerboseOut(c), false)
					if err != nil {
						return err
					}
					for i, sln := range slns {
						err = listProductInline(sln.Solution, sln.DevenvPath, getWarningOut(c))
						if err != nil {
							fmt.Fprintf(os.Stderr, "%s: %w", sln.Path, err)
							continue
						}
						if i == len(slns)-1 {
							fmt.Println()
						} else {
							fmt.Print(" ")
						}
					}
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "list up existing executables and thier version-information with long format",
				Action: func(c *cli.Context) error {
					slns, err := seekSolutions(context2flag(c), c.Args().Slice(), getVerboseOut(c), false)
					if err != nil {
						return err
					}
					uniq := make(map[string]struct{})
					for _, sln := range slns {
						err := listProductLong(uniq, sln.Solution, sln.DevenvPath, getWarningOut(c))
						if err != nil {
							return fmt.Errorf("%s: %w", sln.Path, err)
						}
					}
					return nil
				},
			},
			{
				Name:  "showver",
				Usage: "Show the version information for executables given by parameters",
				Action: func(c *cli.Context) error {
					for _, s := range c.Args().Slice() {
						showVer(s, os.Stdout)
					}
					return nil
				},
			},
			{
				Name:  "eval",
				Usage: "eval the equation given by parameter",
				Action: func(c *cli.Context) error {
					sln, err := seekOneSolution(context2flag(c), c.Args().Slice(), getVerboseOut(c))
					if err != nil {
						return err
					}
					for _, s := range c.Args().Slice() {
						if !strings.HasSuffix(s, ".sln") {
							if err := eval(sln.Solution, sln.DevenvPath, s); err != nil {
								return fmt.Errorf("%s: %w", sln.Path, err)
							}
						}
					}
					return nil
				},
			},
		},
	}
	return app.Run(os.Args)
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
