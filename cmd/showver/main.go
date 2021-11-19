package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	_ "github.com/mattn/getwild"

	"github.com/zetamatta/vo/internal/peinfo"
)

var (
	flagFileVersion = flag.Bool("filever", false, "show the file version")
	flagProdVersion = flag.Bool("prodver", false, "show the product version")
	flagBuildStamp  = flag.Bool("build", false, "show build stamp")
	flagMd5Sum      = flag.Bool("md5", false, "show md5sum")
	flagSize        = flag.Bool("size", false, "show size")
	flag64bit       = flag.Bool("bit", false, "show 64 if 64 bit executable")
)

func mains(args []string) error {
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
