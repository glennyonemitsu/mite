package main

import (
	"fmt"
	"unicode/utf8"
)

type Tag struct {
	spaces int
	tag string
}

type Parser struct {
	Scanner Scanner
	output string

	// nested html tags for backtacking
	//tags []string
	// space counts for each indent level for backtacking
	//indents []int

	// heirarchy
	line []Tag

	// for indent calculations, how many spaces should a tab equal
	tabSpace int

	// states as the Parser object figures out how to output the tokens

	// new line has started
	isNewLine bool

	// rest of tokens before TokNewLine is regular text
	isText bool

	// whitespace is significant. This is to distinguish [] syntax, not indent
	isBracket bool

	isIndent bool

	tag Tag
}

func (p *Parser) Output() string {
	var tokens []rune
	var text string

	p.line = []Tag{}
	p.tabSpace = 8
	p.output = ""
	p.isNewLine = true
	p.isText = false
	p.isIndent = true
	p.tag = Tag{spaces:0, tag:""}

	for {
		tokens = p.Scanner.Scan()
		for _, token := range tokens {
			text = p.Scanner.TokenText()
			p.processToken(token, text)
			fmt.Println(TokenString(token))
		}
		if tokens[0] == TokEOF {
			break
		}
	}
	return p.output
}

func (p *Parser) closeLastTag() {
	if len(p.line) > 0 {
		oldTag := p.line[len(p.line) - 1]
		p.line = p.line[0: len(p.line) - 1]
		if r, _ := utf8.DecodeLastRuneInString(p.output); r == ' ' {
			p.output = p.output[0: len(p.output) - 1]
		}
		p.output += "</" + oldTag.tag + ">"
	}
}

func (p *Parser) processToken(tok rune, text string) {
	switch tok {
	case TokWord:
		if p.isNewLine {
			// since there were no indents or dedents, this tag is the new end of the line
			p.closeLastTag()
			p.tag.tag = text
			p.output += "<" + text + ">"
			p.line = append(p.line, p.tag)
			p.isNewLine = false
		} else {
			p.output += text
			p.output += " "
		}
	case TokIndent:
		p.isNewLine = false
	case TokDedent:
		p.isNewLine = false
		// for dedents, pop off the line and close the tag
		p.closeLastTag()
	case TokNewLine:
		p.isNewLine = true
	case TokEOF:
		// close the remaining tags in the line
		for i := len(p.line) - 1; i >= 0; i -= 1 {
			p.output += "</" + p.line[i].tag + ">"
		}
	}
}

