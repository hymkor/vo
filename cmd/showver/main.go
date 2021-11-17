package main

import (
	"fmt"
	"os"
	"io"

	_ "github.com/mattn/getwild"

	"github.com/zetamatta/vo/internal/peinfo"
)

func mains(args []string) error {
	sep := ""
	for _,fname := range args {
		info := peinfo.New(fname)
		if info == nil {
			fmt.Fprintf(os.Stderr,"%s: not found\n",fname)
			continue
		}
		io.WriteString(os.Stdout,sep)
		info.WriteTo(os.Stdout)
		sep = "\n"
	}
	return nil
}

func main(){
	if err := mains(os.Args[1:]) ; err != nil {
		fmt.Fprintln(os.Stderr,err.Error())
		os.Exit(1)
	}
}

