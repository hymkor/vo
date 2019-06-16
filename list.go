package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

type NetProj struct {
	AssemblyName string   `xml:"PropertyGroup>AssemblyName"`
	OutputType   string   `xml:"PropertyGroup>OutputType"`
	OutputPath   []string `xml:"PropertyGroup>OutputPath"`
}

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

func listProduct(sln *Solution) error {
	for _projPath := range sln.Project {
		projPath := filepath.Join(filepath.Dir(sln.Path), _projPath)
		basedir := filepath.Dir(projPath)

		bin, err := ioutil.ReadFile(projPath)
		if err != nil {
			return err
		}
		if strings.HasSuffix(_projPath, ".vcxproj") {
			vcp := NativeProj{}
			err = xml.Unmarshal(bin, &vcp)
			if err != nil {
				return err
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
					fmt.Println(filepath.Join(basedir,
						outputPath,
						rootNameSpace+suffix))
				}
			}
		} else if strings.HasSuffix(_projPath, ".vbproj") ||
			strings.HasSuffix(_projPath, ".csproj") {

			vbp := NetProj{}
			err = xml.Unmarshal(bin, &vbp)
			if err != nil {
				return err
			}
			filename := vbp.AssemblyName
			if vbp.OutputType == dotNetDLLType {
				filename += ".dll"
			} else {
				filename += ".exe"
			}

			for _, dir := range vbp.OutputPath {
				fmt.Println(filepath.Join(basedir, dir, filename))
			}
		} else {
			continue
		}
	}
	return nil
}
