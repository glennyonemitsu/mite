package main

import (
	"fmt"
	"unicode/utf8"
)

type Node struct {
	tag string
	attrs map[string]string
	text string
}

type Parser struct {
	Scanner Scanner
	output string

	// node heirarchy
	stack []*Node

	// states as the Parser object figures out how to output the tokens

	// new line has started
	isNewLine bool

	isIndent bool
	isDedent bool
	dedentCount int

	node *Node
}

func (p *Parser) Output() string {
	var tokens []rune
	var text string

	p.stack = make([]*Node, 0)
	p.output = ""
	p.isNewLine = true
	p.isIndent = false
	p.isDedent = false
	p.dedentCount = 0
	p.node = new(Node)

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

func (p *Parser) closeLastNode() {
	if len(p.stack) > 0 {
		oldNode := p.stack[len(p.stack) - 1]
		p.stack = p.stack[0: len(p.stack) - 1]
		if r, l := utf8.DecodeLastRuneInString(p.output); r == ' ' {
			p.output = p.output[0: len(p.output) - l]
		}
		p.output += "</" + oldNode.tag + ">"
	}
}

func (p *Parser) lastNode() *Node {
	return p.stack[len(p.stack) - 1]
}

func (p *Parser) outputNode(node *Node) {
	p.output += "<" + node.tag + ">"
	p.output += node.text
}

func (p *Parser) processToken(tok rune, text string) {
	switch tok {
	case TokWord:
		if p.node.tag == "" {
			p.node.tag = text
		} else {
			p.node.text += text
		}
	case TokWhitespace:
		p.node.text += text
	case TokIndent:
		p.isIndent = true
		p.isDedent = false
	case TokDedent:
		p.isIndent = false
		p.isDedent = true
		p.dedentCount += 1
	case TokNewLine:
		// up to this point a node has been constructed but not appended to the
		// output

		fmt.Println(p.node)
		if p.isDedent {
			// for dedents, pop off the stack and close the node
			for i := 0; i < p.dedentCount; i += 1 {
				p.closeLastNode()
			}
			// closing one more time because this new node replaces the top node in the
			// stack
			p.closeLastNode()
			p.outputNode(p.node)
			p.stack = append(p.stack, p.node)
		} else if p.isIndent {
			p.outputNode(p.node)
			p.stack = append(p.stack, p.node)
		} else {
			p.closeLastNode()
			p.outputNode(p.node)
			p.stack = append(p.stack, p.node)
		}
		p.node = new(Node)
		p.isIndent = false
		p.isDedent = false
	case TokEOF:
		if p.node.tag != "" {
			p.stack = append(p.stack, p.node)
			p.outputNode(p.node)
		}
		// close the remaining nodes in the stack 
		for len(p.stack) > 0 {
			p.closeLastNode()
		}
	}
}

