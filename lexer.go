package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

// A Lexer implements reading of Unicode characters and tokens from an io.Reader.
type Lexer struct {
	src string
	reader *strings.Reader
	pos Position

	lexemes []*Lexeme
	tokens []*Token

	// Error is called for each error encountered. If no Error
	// function is set, the error is reported to os.Stderr.
	Error func(l *Lexer, msg string)

	// ErrorCount is incremented by one for each error encountered.
	ErrorCount int

}

// Init initializes a Lexer with a new source and returns l.
// Error is set to nil, ErrorCount is set to 0
func (l *Lexer) Init(src, filename string) *Lexer {
	l.src = src
	l.reader = strings.NewReader(src)

	l.lexemes = make([]*Lexeme, 0)
	l.tokens = make([]*Token, 0)

	l.pos = Position{}
	l.pos.Filename = filename
	l.pos.Line = 1
	l.pos.Column = 0
	l.pos.Offset = 0

	// initialize public fields
	l.Error = nil
	l.ErrorCount = 0

	return l
}

// Next reads and returns the next Unicode character.
// It returns EOF at the end of the source. It reports
// a read error by calling l.Error, if not nil; otherwise
// it prints an error message to os.Stderr.
func (l *Lexer) Next() rune {
	ch, width, err := l.reader.ReadRune()
	if err != nil {
		if err.Error() != "EOF" {
			l.error(err.Error())
		}
	}
	l.pos.Offset += width
	l.pos.Column += 1
	return ch
}

func (l *Lexer) Peek() rune {
	ch, _, _ := l.reader.ReadRune()
	if ch != 0 { // EOF
		l.reader.UnreadRune()
	}
	return ch
}

func (l *Lexer) error(msg string) {
	l.ErrorCount++
	if l.Error != nil {
		l.Error(l, msg)
	} else {
		fmt.Fprintf(os.Stderr, "lexer error: %s\n", msg)
	}
}

func (l *Lexer) scan() {
	var lex *Lexeme
	if len(l.lexemes) == 0 {
		for {
			lex = l.ScanLexeme()
			l.lexemes = append(l.lexemes, lex)
			if lex.Type == LexEOF || lex.Type == LexError {
				break
			}
		}
	}
}

// Scan reads the next token or Unicode character from source and returns it.
// It returns EOF at the end of the source. It reports scanner errors (read and
// token errors) by calling l.Error, if not nil; otherwise it prints an error
// message to os.Stderr.
func (l *Lexer) scanLexeme() *Lexeme {
	lex := new(Lexeme)
	lex.Pos = l.pos

	ch := l.Peek()
	offsetStart := l.pos.Offset

	switch {
	case unicode.IsLetter(ch) || ch == '_':
		lex.Type = LexWord
		l.scanWord()
	case isDecimal(ch):
		lex.Type = l.scanNumber(ch)
	default:
		switch ch {
		case '=':
			lex.Type = LexAssign
			l.Next()
		case ',':
			lex.Type = LexComma
			l.Next()
		case ' ', '\t':
			lex.Type = LexWhitespace
			l.scanWhitespace()
		case '\n', '\r':
			lex.Type = LexNewLine
			l.scanNewLine(ch)
		case '"', '\'':
			lex.Type = LexString
			l.scanString(ch)
		case '.':
			ch = l.Next()
			if isDecimal(ch) {
				lex.Type = LexFloat
				ch = l.scanMantissa(ch)
				ch = l.scanExponent(ch)
			}
		case '/':
			ch = l.Next()
			if ch == '/' || ch == '*' {
				ch = l.scanComment(ch)
				lex.Type = LexComment
			}
		case '`':
			lex.Type = LexStringFlag
			l.Next()
		case 0: // EOF
			lex.Type = LexEOF
		default:
			l.Next()
		}
	}

	if lex.Type != LexEOF {
		offsetEnd := l.pos.Offset
		lex.Data = l.src[offsetStart:offsetEnd]
	}
	return lex
}


// Following are scan<FOO> methods. These must all call l.Next() enough that
// the l.pos.Offset value is at the end of the token. For example if the source
// is `foo bar baz` and current offset is 4, scanWord will end with the 
// l.pos.Offset at 7. It is safe to assume l.Peek() is always the first rune for
// the scan, and the very least at least one l.Next() has to be made. Scanning
// methods should use l.Peek() liberally.
func (l *Lexer) scanWord() {
	for {
		ch := l.Peek()
		if ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			ch = l.Next()
		} else {
			break
		}
	}
}

func digitVal(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch - 'a' + 10)
	case 'A' <= ch && ch <= 'F':
		return int(ch - 'A' + 10)
	}
	return 16 // larger than any legal digit val
}

func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }

func (l *Lexer) scanMantissa(ch rune) rune {
	for isDecimal(ch) {
		ch = l.Next()
	}
	return ch
}

func (l *Lexer) scanFraction(ch rune) rune {
	if ch == '.' {
		ch = l.scanMantissa(l.Next())
	}
	return ch
}

func (l *Lexer) scanExponent(ch rune) rune {
	if ch == 'e' || ch == 'E' {
		ch = l.Next()
		if ch == '-' || ch == '+' {
			ch = l.Next()
		}
		ch = l.scanMantissa(ch)
	}
	return ch
}

func (l *Lexer) scanNumber(ch rune) LexemeType {
	// isDecimal(ch)
	if ch == '0' {
		// int or float
		ch = l.Next()
		if ch == 'x' || ch == 'X' {
			// hexadecimal int
			ch = l.Next()
			hasMantissa := false
			for digitVal(ch) < 16 {
				ch = l.Next()
				hasMantissa = true
			}
			if !hasMantissa {
				l.error("illegal hexadecimal number")
			}
		} else {
			// octal int or float
			has8or9 := false
			for isDecimal(ch) {
				if ch > '7' {
					has8or9 = true
				}
				ch = l.Next()
			}
			if ch == '.' || ch == 'e' || ch == 'E' {
				// float
				ch = l.scanFraction(ch)
				ch = l.scanExponent(ch)
				return LexFloat
			}
			// octal int
			if has8or9 {
				l.error("illegal octal number")
			}
		}
		return LexInt
	}
	// decimal int or float
	ch = l.scanMantissa(ch)
	if ch == '.' || ch == 'e' || ch == 'E' {
		// float
		ch = l.scanFraction(ch)
		ch = l.scanExponent(ch)
		return LexFloat
	}
	return LexInt
}

func (l *Lexer) scanDigits(ch rune, base, n int) rune {
	for n > 0 && digitVal(ch) < base {
		ch = l.Next()
		n--
	}
	if n > 0 {
		l.error("illegal char escape")
	}
	return ch
}

func (l *Lexer) scanEscape(quote rune) rune {
	ch := l.Next() // read character after '/'
	switch ch {
	case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', quote:
		// nothing to do
		ch = l.Next()
	case '0', '1', '2', '3', '4', '5', '6', '7':
		ch = l.scanDigits(ch, 8, 3)
	case 'x':
		ch = l.scanDigits(l.Next(), 16, 2)
	case 'u':
		ch = l.scanDigits(l.Next(), 16, 4)
	case 'U':
		ch = l.scanDigits(l.Next(), 16, 8)
	default:
		l.error("illegal char escape")
	}
	return ch
}

func (l *Lexer) scanString(quote rune) (n int) {
	ch := l.Next() // read character after quote
	for ch != quote {
		if ch == '\n' || ch < 0 {
			l.error("literal not terminated")
			return
		}
		if ch == '\\' {
			ch = l.scanEscape(quote)
		} else {
			ch = l.Next()
		}
		n++
	}
	return
}

func (l *Lexer) scanComment(ch rune) rune {
	// ch == '/' || ch == '*'
	if ch == '/' {
		// line comment
		ch = l.Next() // read character after "//"
		for ch != '\n' && ch >= 0 {
			ch = l.Next()
		}
		return ch
	}

	// general comment
	ch = l.Next() // read character after "/*"
	for {
		if ch < 0 {
			l.error("comment not terminated")
			break
		}
		ch0 := ch
		ch = l.Next()
		if ch0 == '*' && ch == '/' {
			ch = l.Next()
			break
		}
	}
	return ch
}

func (l *Lexer) scanWhitespace() {
	for {
		ch := l.Peek()
		if ch == ' ' || ch == '\t' {
			ch = l.Next()
		} else {
			break
		}
	}
}

func (l *Lexer) scanNewLine(ch rune) {
	ch = l.Next()
	n := l.Peek()
	if ch == '\n' && n == '\r' {
		l.Next()
	} else if ch == '\r' && n == '\n' {
		l.Next()
	}
}

func (l *Lexer) evaluate() {
}

func (l *Lexer) GetTokens() []*Token {
	if len(l.tokens) == 0 {
		l.scan()
		l.evaluate()
	}
	return l.tokens
}
