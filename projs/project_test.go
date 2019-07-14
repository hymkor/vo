package projs

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
	properties := Properties(map[string]string{
		"Platform": "Win32",
	})
	err := properties.ReadProject(strings.NewReader(xml))
	if err != nil {
		t.Fatal()
	}
	if properties["Hoge"] != "Win32Hoge" {
		t.Fatal()
	}
}
