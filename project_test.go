package main

import (
	"strings"
	"testing"
)

func TestProjectRead(t *testing.T) {
	xml := `
		<Project>
			<PropertyGroup Condition="'$(Platform)'=='x64'">
				<Hoge>x64Hoge</Hoge>
			</PropertyGroup>
			<PropertyGroup Condition="'$(Platform)'=='Win32'">
				<Hoge>Win32Hoge</Hoge>
			</PropertyGroup>
		</Project>
		`
	properties := map[string]string{
		"Platform": "Win32",
	}
	err := projectRead(strings.NewReader(xml), properties)
	if err != nil {
		t.Fatal()
	}
	if properties["Hoge"] != "Win32Hoge" {
		t.Fatal()
	}
}
