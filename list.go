package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zetamatta/vf1s/peinfo"
)

const dotNetDLLType = "Library"
const nativeDLLType = "DynamicLibrary"

var rxCondition = regexp.MustCompile(`^\s*'([^']*)'\s*==\s*'([^']*)'`)

func listupProduct(sln *Solution) ([]string, error) {
	result := []string{}
	for _projPath := range sln.Project {
		projPath := filepath.Join(filepath.Dir(sln.Path), _projPath)
		basedir := filepath.Dir(projPath)

		for _, configuration := range sln.Configuration {
			piece := strings.Split(configuration, "|")
			props := Properties{
				"Configuration": strings.ReplaceAll(strings.TrimSpace(piece[0]), " ", ""),
				"Platform":      strings.ReplaceAll(strings.TrimSpace(piece[1]), " ", ""),
			}
			err := props.LoadProject(projPath, os.Stderr)
			if err != nil {
				return result, err
			}

			outputFile := props["OutputFile"]
			if outputFile == "" {
				filename := props["AssemblyName"]
				if filename == "" {
					filename = filepath.Base(projPath)
					filename = filename[:len(filename)-len(filepath.Ext(filename))]
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
			target := filepath.Join(basedir, outputFile)
			result = append(result, target)
		}
	}
	return result, nil
}

func listProductInline(sln *Solution) error {
	list, err := listupProduct(sln)
	if err != nil {
		return err
	}
	ofs := ""
	for _, s := range list {
		fmt.Printf(`%s"%s"`, ofs, s)
		ofs = "\t"
	}
	fmt.Println()
	return nil
}

func showVer(fname string, w io.Writer) {
	if spec := peinfo.New(fname); spec != nil {
		spec.WriteTo(w)
	} else {
		fmt.Fprintln(w, fname)
	}
}

func listProductLong(sln *Solution) error {
	list, err := listupProduct(sln)
	if err != nil {
		return err
	}
	for _, fname := range list {
		showVer(fname, os.Stdout)
	}
	return nil
}
