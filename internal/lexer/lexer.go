package lexer

import (
	"parrot/internal/token"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	input        []rune
	position     int
	nextPosition int
	ch           rune // current char being scanned
}

func New(input string) *Lexer {
	l := &Lexer{input: []rune(input)}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() (tok token.Token) {
	l.skipWhitespace()
	switch l.ch {
	case '+':
		tok = l.newToken(token.ADD, "+")
	case '-':
		tok = l.newToken(token.MINUS, "-")
	case '*':
		tok = l.newToken(token.MUL, "*")
	case '/':
		tok = l.newToken(token.DIV, "/")
	case '%':
		tok = l.newToken(token.MOD, "%")
	case '=':
		if l.peek() == '=' {
			tok = l.newToken(token.EQ, "==")
			l.readChar()
		} else {
			tok = l.newToken(token.ASSIGN, "=")
		}
	case '>':
		if l.peek() == '=' {
			tok = l.newToken(token.GE, ">=")
			l.readChar()
		} else {
			tok = l.newToken(token.GT, ">")
		}
	case '<':
		if l.peek() == '=' {
			tok = l.newToken(token.LE, "<=")
			l.readChar()
		} else {
			tok = l.newToken(token.LT, "<")
		}
	case '!':
		if l.peek() == '=' {
			tok = l.newToken(token.NOTEQ, "!=")
			l.readChar()
		} else {
			tok = l.newToken(token.BANG, "!")
		}
	case '.':
		if l.peek() == '.' {
			tok = l.newToken(token.DOTDOT, "..")
			l.readChar()
		} else {
			tok = l.newToken(token.DOT, ".")
		}
	case '(':
		tok = l.newToken(token.LPAR, "(")
	case ')':
		tok = l.newToken(token.RPAR, ")")
	case '[':
		tok = l.newToken(token.LBRK, "[")
	case ']':
		tok = l.newToken(token.RBRK, "]")
	case '{':
		tok = l.newToken(token.LBRACE, "{")
	case '}':
		tok = l.newToken(token.RBRACE, "}")
	case ',':
		tok = l.newToken(token.COMMA, ",")
	case ';':
		tok = l.newToken(token.SEMICOLON, ";")
	case ':':
		tok = l.newToken(token.COLON, ":")
	case '\'':
		pos := l.position
		value := l.readString('\'')
		tok = l.newToken(token.STR, value, pos)
	case '"':
		pos := l.position
		value := l.readString('"')
		tok = l.newToken(token.STR, value, pos)
	case 0:
		tok = l.newToken(token.EOF, "")
	default:
		pos := l.position
		if isDigit(l.ch) {
			number := l.readNumber()
			return l.newToken(token.NUM, number, pos)
		}
		identifier := l.readIdentifier()
		return l.newToken(token.LookupKeyWord(identifier), identifier, pos)
	}
	l.readChar()
	return
}

func (l *Lexer) newToken(tokenType token.Type, literal string, pos ...int) token.Token {
	var p int
	if len(pos) > 0 {
		p = pos[0]
	} else {
		p = l.position
	}
	return token.Token{
		Type:    tokenType,
		Literal: literal,
		Pos:     p,
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isIdentifier(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' ||
		ch >= utf8.RuneSelf && unicode.IsLetter(ch) || isDigit(ch)
}

func (l *Lexer) readIdentifier() string {
	pos := l.position
	for isIdentifier(l.ch) {
		l.readChar()
	}
	return string(l.input[pos:l.position])
}

func (l *Lexer) readNumber() string {
	pos := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return string(l.input[pos:l.position])
}

func (l *Lexer) readChar() {
	l.ch = l.peek()
	l.position = l.nextPosition
	l.nextPosition += 1
}

func (l *Lexer) readString(closing rune) string {
	pos := l.position + 1
	for {
		l.readChar()
		if l.ch == closing || l.ch == 0 {
			break
		}
	}
	return string(l.input[pos:l.position])
}

func (l *Lexer) peek() rune {
	if l.nextPosition >= len(l.input) {
		return 0
	}
	return l.input[l.nextPosition]
}
