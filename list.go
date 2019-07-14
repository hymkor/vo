package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zetamatta/vf1s/peinfo"
)

const dotNetDLLType = "Library"

type OutDirT struct {
	Condition string `xml:"Condition,attr"`
	Text      string `xml:",chardata"`
}

type CppPropertyGroup struct {
	Condition         string    `xml:"Condition,attr"`
	RootNamespace     string    `xml:"RootNamespace"`
	ConfigurationType string    `xml:"ConfigurationType"`
	OutDir            []OutDirT `xml:"OutDir"`
	TargetExt         string    `xml:"TargetExt"`
}

type NativeProj struct {
	PropertyGroup []CppPropertyGroup `xml:"PropertyGroup"`
}

const nativeDLLType = "DynamicLibrary"

var rxCondition = regexp.MustCompile(`^\s*'([^']*)'\s*==\s*'([^']*)'`)

func cond2replacer(cond string) *strings.Replacer {
	m := rxCondition.FindStringSubmatch(cond)
	if m == nil {
		return nil
	}
	table := make([]string, 0, 4)
	left := strings.Split(m[1], "|")
	right := strings.Split(m[2], "|")
	for i, s := range left {
		table = append(table, s)
		table = append(table, right[i])
	}
	return strings.NewReplacer(table...)
}

func listupProduct(sln *Solution) ([]string, error) {
	result := []string{}
	for _projPath := range sln.Project {
		projPath := filepath.Join(filepath.Dir(sln.Path), _projPath)
		basedir := filepath.Dir(projPath)

		bin, err := ioutil.ReadFile(projPath)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(_projPath, ".vcxproj") {
			vcp := NativeProj{}
			err = xml.Unmarshal(bin, &vcp)
			if err != nil {
				return nil, err
			}

			rootNameSpace := filepath.Base(_projPath)
			rootNameSpace = rootNameSpace[:len(rootNameSpace)-len(filepath.Ext(rootNameSpace))]

			for _, p := range vcp.PropertyGroup {
				if p.RootNamespace != "" {
					rootNameSpace = p.RootNamespace
				}
				for _, outDir := range p.OutDir {
					outputPath := outDir.Text
					if rep := cond2replacer(outDir.Condition); rep != nil {
						outputPath = rep.Replace(outputPath)
					} else if rep := cond2replacer(p.Condition); rep != nil {
						outputPath = rep.Replace(outputPath)
					}

					var suffix string
					if p.TargetExt != "" {
						suffix = p.TargetExt
					} else if p.ConfigurationType == nativeDLLType {
						suffix = ".dll"
					} else {
						suffix = ".exe"
					}
					result = append(result, filepath.Join(basedir, outputPath, rootNameSpace+suffix))
				}
			}
		} else if strings.HasSuffix(_projPath, ".vbproj") ||
			strings.HasSuffix(_projPath, ".csproj") {

			for _, configuration := range sln.Configuration {
				piece := strings.Split(configuration, "|")
				props := Properties{
					"Configuration": strings.ReplaceAll(strings.TrimSpace(piece[0]), " ", ""),
					"Platform":      strings.ReplaceAll(strings.TrimSpace(piece[1]), " ", ""),
				}
				props.LoadProject(projPath)

				filename := props["AssemblyName"]
				if ext, ok := props["TargetExt"]; ok {
					filename += ext
				} else if props["OutputType"] == dotNetDLLType {
					filename += ".dll"
				} else {
					filename += ".exe"
				}
				target := filepath.Join(basedir, props["OutputPath"], filename)
				result = append(result, target)
			}
		} else {
			continue
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
