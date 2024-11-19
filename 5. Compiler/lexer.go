package main

import "fmt"

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace() // 공백 및 개행 제거

	fmt.Printf("Processing Char: '%c'\n", l.ch) // 디버깅 메시지

	switch l.ch {
	case '=':
		tok = Token{Type: TokenAssign, Literal: string(l.ch)}
	case '+':
		tok = Token{Type: TokenPlus, Literal: string(l.ch)}
	case ';':
		tok = Token{Type: TokenSemicolon, Literal: string(l.ch)}
	case '(':
		tok = Token{Type: TokenLParen, Literal: string(l.ch)}
	case ')':
		tok = Token{Type: TokenRParen, Literal: string(l.ch)}
	case 0: // EOF
		tok = Token{Type: TokenEOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			literal := l.readIdentifier()
			tok.Type = lookupIdent(literal)
			tok.Literal = literal
			return tok
		} else if isDigit(l.ch) {
			tok.Type = TokenInt
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = Token{Type: TokenEOF, Literal: ""}
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\n' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func lookupIdent(ident string) TokenType {
	if ident == "let" {
		return TokenLet
	} else if ident == "print" {
		return TokenPrint
	}
	return TokenIdent
}
