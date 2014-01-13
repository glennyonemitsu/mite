package main

import (
	"fmt"
)

type Parser struct {

	debug bool

	Scanner Scanner
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
	dedentCount int

	// flag indicates we are still checking for attribute assignments
	isAttr bool

	isStringFlag bool

	// attrName and attrValue are buffers to determine if TokWord for a tag are 
	// potentially for an attribute, or if they are just normal text
	attrName string
	attrValue string
	attrAssigned bool

}

func (p *Parser) Output() string {
	var tokens []rune
	var text string

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
	p.dedentCount = 0

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
				fmt.Printf("[debug] node:  [%s]\n\n", p.node.Debug())
			}
		}
		if tokens[0] == TokEOF {
			break
		}
	}
	return p.output
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

func (p *Parser) processToken(tok rune, text string) {
	switch tok {
	// indent/dedent/nodent is the first place to look to see if new nodes need
	// to be made. checks are for parent node being certain types like NodeText
	// or NodeComment
	case TokIndent:
		p.isIndent = true
		p.isDedent = false
		p.output += p.node.OpenString()
		p.node = p.newNode()
		p.pushNode(p.node)
	case TokDedent:
		p.isIndent = false
		p.isDedent = true
		p.dedentCount += 1
		n := p.popNode()
		p.output += n.OpenString()
		p.output += n.CloseString()
		p.node = p.newNode()
		p.pushNode(p.node)
	case TokNodent:
		p.isIndent = false
		p.isDedent = false
		switch p.node.Type {
		case NodeRoot:
			p.node = p.newNode()
			p.pushNode(p.node)
		case NodeText:
		default:
			// replace top of stack with new node (pop then push)
			n := p.popNode()
			p.output += n.OpenString()
			p.output += n.CloseString()
			p.node = p.newNode()
			p.pushNode(p.node)
		}

	case TokWord, TokComma:
		switch p.node.Type {
		case NodeNil:
			p.node.Type = NodeTag
			p.node.tag = text
			// found the tag, now check for attributes
			p.isAttr = true
			p.attrAssigned = false
		case NodeTag:
			if p.isAttr {
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
					p.node.attrString = ""
				} else if p.attrName == "" {
					p.node.attrString += text
					p.attrName = text
				} else {
					p.node.text += p.node.attrString
					p.node.text += text
					p.node.attrString = ""
					p.isAttr = false
					p.attrAssigned = false
					p.attrName = ""
					p.attrValue = ""
				}
			} else {
				p.node.text += text
			}
		}
	case TokString:
		switch p.node.Type {
		case NodeTag:
			if p.isAttr {
				if p.attrAssigned {
					if _, found := p.node.attrs[p.attrName]; found {
						p.node.attrs[p.attrName] += " "
						p.node.attrs[p.attrName] += text[1:len(text)-1]
					} else {
						p.node.attrs[p.attrName] = text[1:len(text)-1]
					}

					// reset to look for a new attribute assignment
					p.attrName = ""
					p.attrValue = ""
					p.attrAssigned = false
					p.node.attrString = ""
				} else if p.attrName != "" {
					// TODO using string in attrName shouldn't be allowed
					p.node.text += p.node.attrString
					p.node.text += text[1:len(text)-1]
					p.node.attrString = ""
					p.isAttr = false
					p.attrName = ""
					p.attrValue = ""
					p.attrAssigned = false
				}
			} else {
				p.node.text += text
			}
		case NodeText:
			p.node.text += text
		}
	case TokAssign:
		switch p.node.Type {
		case NodeTag:
			if p.isAttr {
				if p.attrName != "" && p.attrValue == "" {
					p.attrAssigned = true
					p.node.attrString += text
				}
				if p.attrName == "" || p.attrValue != "" {
					p.attrAssigned = false
					p.isAttr = false
					p.node.text += p.node.attrString
					p.node.attrString = ""
				}
			} else {
				p.node.text += text
			}
		case NodeText:
			p.node.text += text
		}
	case TokWhitespace:
		switch p.node.Type {
		case NodeTag:
			if p.isAttr && p.node.attrString != "" {
				p.node.attrString += text

			// skip initial whitespace for text
			} else if p.node.text != "" {
				p.node.text += text
			}
		case NodeText:
			if p.node.text != "" {
				p.node.text += text
			}
		}
	case TokStringFlag:
		switch p.node.Type {
		case NodeNil:
			// new node, but indent indicates continue with text of previous node
			if p.isIndent {
				p.node = p.popNode()
				p.node.text += " "
			} else if !p.isIndent && !p.isDedent {
				p.node = p.popNode()
				p.node.text += " "
				p.node.text += text
			} else {
				p.node.Type = NodeText
			}
		case NodeTag:
			if p.isAttr {
				p.node.text += p.node.attrString
				p.node.attrString = ""
			} else {
				p.node.text += text
			}
			p.isAttr = false
		case NodeText:
			p.node.text += text
		}
	case TokNewLine:
		// up to this point a node has been constructed but not appended to the
		// output
		/*
		if p.isDedent {
			// for dedents, pop off the stack and close the node
			for p.dedentCount > 0 {
				p.closeLastNode()
				p.dedentCount -= 1
			}
			p.isStringFlag = false
		}
		*/
		/*
		if p.isDedent || !p.isIndent {
			// closing one more time because this new node replaces the top node in the
			// stack for dedents or siblings (non indent)
			p.closeLastNode()
		}
		p.outputNode(p.node)
		p.pushNode(p.node)
		*/

		// reset
		//p.node = p.newNode()
		p.isIndent = false
		p.isDedent = false
		p.isAttr = false
		p.attrName = ""
		p.attrValue = ""
	case TokEOF:
		// since p.node is never on the stack until TokNewLine we push to the stack
		// in this scenario
		//p.pushNode(p.node)
		//p.outputNode(p.node)
		// close the remaining nodes in the stack 
		for len(p.stack) > 0 {
			lastNode := p.popNode()
			p.output += lastNode.CloseString()
		}
	default:
		if !p.isAttr {
			p.node.text += text
		}
	}
}

