package main

import (
	"encoding/xml"
	"io"
	"os"
	"strings"
)

func (properties Properties) ReadProject(r io.Reader) error {
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
			for _, attr1 := range se.Attr {
				if attr1.Name.Local == "Condition" {
					status, err := (properties).EvalText(attr1.Value)
					if err != nil {
						return err
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
				value := strings.TrimSpace(string(se))
				properties[lastElement] = value
			}
			break
		}
	next:
	}
}

func (properties Properties) LoadProject(projname string) error {
	fd, err := os.Open(projname)
	if err != nil {
		return err
	}
	defer fd.Close()
	return properties.ReadProject(fd)
}
