package main

import (
	"encoding/xml"
	"io"
	"os"
	"strings"
)

func projectRead(r io.Reader, properties map[string]string) error {
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
					status, err := EvalProperties(properties, attr1.Value)
					if err != nil {
						return err
					}
					if !status {
						decoder.Skip()
						break
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
	}
}

func projectLoad(projname string, properties map[string]string) error {
	fd, err := os.Open(projname)
	if err != nil {
		return err
	}
	defer fd.Close()
	return projectRead(fd, properties)
}
