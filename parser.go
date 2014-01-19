package main

import (
	"fmt"
	"strings"
)

type Parser struct {

	Debug bool

	Lexer *Lexer
	output string

	// node heirarchy stack
	stack []*Node

	// edge in stack
	node *Node

	// states as the Parser object figures out how to output the tokens

	// new line has started
	isNewLine bool

	isIndent bool
	isDedent bool

	// flag indicates we are still checking for attribute assignments
	isAttr bool

	isStringFlag bool

	// attrName and attrValue are buffers to determine if TokWord for a tag are 
	// potentially for an attribute, or if they are just normal text
	attrName string
	attrValue string
	attrAssigned bool

}

func (p *Parser) BuildTree() string {
}

func (p *Parser) Output() string {
	var tok *Token

	p.node = p.newNode()
	p.node.Type = NodeRoot
	p.stack = make([]*Node, 0)
	p.stack = append(p.stack, p.node)

	p.output = ""
	p.isNewLine = true
	p.isIndent = false
	p.isDedent = false
	p.isAttr = false
	p.attrName = ""
	p.attrValue = ""
	p.attrAssigned = false

	for {
		tok = p.Lexer.Scan()
		if p.Debug {
			fmt.Printf("[debug] token: [%s]\n", tok.Debug())
		}
		p.processToken(tok)
		if p.Debug {
			fmt.Printf("[debug] node:  [%s]\n\n", p.node.Debug())
		}
		if tok.Type == TokEOF {
			break
		}
	}
	return p.output
}

func (p *Parser) trimOutput() {
	p.output = strings.TrimSpace(p.output)
}

func (p *Parser) newNode() *Node {
	n := new(Node)
	n.attrs = make(map[string]string)
	return n
}

func (p *Parser) lastNode() *Node {
	if len(p.stack) > 0 {
		return p.stack[len(p.stack) - 1]
	} else {
		return nil
	}
}

func (p *Parser) popNode() *Node {
	n := p.lastNode()
	if n != nil {
		p.stack = p.stack[0:len(p.stack)-1]
	}
	return n
}

func (p *Parser) pushNode(n *Node) {
	ln := p.lastNode()
	n.parent = ln
	p.stack = append(p.stack, n)
}

func (p *Parser) dedentFromStack() {
	n := p.lastNode()
	if n != nil && n.Type != NodeRoot {
		p.output += n.OpenString()
		p.trimOutput()
		// newlines in NodeText have spaces. Adding one here for consistency
		if n.Type == NodeText {
			p.output += " "
		}
		p.output += n.CloseString()
		p.popNode()
	}
}

func (p *Parser) processToken(tok *Token) {
	if p.node.Type == NodeText {
		if tok.Type == TokNewLine {
			p.node.text += " "
		} else if tok.Type != TokWhitespace || p.node.text != "" {
			p.node.text += tok.Data
		}
		return
	}

	switch tok.Type {
	case TokWord, TokComma:
		switch p.node.Type {
		case NodeNil:
			p.node.Type = NodeTag
			p.node.tag = tok.Data
			// found the tag, now check for attributes
			p.isAttr = true
			p.attrAssigned = false
		case NodeTag:
			if p.isAttr {
				if p.attrAssigned {
					if _, found := p.node.attrs[p.attrName]; found {
						p.node.attrs[p.attrName] += " "
						p.node.attrs[p.attrName] += tok.Data
					} else {
						p.node.attrs[p.attrName] = tok.Data
					}

					// reset to look for a new attribute assignment
					p.attrName = ""
					p.attrValue = ""
					p.attrAssigned = false
					p.node.attrString = ""
				} else if p.attrName == "" {
					p.node.attrString += tok.Data
					p.attrName = tok.Data
				} else {
					p.node.text += p.node.attrString
					p.node.text += tok.Data
					p.node.attrString = ""
					p.isAttr = false
					p.attrAssigned = false
					p.attrName = ""
					p.attrValue = ""
				}
			} else {
				p.node.text += tok.Data
			}
		}
	case TokString:
		switch p.node.Type {
		case NodeTag:
			if p.isAttr {
				if p.attrAssigned {
					if _, found := p.node.attrs[p.attrName]; found {
						p.node.attrs[p.attrName] += " "
						p.node.attrs[p.attrName] += tok.Data[1:len(tok.Data)-1]
					} else {
						p.node.attrs[p.attrName] = tok.Data[1:len(tok.Data)-1]
					}

					// reset to look for a new attribute assignment
					p.attrName = ""
					p.attrValue = ""
					p.attrAssigned = false
					p.node.attrString = ""
				} else if p.attrName != "" {
					// TODO using string in attrName shouldn't be allowed
					p.node.text += p.node.attrString
					p.node.text += tok.Data[1:len(tok.Data)-1]
					p.node.attrString = ""
					p.isAttr = false
					p.attrName = ""
					p.attrValue = ""
					p.attrAssigned = false
				}
			} else {
				p.node.text += tok.Data
			}
		case NodeText:
			p.node.text += tok.Data
		}
	case TokAssign:
		switch p.node.Type {
		case NodeTag:
			if p.isAttr {
				if p.attrName != "" && p.attrValue == "" {
					p.attrAssigned = true
					p.node.attrString += tok.Data
				}
				if p.attrName == "" || p.attrValue != "" {
					p.attrAssigned = false
					p.isAttr = false
					p.node.text += p.node.attrString
					p.node.attrString = ""
				}
			} else {
				p.node.text += tok.Data
			}
		case NodeText:
			p.node.text += tok.Data
		}
	case TokWhitespace:
		switch p.node.Type {
		case NodeTag:
			if p.isAttr && p.node.attrString != "" {
				p.node.attrString += tok.Data

			// skip initial whitespace for text
			} else if p.node.text != "" {
				p.node.text += tok.Data
			}
		}
	case TokStringFlag:
		switch p.node.Type {
		case NodeNil:
			// new node, but indent indicates continue with text of previous node
			p.node.Type = NodeText
		case NodeTag:
			if p.isAttr {
				p.node.text += p.node.attrString
				p.node.attrString = ""
			} else {
				p.node.text += tok.Data
			}
			p.isAttr = false
		}
	case TokNewLine:
		// reset
		p.isIndent = false
		p.isDedent = false
		p.isAttr = false
		p.attrName = ""
		p.attrValue = ""
		p.attrAssigned = false
	case TokEOF:
		// close the remaining nodes in the stack 
		for len(p.stack) > 0 {
			lastNode := p.popNode()
			p.trimOutput()
			p.output += lastNode.CloseString()
		}
	// nop
	default:
		if !p.isAttr {
			p.node.text += tok.Data
		}
	}
}

