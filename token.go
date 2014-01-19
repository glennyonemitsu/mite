package main

import (
	"fmt"
)

type TokenType int

// The result of Scan is one of the following tokens or a Unicode character.
const (
	TokEOF TokenType = -(iota + 1)
	TokError
	TokNull
	TokWord
	TokNumber
	TokInt
	TokFloat
	TokString
	TokStringFlag
	TokComment
	TokNewLine
	TokLineContinue
	TokComma
	TokWhitespace
	TokAssign
)

var tokenString = map[TokenType]string{
	TokEOF:      "EOF",
	TokError:     "Error",
	TokNull:     "Null",
	TokWord:     "Word",
	TokInt:       "Int",
	TokFloat:     "Float",
	TokString:    "String",
	TokStringFlag: "StringFlag",
	TokComment:   "Comment",
	TokNewLine:   "NewLine",
	TokLineContinue:   "LineContinue",
	TokWhitespace:	"Whitespace",
	TokAssign:	"Assign",
	TokComma:	"Comma",
}

type Token struct {
	Type TokenType
	Data string
	Indent int
	Pos Position
}

func (t *Token) TypeString() string {
	return tokenString[t.Type]
}

func (t *Token) String() string {
	return t.Data
}

func (t *Token) Debug() string {
	output := ""
	output += fmt.Sprintf("[Type:%s]", t.TypeString())
	output += fmt.Sprintf("[Data:%s]", t.String())
	output += fmt.Sprintf("[Indent:%d]", t.Indent)
	output += fmt.Sprintf("[Pos.Filename:%s]", t.Pos.Filename)
	output += fmt.Sprintf("[Pos.Offset:%d]", t.Pos.Offset)
	output += fmt.Sprintf("[Pos.Line:%d]", t.Pos.Line)
	output += fmt.Sprintf("[Pos.Column:%d]", t.Pos.Column)
	return output
}

