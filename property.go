package main

import (
	"errors"
	"io"
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
		return "", errors.New("not string")
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
	if r != '=' {
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
	return first == second, nil
}

func evalCondition(s string) (bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return true, nil
	}
	return evalEquation(strings.NewReader(s))
}

var rxEnvPattern = regexp.MustCompile(`\$\([^\)]+\)`)

func EvalProperties(properties map[string]string, text string) (bool, error) {
	text = rxEnvPattern.ReplaceAllStringFunc(text,
		func(s string) string {
			return properties[s[2:len(s)-1]]
		})
	return evalCondition(text)
}
