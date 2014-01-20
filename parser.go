package main

import (
	"fmt"
)

type Parser struct {

	Debug bool

	Lexer *Lexer
	output string

	root *Node

}

func (p *Parser) BuildTree() {
	// build state
	newNode := true
	skip := 0

	// tree start
	var n, edge *Node
	edge = new(Node)
	edge.Type = NodeRoot
	edge.parent = nil
	edge.index = 0
	p.root = edge

	tokens := p.Lexer.GetTokens()
	for _, tok := range tokens {
		if skip > 0 {
			skip -= 1
			continue
		}
		p.debugLog(fmt.Sprintf("[debug] token: [%s]\n", tok.Debug()))

		switch tok.Type {
		case TokEOF:
			break
		case TokIdent:
			if newNode {
				n = new(Node)
				n.Type = NodeTag
				n.tag = tok.Value
				edge.appendChild(n)
				edge = n
				newNode = false
			} else {
				// currently doing a tag, so see if attributes are correct or switch to
				// new text node
				if edge.Type == NodeTag {
					edge.text += " "
					edge.text += tok.Value
				}
			}
		case TokDedent:
			if edge.Type != NodeRoot {
				edge = edge.parent
			}
		case TokError:
		case TokString:
		case TokNumber:
		case TokInt:
		case TokFloat:
		case TokStringFlag:
		case TokComment:
		case TokNewLine:
		case TokIndent:
		case TokComma:
		case TokAssign:
		}
	}
}

func (p *Parser) debugLog(msg string) {
	if p.Debug {
		fmt.Printf(msg)
	}
}

