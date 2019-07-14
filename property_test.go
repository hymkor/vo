package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestRead1st(t *testing.T) {
	a, err := read1st(strings.NewReader(" a"))
	if a != 'a' || err != nil {
		t.Fatal()
		return
	}
	fmt.Printf("OK:a=%c\n", a)
	b, err := read1st(strings.NewReader("b"))
	if b != 'b' || err != nil {
		t.Fatal()
		return
	}
	fmt.Printf("OK:b=%c\n", b)
	_, err = read1st(strings.NewReader("  "))
	if err == nil {
		t.Fatal()
		return
	}
	fmt.Printf("OK:%v\n", err)
}

func TestEvalString(t *testing.T) {
	s, err := evalString(strings.NewReader(" 'abcdef'"))
	if err != nil {
		t.Fatal()
		return
	}
	if s != "abcdef" {
		t.Fatalf("value=[%s]", s)
		return
	}
	s, err = evalString(strings.NewReader(" 'abcdef"))
	if err == nil {
		t.Fatal()
		return
	}
	fmt.Printf("OK")
}

func TestEvalPropertiesition(t *testing.T) {
	s, err := evalEquation(strings.NewReader(" '123' == '123'"))
	if err != nil {
		t.Fatal()
	}
	if !s {
		t.Fatal()
	}
	s, err = evalEquation(strings.NewReader(" '123' == '124'"))
	if err != nil {
		t.Fatal()
	}
	if s {
		t.Fatal()
	}
}

func TestPropertiesition(t *testing.T) {
	properties := Properties(map[string]string{
		"Platform":      "x86",
		"Configuration": "Debug",
	})

	status, err := properties.EvalText("'$(Platform)' == 'x86'")
	if err != nil {
		t.Fatal()
		return
	}
	if !status {
		t.Fatal()
		return
	}

	status, err = properties.EvalText("'$(Platform)' == 'Win32'")
	if err != nil {
		t.Fatal()
		return
	}
	if status {
		t.Fatal()
		return
	}
}
