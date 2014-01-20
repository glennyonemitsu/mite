package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

const tabIndentValue = 8

// A Lexer implements reading of Unicode characters and tokens from an io.Reader.
type Lexer struct {
	src string
	reader *strings.Reader
	pos Position
	tabIndentValue int

	Debug bool

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
	l.tabIndentValue = tabIndentValue

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
	l.Debug = false

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
			lex = l.scanLexeme()
			l.lexemes = append(l.lexemes, lex)
			if lex.Type == LexEOF || lex.Type == LexError {
				break
			}
		}
	}
}

// scanLexeme reads the next token or Unicode character from source and returns it.
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
			l.pos.Line += 1
			l.pos.Column = 0
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
		lex.Value = l.src[offsetStart:offsetEnd]
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

func (l *Lexer) pushToken(token *Token) {
	if token != nil {
		l.tokens = append(l.tokens, token)
	}
}

func (l *Lexer) pushTokenType(t TokenType, pos Position) {
	token := new(Token)
	token.Type = t
	token.Pos = pos
	l.pushToken(token)
}

func (l *Lexer) evaluate() {
	// stack of indentation levels
	indents := make([]int, 0)
	// nested bracket level
	//brackets := 0
	// grabbing all other lexemes as strings, usually after attribute assignment
	textMode := false
	firstWord := true
	newLine := true
	// some lexemes will require looking ahead and determining if any can be 
	// skipped such as consecutive newlines.
	skip := 0

	var token *Token
	//var lastToken *Token // only use when needed (ex. flattening newlines)
	indents = append(indents, 0)

	for i, lex := range l.lexemes {
		if l.Debug {
			fmt.Printf("[debug] lexeme: [%s]\n", lex.Debug())
		}
		if skip > 0 {
			skip -= 1
			continue
		}
		switch lex.Type {
		case LexEOF:
			for len(indents) > 1 { // 1 since first element is 0 indent
				token = new(Token)
				token.Type = TokDedent
				token.Pos = lex.Pos
				l.pushToken(token)
				indents = indents[0:len(indents)-1]
			}
			token = new(Token)
			token.Type = TokEOF
			token.Pos = lex.Pos
			l.pushToken(token)
		case LexError:
			token = new(Token)
			token.Type = TokError
			token.Value = lex.Value
			token.Pos = lex.Pos
			l.pushToken(token)
		case LexNull:
		case LexWord:
			if firstWord {
				l.pushToken(token)
				token = new(Token)
				token.Type = TokIdent
				token.Value = lex.Value
				token.Pos = lex.Pos
				l.pushToken(token)
				token = nil

				firstWord = false
				textMode = true
			} else if textMode {
				if token == nil {
					token = new(Token)
					token.Type = TokString
					token.Value = lex.Value
					token.Pos = lex.Pos
				} else {
					token.Value += lex.Value
				}
			} else {
				l.pushToken(token)
				token = new(Token)
				token.Type = TokIdent
				token.Value = lex.Value
				token.Pos = lex.Pos
				l.pushToken(token)
				token = nil
			}
		case LexNumber:
		case LexInt:
		case LexFloat:
		case LexString:
			l.pushToken(token)
			token = nil
			l.pushTokenType(TokString, l.pos)
		case LexStringFlag:
			l.pushToken(token)
			token = nil
			l.pushTokenType(TokStringFlag, l.pos)
		case LexComment:
		case LexNewLine:
			if token != nil {
				l.pushToken(token)
				token = nil
			}
			l.pushTokenType(TokNewLine, l.pos)
			textMode = false
			firstWord = true
			newLine = true
			// keep looking ahead fo LexNewLine to flatten
			nonNewLine := false
			for j := 1; !nonNewLine; j += 1 {
				nl := l.lexemes[i + j]
				if nl.Type == LexNewLine {
					skip += 1
				} else {
					nonNewLine = true
				}
			}
		case LexLineContinue:
			if textMode {
				token.Value += lex.Value
			}
		case LexComma:
			l.pushToken(token)
			token = nil
			l.pushTokenType(TokComma, l.pos)
		case LexWhitespace:
			// determine indent level
			if newLine {
				indentValue := 0
				lastIndentValue := indents[len(indents) - 1]
				for ch := range lex.Value {
					if ch == '\t' {
						indentValue += l.tabIndentValue
					} else if ch == ' ' {
						indentValue += 1
					}
				}
				if lastIndentValue > indentValue {
					l.pushTokenType(TokIndent, l.pos)
				} else if lastIndentValue < indentValue {
					// keep popping indents
					for lastIndentValue < indentValue {
						l.pushTokenType(TokDedent, l.pos)
						indents = indents[0:len(indents)-1]
						lastIndentValue = indents[len(indents) - 1]
					}
				}
				newLine = false
			}
		case LexAssign:
			l.pushToken(token)
			token = nil
			l.pushTokenType(TokAssign, l.pos)
		}
	}
}

func (l *Lexer) GetTokens() []*Token {
	if len(l.tokens) == 0 {
		l.scan()
		l.evaluate()
	}
	return l.tokens
}