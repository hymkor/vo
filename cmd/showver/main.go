package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"path/filepath"

	"github.com/hymkor/vo/internal/peinfo"
)

var (
	flagFileVersion = flag.Bool("filever", false, "show the file version")
	flagProdVersion = flag.Bool("prodver", false, "show the product version")
	flagBuildStamp  = flag.Bool("build", false, "show build stamp")
	flagMd5Sum      = flag.Bool("md5", false, "show md5sum")
	flagSize        = flag.Bool("size", false, "show size")
	flag64bit       = flag.Bool("bit", false, "show 64 if 64 bit executable")
	flagOneLinear   = flag.Bool("1", false, "show one line")
)

func globs(patterns []string) []string {
	result := make([]string, 0, len(patterns))
	for _, s := range patterns {
		matches, err := filepath.Glob(s)
		if err != nil {
			result = append(result, s)
		} else {
			result = append(result, matches...)
		}
	}
	return result
}

func mains(args []string) error {
	args = globs(args)
	sep := ""
	for _, fname := range args {
		info := peinfo.New(fname)
		if info == nil {
			fmt.Fprintf(os.Stderr, "%s: not found\n", fname)
			continue
		}

		if *flagFileVersion {
			fmt.Println(info.FileVersion)
		} else if *flagProdVersion {
			fmt.Println(info.ProductVersion)
		} else if *flagBuildStamp {
			fmt.Println(info.Stamp.Format("2006-01-02 15:04:05"))
		} else if *flagMd5Sum {
			fmt.Println(info.Md5Sum)
		} else if *flagSize {
			fmt.Printf("%d\n", info.Size)
		} else if *flag64bit {
			if info.Is64bit {
				fmt.Println("64")
			}
		} else if *flagOneLinear {
			fmt.Printf("%s\t%s\t%s\n",
				fname,
				info.FileVersion,
				info.Md5Sum)
		} else {
			io.WriteString(os.Stdout, sep)
			info.WriteTo(os.Stdout)
			sep = "\n"
		}
	}
	return nil
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
