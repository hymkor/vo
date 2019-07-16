package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zetamatta/vf1s/peinfo"
	"github.com/zetamatta/vf1s/projs"
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

func listupProduct(sln *Solution, devenvPath string, warning io.Writer) ([]string, error) {
	result := []string{}

	vcTargetsPath, _ := getVCTargetsPath(devenvPath)
	// println("vctargetpath=", vcTargetsPath)

	for _projPath := range sln.Project {
		projPath := filepath.Join(filepath.Dir(sln.Path), _projPath)
		basedir := filepath.Dir(projPath)

		for _, configuration := range sln.Configuration {
			piece := strings.Split(configuration, "|")
			props := projs.Properties{
				"Configuration": strings.ReplaceAll(strings.TrimSpace(piece[0]), " ", ""),
				"Platform":      strings.ReplaceAll(strings.TrimSpace(piece[1]), " ", ""),
				"VCTargetsPath": vcTargetsPath,
				"ProjectName":   withoutExt(filepath.Base(projPath)),
			}
			err := props.LoadProject(projPath, warning)
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

func listProductInline(sln *Solution, devenvPath string, warning io.Writer) error {
	list, err := listupProduct(sln, devenvPath, warning)
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

func listProductLong(sln *Solution, devenvPath string, warning io.Writer) error {
	list, err := listupProduct(sln, devenvPath, warning)
	if err != nil {
		return err
	}
	for _, fname := range list {
		showVer(fname, os.Stdout)
	}
	return nil
}
