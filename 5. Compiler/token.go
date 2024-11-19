package main

type TokenType string

const (
	TokenLet       TokenType = "LET"
	TokenIdent     TokenType = "IDENT"
	TokenAssign    TokenType = "ASSIGN"
	TokenInt       TokenType = "INT"
	TokenPlus      TokenType = "PLUS"
	TokenPrint     TokenType = "PRINT"
	TokenSemicolon TokenType = "SEMICOLON"
	TokenLParen    TokenType = "LPAREN"
	TokenRParen    TokenType = "RPAREN"
	TokenEOF       TokenType = "EOF"
)

type Token struct {
	Type    TokenType
	Literal string
}
