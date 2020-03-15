package main

import (
	"fmt"
	"io/ioutil"

	"github.com/zetamatta/vo/solution"
)

func eval(sln *solution.Solution, devenvPath, varname string) error {
	projToConfigToProps, err := getProjToConfigToProps(sln, devenvPath, ioutil.Discard)
	if err != nil {
		return err
	}
	for proj, configToProps := range projToConfigToProps {
		fmt.Printf("%s: \n", proj)
		for config, Props := range configToProps {
			fmt.Printf("  %s: %s\n", config, Props[varname])
		}
	}
	return nil
}
