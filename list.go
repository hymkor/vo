package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type NetProj struct {
	AssemblyName string   `xml:"PropertyGroup>AssemblyName"`
	OutputType   string   `xml:"PropertyGroup>OutputType"`
	OutputPath   []string `xml:"PropertyGroup>OutputPath"`
}

var type2suffx = map[string]string{
	"WinExe":  ".exe",
	"Library": ".dll",
	"Exe":     ".exe",
}

func listProduct(sln *Solution) error {
	for _projPath := range sln.Project {
		if strings.HasSuffix(_projPath, ".vbproj") ||
			strings.HasSuffix(_projPath, ".csproj") {
			projPath := filepath.Join(filepath.Dir(sln.Path), _projPath)
			bin, err := ioutil.ReadFile(projPath)
			if err != nil {
				return err
			}
			vbp := NetProj{}
			err = xml.Unmarshal(bin, &vbp)
			if err != nil {
				return err
			}
			basedir := filepath.Dir(projPath)
			filename := vbp.AssemblyName + type2suffx[vbp.OutputType]
			for _, dir := range vbp.OutputPath {
				fmt.Println(filepath.Join(basedir, dir, filename))
			}
		}
	}
	return nil
}
