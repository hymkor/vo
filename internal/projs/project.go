package projs

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

const trace = false

func (properties Properties) ReadProject(r io.Reader, log io.Writer) error {
	decoder := xml.NewDecoder(r)
	var lastElement string
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "ProjectConfiguration" {
				target := strings.TrimSpace(properties["Configuration"]) +
					"|" + strings.TrimSpace(properties["Platform"])
				for _, attr1 := range se.Attr {
					if attr1.Name.Local == "Include" && attr1.Value != target {
						decoder.Skip()
						// println("skip for", attr1.Value, "!=", target)
						goto next
					}
				}
			}
			for _, attr1 := range se.Attr {
				if attr1.Name.Local == "Condition" {
					value := properties.Expand(attr1.Value, func(s string) string {
						fmt.Fprintf(log, "Condition: variable $(%s) not found.\n", s)
						return ""
					})
					status, err := EvalCondition(value)
					if err != nil {
						fmt.Fprintf(log, "Condition: `%s` could not parse.(%s)\n",
							value, err.Error())
						continue
					}
					if !status {
						decoder.Skip()
						goto next
					}
				}
			}
			if se.Name.Local == "Import" {
				for _, attr1 := range se.Attr {
					if attr1.Name.Local == "Project" {
						value := properties.Expand(attr1.Value, func(s string) string {
							fmt.Fprintf(log, "Condition: variable $(%s) not found.\n", s)
							return ""
						})
						err := properties.LoadProject(value, log)
						if err != nil {
							fmt.Fprintf(log, "Imports: `%s` could not open.\n", value)
						}
					}
				}
			}
			lastElement = se.Name.Local
			break
		case xml.EndElement:
			lastElement = ""
		case xml.CharData:
			if lastElement != "" {
				properties[lastElement] =
					properties.Expand(strings.TrimSpace(string(se)), func(s string) string {
						fmt.Fprintf(log, "$(%s) not found.\n", s)
						return ""
					})
			}
			break
		}
	next:
	}
}

func (properties Properties) LoadProject(projname string, log io.Writer) error {
	fd, err := os.Open(projname)
	if err != nil {
		return err
	}
	fmt.Fprintf(log, "*** Start to read project `%s` ***\n", projname)
	rc := properties.ReadProject(fd, log)
	fmt.Fprintf(log, "*** End to read project `%s` ***\n", projname)
	fd.Close()
	return rc
}
