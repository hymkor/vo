package projs

import (
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"
)

func read1st(sc io.RuneScanner) (rune, error) {
	for {
		r, _, err := sc.ReadRune()
		if err != nil {
			return r, err
		}
		if !unicode.IsSpace(r) {
			return r, nil
		}
	}
}

func evalString(sc io.RuneScanner) (string, error) {
	r, err := read1st(sc)
	if err != nil {
		return "", err
	}
	if r != '\'' {
		return "", errors.New("not string literal")
	}
	var buffer strings.Builder
	for {
		r, _, err = sc.ReadRune()
		if err != nil {
			return "", err
		}
		if r == '\'' {
			return buffer.String(), nil
		}
		buffer.WriteRune(r)
	}
}

func evalEquation(sc io.RuneScanner) (bool, error) {
	first, err := evalString(sc)
	if err != nil {
		return false, err
	}
	r, err := read1st(sc)
	if err != nil {
		return false, err
	}
	var op bool
	if r == '=' {
		op = true
	} else if r == '!' {
		op = false
	} else {
		return false, errors.New("1st equal-mark not found")
	}
	r, _, err = sc.ReadRune()
	if err != nil {
		return false, err
	}
	if r != '=' {
		return false, errors.New("2nd equal-mark not found")
	}
	second, err := evalString(sc)
	if err != nil {
		return false, err
	}
	if op {
		return first == second, nil
	} else {
		return first != second, nil
	}
}

var rxExists = regexp.MustCompile(`^\s*[eE]xists\((.*)\)\s*$`)

func EvalCondition(s string) (bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return true, nil
	}
	if m := rxExists.FindStringSubmatch(s); m != nil {
		s, err := evalString(strings.NewReader(m[1]))
		if err != nil {
			return false, err
		}
		if fd, err := os.Open(s); err == nil {
			fd.Close()
			return true, nil
		} else {
			return false, nil
		}
	}
	return evalEquation(strings.NewReader(s))
}

var rxEnvPattern = regexp.MustCompile(`\$\([^\)]+\)`)

type Properties map[string]string

// Expand replaces $(var) to the value of the property.
func (properties Properties) Expand(text string, onNotFound func(string) string) string {
	return rxEnvPattern.ReplaceAllStringFunc(text,
		func(s string) string {
			name := s[2 : len(s)-1]
			if s, ok := properties[name]; ok {
				return s
			} else if onNotFound != nil {
				return onNotFound(name)
			} else {
				return ""
			}
		})
}

// EvalCondition expands $(var) of text and evalute it as an equation.
func (properties Properties) EvalCondition(text string) (bool, error) {
	rc, err := EvalCondition(properties.Expand(text, nil))
	if trace {
		println("EvalText:", text, rc)
	}
	return rc, err
}
