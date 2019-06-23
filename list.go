package main

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

			vbp := NetProj{}
			err = xml.Unmarshal(bin, &vbp)
			if err != nil {
				return nil, err
			}
			filename := vbp.AssemblyName
			if vbp.OutputType == dotNetDLLType {
				filename += ".dll"
			} else {
				filename += ".exe"
			}

			for _, dir := range vbp.OutputPath {
				result = append(result, filepath.Join(basedir, dir, filename))
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

const FILE_ADDRESS_OF_NEW_EXE_HEADER = 60

func bytes2dword(array []byte) int64 {
	return int64(array[0]) +
		int64(array[1])*256 +
		int64(array[2])*256*256 +
		int64(array[3])*256*256*256
}

func GetPeHeaderPos(fd io.ReadSeeker) (int64, error) {
	var array [4]byte

	_, err := fd.Seek(FILE_ADDRESS_OF_NEW_EXE_HEADER, io.SeekStart)
	if err != nil {
		return 0, err
	}
	_, err = fd.Read(array[:])
	if err != nil {
		return 0, err
	}
	return bytes2dword(array[:]), nil
}

func getExeStamp(fd io.ReadSeeker) (time.Time, error) {
	var array [4]byte

	peHeaderPos, err := GetPeHeaderPos(fd)
	if err != nil {
		return time.Time{}, err
	}

	_, err = fd.Seek(peHeaderPos+8, io.SeekStart)
	if err != nil {
		return time.Time{}, err
	}

	_, err = fd.Read(array[:])
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(bytes2dword(array[:]), 0), nil
}

type exeSpec struct {
	Name           string
	Md5Sum         string
	FileVersion    string
	ProductVersion string
	Size           int64
	Stamp          time.Time
}

func getExeSpec(fname string) *exeSpec {
	fd, err := os.Open(fname)
	if err != nil {
		return nil
	}
	defer fd.Close()

	var size int64

	if stat, err := fd.Stat(); err == nil {
		size = stat.Size()
	}

	var fileVer string
	var prodVer string

	if v, err := GetVersionInfo(fname); err == nil {
		if fv, pv, err := v.Number(); err == nil {
			fileVer = fmt.Sprintf("%d.%d.%d.%d", fv[0], fv[1], fv[2], fv[3])
			prodVer = fmt.Sprintf("%d.%d.%d.%d", pv[0], pv[1], pv[2], pv[3])
		}
	}

	h := md5.New()
	if _, err := io.Copy(h, fd); err != nil {
		return nil
	}

	stamp, _ := getExeStamp(fd)

	return &exeSpec{
		Name:           fname,
		Md5Sum:         fmt.Sprintf("%x", h.Sum(nil)),
		FileVersion:    fileVer,
		ProductVersion: prodVer,
		Size:           size,
		Stamp:          stamp,
	}
}

func listProductLong(sln *Solution) error {
	list, err := listupProduct(sln)
	if err != nil {
		return err
	}
	for _, fname := range list {
		fmt.Println(fname)
		if spec := getExeSpec(fname); spec != nil {
			fmt.Printf("\t%-18s%-18s%-18s\n",
				spec.FileVersion,
				spec.ProductVersion,
				spec.Stamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("\t%d bytes md5sum:%s\n", spec.Size, spec.Md5Sum)
		}
	}
	return nil
}
