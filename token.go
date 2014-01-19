package main

import (
	"fmt"
)

type TokenType int

// The result of Scan is one of the following tokens or a Unicode character.
const (
	TokEOF TokenType = -(iota + 1)
	TokError
	TokIdent
	TokString
	TokNumber
	TokInt
	TokFloat
	TokStringFlag
	TokComment
	TokNewLine
	TokIndent
	TokDedent
	TokComma
	TokAssign
)

var tokenString = map[TokenType]string{
	TokEOF:      "EOF",
	TokError:     "Error",
	TokIdent:     "Identifier",
	TokString:    "String",
	TokInt:       "Int",
	TokFloat:     "Float",
	TokStringFlag: "StringFlag",
	TokComment:   "Comment",
	TokNewLine:   "NewLine",
	TokIndent: "Indent",
	TokDedent: "Dedent",
	TokAssign:	"Assign",
	TokComma:	"Comma",
}

type Token struct {
	Type TokenType
	Value string
	Pos Position
}

func (t *Token) TypeString() string {
	return tokenString[t.Type]
}

func (t *Token) String() string {
	return t.Value
}

func (t *Token) Debug() string {
	output := ""
	output += fmt.Sprintf("[Type:%s]", t.TypeString())
	output += fmt.Sprintf("[Value:%s]", t.Value)
	output += fmt.Sprintf("[Pos.Filename:%s]", t.Pos.Filename)
	output += fmt.Sprintf("[Pos.Offset:%d]", t.Pos.Offset)
	output += fmt.Sprintf("[Pos.Line:%d]", t.Pos.Line)
	output += fmt.Sprintf("[Pos.Column:%d]", t.Pos.Column)
	return output
}

