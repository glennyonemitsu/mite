package main

import (
	"fmt"
)

type LexemeType int

// The result of Scan is one of the following tokens or a Unicode character.
const (
	LexEOF LexemeType = -(iota + 1)
	LexError
	LexNull
	LexWord
	LexNumber
	LexInt
	LexFloat
	LexString
	LexStringFlag
	LexComment
	LexNewLine
	LexLineContinue
	LexComma
	LexWhitespace
	LexAssign
)

var lexemeString = map[LexemeType]string{
	LexEOF:      "EOF",
	LexError:     "Error",
	LexNull:     "Null",
	LexWord:     "Word",
	LexInt:       "Int",
	LexFloat:     "Float",
	LexString:    "String",
	LexStringFlag: "StringFlag",
	LexComment:   "Comment",
	LexNewLine:   "NewLine",
	LexLineContinue:   "LineContinue",
	LexWhitespace:	"Whitespace",
	LexAssign:	"Assign",
	LexComma:	"Comma",
}

type Lexeme struct {
	Type LexemeType
	Value string
	Pos Position
}

func (l *Lexeme) TypeString() string {
	return lexemeString[l.Type]
}

func (l *Lexeme) String() string {
	return l.Value
}

func (l *Lexeme) Debug() string {
	output := ""
	output += fmt.Sprintf("[Type:%s]", l.TypeString())
	output += fmt.Sprintf("[Value:%s]", l.Value)
	output += fmt.Sprintf("[Pos.Filename:%s]", l.Pos.Filename)
	output += fmt.Sprintf("[Pos.Offset:%d]", l.Pos.Offset)
	output += fmt.Sprintf("[Pos.Line:%d]", l.Pos.Line)
	output += fmt.Sprintf("[Pos.Column:%d]", l.Pos.Column)
	return output
}

