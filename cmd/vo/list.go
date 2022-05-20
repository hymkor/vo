package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hymkor/vo/internal/peinfo"
	"github.com/hymkor/vo/internal/projs"
	"github.com/hymkor/vo/internal/solution"
)

const dotNetDLLType = "Library"
const nativeDLLType = "DynamicLibrary"

var rxCondition = regexp.MustCompile(`^\s*'([^']*)'\s*==\s*'([^']*)'`)

func getVCTargetsPath(compath string) (string, error) {
	vcpath := filepath.Join(filepath.Dir(compath), `..\..\MSBuild\Microsoft\VC`)
	files, err := ioutil.ReadDir(vcpath)
	if err != nil {
		return "", err
	}
	result := ""
	for _, file := range files {
		name := file.Name()
		if name[0] != '.' && file.IsDir() && name > result {
			result = file.Name()
		}
	}
	return filepath.Join(vcpath, result), nil
}

func withoutExt(fname string) string {
	return fname[:len(fname)-len(filepath.Ext(fname))]
}

func getProjToConfigToProps(sln *solution.Solution, devenvPath string, warning io.Writer) (map[string]map[string]projs.Properties, error) {
	var vcTargetsPath string
	if devenvPath != "" {
		vcTargetsPath, _ = getVCTargetsPath(devenvPath)
	}

	projToConfigToProps := map[string]map[string]projs.Properties{}

	for _projPath := range sln.Project {
		projPath := filepath.Join(filepath.Dir(sln.Path), _projPath)
		configToProps := map[string]projs.Properties{}
		for _, configuration := range sln.Configuration {
			piece := strings.Split(configuration, "|")
			props := projs.Properties{
				"Configuration": strings.ReplaceAll(strings.TrimSpace(piece[0]), " ", ""),
				"Platform":      strings.ReplaceAll(strings.TrimSpace(piece[1]), " ", ""),
				"VCTargetsPath": vcTargetsPath,
				"ProjectName":   withoutExt(filepath.Base(projPath)),
				"ProjectDir":    filepath.Dir(projPath),
			}
			err := props.LoadProject(projPath, warning)
			if err != nil {
				continue
				// return nil, err
			}
			configToProps[configuration] = props
		}
		if len(configToProps) > 0 {
			projToConfigToProps[_projPath] = configToProps
		}
	}
	return projToConfigToProps, nil
}

func listupProduct(sln *solution.Solution, devenvPath string, warning io.Writer) (map[string]map[string]string, error) {
	projToConfigToProp, err := getProjToConfigToProps(sln, devenvPath, warning)
	if err != nil {
		return nil, err
	}
	projToConfigToProduct := map[string]map[string]string{}
	for proj, configToProps := range projToConfigToProp {
		configToProduct := map[string]string{}
		for config, props := range configToProps {
			outputFile := props["OutputFile"]
			if outputFile == "" {
				filename := props["AssemblyName"]
				if filename == "" {
					filename = props["ProjectName"]
				}
				if ext, ok := props["TargetExt"]; ok {
					filename += ext
				} else if props["OutputType"] == dotNetDLLType {
					filename += ".dll"
				} else if props["ConfigurationType"] == nativeDLLType {
					filename += ".dll"
				} else {
					filename += ".exe"
				}
				outdir := props["OutputPath"]
				if outdir == "" {
					outdir = props["OutDir"]
				}
				outputFile = filepath.Join(outdir, filename)
			}
			target := filepath.Join(props["ProjectDir"], outputFile)
			configToProduct[config] = target
		}
		projToConfigToProduct[proj] = configToProduct
	}
	return projToConfigToProduct, nil
}

func listProductInline(sln *solution.Solution, devenvPath string, warning io.Writer) error {
	projToConfigToProduct, err := listupProduct(sln, devenvPath, warning)
	if err != nil {
		return err
	}
	uniq := make(map[string]struct{})
	ofs := ""
	for _, configToProduct := range projToConfigToProduct {
		for _, s := range configToProduct {
			if _, ok := uniq[s]; !ok {
				if strings.ContainsRune(s, ' ') {
					fmt.Printf(`%s"%s"`, ofs, s)
				} else {
					fmt.Print(ofs, s)
				}
				ofs = " "
				uniq[s] = struct{}{}
			}
		}
	}
	return nil
}

func showVer(fname string, w io.Writer) {
	if spec := peinfo.New(fname); spec != nil {
		spec.WriteTo(w)
	} else {
		fmt.Fprintln(w, fname)
	}
}

func listProductLong(uniq map[string]struct{}, sln *solution.Solution, devenvPath string, warning io.Writer) error {
	projToConfigToProduct, err := listupProduct(sln, devenvPath, warning)
	if err != nil {
		return err
	}
	for proj, configToProduct := range projToConfigToProduct {
		if _, ok := uniq[proj]; ok {
			continue
		}
		uniq[proj] = struct{}{}

		var buffer strings.Builder

		fmt.Fprintf(&buffer, "%s:\n", proj)
		for config, fname := range configToProduct {
			if fd, err := os.Open(fname); err == nil {
				fd.Close()
				fmt.Print(buffer.String())
				buffer.Reset()
				fmt.Printf("  %s:\n    ", config)
				showVer(fname, os.Stdout)
			}
		}
	}
	return nil
}
