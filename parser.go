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

	debug bool

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

	// flag indicates we are still checking for attribute assignments
	isAttr bool

	// attrName and attrValue are buffers to determine if TokWord for a tag are 
	// potentially for an attribute, or if they are just normal text
	attrName string
	attrValue string
	attrAssigned bool
	// dumb appended string while attribute assignment is still being determined
	// this is used if it turns out tokens are NOT part of an attribute assignment
	// and the string (with whitespace if any) gets appended to the node text
	attrString string

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
	p.isAttr = false
	p.attrName = ""
	p.attrValue = ""
	p.attrAssigned = false
	p.attrString = ""
	p.dedentCount = 0
	p.node = p.newNode()

	for {
		tokens = p.Scanner.Scan()
		for _, token := range tokens {
			text = p.Scanner.TokenText()
			if p.debug {
				fmt.Printf("[debug] token: [%s]\n", TokenString(token))
				fmt.Printf("[debug] text:  [%s]\n", text)
			}
			p.processToken(token, text)
			if p.debug {
				fmt.Printf("[debug] node:  [%s]\n\n", p.node)
			}
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
	p.output += "<" + node.tag
	if len(node.attrs) > 0 {
		for k, v := range node.attrs {
			p.output += fmt.Sprintf(" %s='%s'", k, v)
		}
	}
	p.output += ">"
	p.output += node.text
}

func (p *Parser) processToken(tok rune, text string) {
	switch tok {
	case TokWord, TokComma:
		if p.node.tag == "" {
			p.node.tag = text
			// found the tag, now check for attributes
			p.isAttr = true
		} else if p.isAttr {
			if p.attrAssigned {
				if _, found := p.node.attrs[p.attrName]; found {
					p.node.attrs[p.attrName] += " "
					p.node.attrs[p.attrName] += text
				} else {
					p.node.attrs[p.attrName] = text
				}

				// reset to look for a new attribute assignment
				p.attrName = ""
				p.attrValue = ""
				p.attrAssigned = false
			} else if p.attrName == "" {
				p.attrString += text
				p.attrName = text
			} else {
				p.node.text += p.attrString
				p.node.text += text
				p.isAttr = false
				p.attrName = ""
				p.attrValue = ""
				p.attrAssigned = false
			}
		} else {
			p.node.text += text
		}
	case TokAssign:
		if p.isAttr {
			if p.attrName != "" && p.attrValue == "" {
				p.attrAssigned = true
				p.attrString += text
			}
			if p.attrName == "" || p.attrValue != "" {
				p.attrAssigned = false
				p.isAttr = false
				p.node.text += p.attrString
			}
		} else {
			p.node.text += text
		}
	case TokWhitespace:
		if p.isAttr && p.attrString != "" {
			p.attrString += text

		// skip initial whitespace for text
		} else if p.node.text != "" {
			p.node.text += text
		}
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
		if p.isDedent {
			// for dedents, pop off the stack and close the node
			for p.dedentCount > 0 {
				p.closeLastNode()
				p.dedentCount -= 1
			}
		}
		if p.isDedent || !p.isIndent {
			// closing one more time because this new node replaces the top node in the
			// stack for dedents or siblings (non indent)
			p.closeLastNode()
		}
		p.outputNode(p.node)
		p.stack = append(p.stack, p.node)

		// reset
		p.node = p.newNode()
		p.isIndent = false
		p.isDedent = false
		p.isAttr = false
		p.attrName = ""
		p.attrValue = ""
		p.attrString = ""
	case TokEOF:
		// since p.node is never on the stack until TokNewLine we push to the stack
		// in this scenario
		if p.node.tag != "" {
			p.stack = append(p.stack, p.node)
			p.outputNode(p.node)
		}
		// close the remaining nodes in the stack 
		for len(p.stack) > 0 {
			p.closeLastNode()
		}
	default:
		if !p.isAttr {
			p.node.text += text
		}
	}
}

func (p *Parser) newNode() *Node {
	n := new(Node)
	n.attrs = make(map[string]string)
	return n
}
