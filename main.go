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
	SolutionPath string
	DevenvPath   string
}

func seekSolutions(flags *vswhere.Flag, args []string) ([]*TargetSolution, error) {
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

		devenvPath, err := flags.SeekDevenv(sln, os.Stdout)
		if err != nil {
			return nil, fmt.Errorf("%s: devenv.com not found", slnPath)
		}

		targets = append(targets, &TargetSolution{
			SolutionPath: slnPath,
			DevenvPath:   devenvPath,
			Solution:     sln,
		})
	}
	return targets, nil
}

func seekOneSolution(flags *vswhere.Flag, args []string) (*TargetSolution, error) {
	slns, err := seekSolutions(flags, args)
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
			buffer.WriteString(s.SolutionPath)
		}
		return nil, fmt.Errorf("%s: too many solution files", buffer.String())
	}
	return slns[0], nil
}

func seekConfig(c *cli.Context, sln *solution.Solution) string {
	if conf := c.String("config"); conf != "" {
		return conf
	}

	var filter func(string) bool
	if c.Bool("a") {
		filter = func(c string) bool { return true }
	} else if c.Bool("d") {
		filter = func(c string) bool { return strings.Contains(c, "debug") }
	} else if c.Bool("r") {
		filter = func(c string) bool { return strings.Contains(c, "release") }
	} else {
		return ""
	}

	for _, conf := range sln.Configuration {
		if filter(strings.ToLower(conf)) {
			return conf
		}
	}
	return ""
}

func warning(c *cli.Context) io.Writer {
	w := ioutil.Discard
	if c.Bool("w") {
		w = os.Stderr
	}
	return w
}

func verbose(c *cli.Context) io.Writer {
	v := ioutil.Discard
	if c.Bool("v") {
		v = os.Stderr
	}
	return v
}

func context2flag(c *cli.Context) *vswhere.Flag {
	return &vswhere.Flag{
		V2017: c.Bool("2017"),
		V2019: c.Bool("2019"),
		V2015: c.Bool("2015"),
		V2013: c.Bool("2013"),
		V2010: c.Bool("2010"),
	}
}

func mains() error {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "2010",
				Usage: "use Visual Studio 2010",
			},
			&cli.BoolFlag{
				Name:  "2013",
				Usage: "use Visual Studio 2013",
			},
			&cli.BoolFlag{
				Name:  "2015",
				Usage: "use Visual Studio 2015",
			},
			&cli.BoolFlag{
				Name:  "2017",
				Usage: "use Visual Studio 2017",
			},
			&cli.BoolFlag{
				Name:  "2019",
				Usage: "use Visual Studio 2019",
			},
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
			&cli.BoolFlag{
				Name:  "w",
				Usage: "show warnings",
			},
			&cli.BoolFlag{
				Name:  "v",
				Usage: "verbose",
			},
			&cli.StringFlag{
				Name:  "c",
				Usage: "specify the configuraion to build",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "showver",
				Action: func(c *cli.Context) error {
					for _, s := range c.Args().Slice() {
						showVer(s, os.Stdout)
					}
					return nil
				},
			},
			{
				Name: "eval",
				Action: func(c *cli.Context) error {
					sln, err := seekOneSolution(context2flag(c), c.Args().Slice())
					if err != nil {
						return err
					}
					for _, s := range c.Args().Slice() {
						if !strings.HasSuffix(s, ".sln") {
							if err := eval(sln.Solution, sln.DevenvPath, s); err != nil {
								return fmt.Errorf("%s: %w", sln.SolutionPath, err)
							}
						}
					}
					return nil
				},
			},
			{
				Name: "ls",
				Action: func(c *cli.Context) error {
					slns, err := seekSolutions(context2flag(c), c.Args().Slice())
					if err != nil {
						return err
					}
					for i, sln := range slns {
						err = listProductInline(sln.Solution, sln.DevenvPath, os.Stdout)
						if err != nil {
							return fmt.Errorf("%s: %w", sln.SolutionPath, err)
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
				Name: "list",
				Action: func(c *cli.Context) error {
					slns, err := seekSolutions(context2flag(c), c.Args().Slice())
					if err != nil {
						return err
					}
					for _, sln := range slns {
						err := listProductLong(sln.Solution, sln.DevenvPath, os.Stdout)
						if err != nil {
							return fmt.Errorf("%s: %w", sln.SolutionPath, err)
						}
					}
					return nil
				},
			},
			{
				Name: "ide",
				Action: func(c *cli.Context) error {
					sln, err := seekOneSolution(context2flag(c), c.Args().Slice())
					if err != nil {
						return err
					}
					err = run(c.Bool("n"), sln.DevenvPath, sln.SolutionPath)
					if err != nil {
						return fmt.Errorf("%s: %w", sln.SolutionPath, err)
					}
					return nil
				},
			},
			{
				Name: "build",
				Action: func(c *cli.Context) error {
					sln, err := seekOneSolution(context2flag(c), c.Args().Slice())
					if err != nil {
						return err
					}
					conf := seekConfig(c, sln.Solution)
					if conf == "" {
						return nil
					}
					return run(c.Bool("n"), sln.DevenvPath, sln.SolutionPath, "build", conf)
				},
			},
			{
				Name: "rebuild",
				Action: func(c *cli.Context) error {
					sln, err := seekOneSolution(context2flag(c), c.Args().Slice())
					if err != nil {
						return err
					}
					conf := seekConfig(c, sln.Solution)
					if conf == "" {
						return nil
					}
					return run(c.Bool("n"), sln.DevenvPath, sln.SolutionPath, "rebuild", conf)
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
