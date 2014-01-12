package main

import (
	"fmt"
	"strings"
)

var template string

func main() {
	fmt.Println(template)
	tmpl := strings.NewReader(template)
	var s Scanner
	s.Init(tmpl)
	/*
	tok := s.Scan()
	for tok != TokEOF {
		fmt.Println(TokenString(tok))
		fmt.Println(s.TokenText())
		fmt.Println("New Scan")
		tok = s.Scan()
	}
	*/
	p := Parser{}
	p.Scanner = s
	fmt.Println("Parser output:")
	fmt.Println(p.Output())

}

func init() {
	template = `h1 Some Big Header
div
	p Hello, World!`
}

