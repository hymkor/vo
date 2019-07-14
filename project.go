package main

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
					status, err := (properties).EvalText(attr1.Value)
					if err != nil {
						fmt.Fprintf(log, "Condition: '%s' could not parse.(%s)\n",
							attr1.Value, err.Error())
						continue
					}
					if !status {
						decoder.Skip()
						goto next
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
					properties.Replace(strings.TrimSpace(string(se)))
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
	defer fd.Close()
	return properties.ReadProject(fd, log)
}
