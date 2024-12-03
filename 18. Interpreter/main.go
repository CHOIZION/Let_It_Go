package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type TokenType string

const (
	TokenEOF       TokenType = "EOF"
	TokenNumber    TokenType = "NUMBER"
	TokenIdent     TokenType = "IDENT"
	TokenAssign    TokenType = "ASSIGN"
	TokenPlus      TokenType = "PLUS"
	TokenMinus     TokenType = "MINUS"
	TokenAsterisk  TokenType = "ASTERISK"
	TokenSlash     TokenType = "SLASH"
	TokenLParen    TokenType = "LPAREN"
	TokenRParen    TokenType = "RPAREN"
	TokenIf        TokenType = "IF"
	TokenElse      TokenType = "ELSE"
	TokenWhile     TokenType = "WHILE"
	TokenLBrace    TokenType = "LBRACE"
	TokenRBrace    TokenType = "RBRACE"
	TokenSemicolon TokenType = "SEMICOLON"
	TokenEq        TokenType = "EQ"
	TokenNeq       TokenType = "NEQ"
	TokenLt        TokenType = "LT"
	TokenGt        TokenType = "GT"
)

type Token struct {
	Type    TokenType
	Literal string
}

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
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	var tok Token

	switch l.ch {
	case '+':
		tok = Token{Type: TokenPlus, Literal: "+"}
	case '-':
		tok = Token{Type: TokenMinus, Literal: "-"}
	case '*':
		tok = Token{Type: TokenAsterisk, Literal: "*"}
	case '/':
		tok = Token{Type: TokenSlash, Literal: "/"}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: TokenEq, Literal: "=="}
		} else {
			tok = Token{Type: TokenAssign, Literal: "="}
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = Token{Type: TokenNeq, Literal: "!="}
		}
	case '<':
		tok = Token{Type: TokenLt, Literal: "<"}
	case '>':
		tok = Token{Type: TokenGt, Literal: ">"}
	case '(':
		tok = Token{Type: TokenLParen, Literal: "("}
	case ')':
		tok = Token{Type: TokenRParen, Literal: ")"}
	case '{':
		tok = Token{Type: TokenLBrace, Literal: "{"}
	case '}':
		tok = Token{Type: TokenRBrace, Literal: "}"}
	case ';':
		tok = Token{Type: TokenSemicolon, Literal: ";"}
	case 0:
		tok.Literal = ""
		tok.Type = TokenEOF
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			tokType := LookupIdent(ident)
			tok = Token{Type: tokType, Literal: ident}
			return tok
		} else if isDigit(l.ch) {
			num := l.readNumber()
			tok = Token{Type: TokenNumber, Literal: num}
			return tok
		} else {
			tok = Token{Type: TokenEOF, Literal: ""}
		}
	}
	l.readChar()
	return tok
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

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

var keywords = map[string]TokenType{
	"if":    TokenIf,
	"else":  TokenElse,
	"while": TokenWhile,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdent
}

type Node interface{}

type Program struct {
	Statements []Statement
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type LetStatement struct {
	Name  string
	Value Expression
}

func (ls *LetStatement) statementNode() {}

type ExpressionStatement struct {
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

type IntegerLiteral struct {
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

type Identifier struct {
	Value string
}

func (id *Identifier) expressionNode() {}

type PrefixExpression struct {
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}

type InfixExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}

type IfExpression struct {
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode() {}

type BlockStatement struct {
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

type WhileStatement struct {
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode() {}

type Parser struct {
	l              *Lexer
	curToken       Token
	peekToken      Token
	prefixParseFns map[TokenType]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{
		l:              l,
		prefixParseFns: make(map[TokenType]prefixParseFn),
		infixParseFns:  make(map[TokenType]infixParseFn),
	}
	p.nextToken()
	p.nextToken()

	p.registerPrefix(TokenIdent, p.parseIdentifier)
	p.registerPrefix(TokenNumber, p.parseIntegerLiteral)
	p.registerPrefix(TokenMinus, p.parsePrefixExpression)
	p.registerPrefix(TokenLParen, p.parseGroupedExpression)

	p.registerInfix(TokenPlus, p.parseInfixExpression)
	p.registerInfix(TokenMinus, p.parseInfixExpression)
	p.registerInfix(TokenAsterisk, p.parseInfixExpression)
	p.registerInfix(TokenSlash, p.parseInfixExpression)
	p.registerInfix(TokenEq, p.parseInfixExpression)
	p.registerInfix(TokenNeq, p.parseInfixExpression)
	p.registerInfix(TokenLt, p.parseInfixExpression)
	p.registerInfix(TokenGt, p.parseInfixExpression)

	return p
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

func (p *Parser) registerPrefix(tokenType TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	for p.curToken.Type != TokenEOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[TokenType]int{
	TokenEq:       EQUALS,
	TokenNeq:      EQUALS,
	TokenLt:       LESSGREATER,
	TokenGt:       LESSGREATER,
	TokenPlus:     SUM,
	TokenMinus:    SUM,
	TokenAsterisk: PRODUCT,
	TokenSlash:    PRODUCT,
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case TokenIdent:
		if p.peekToken.Type == TokenAssign {
			return p.parseAssignmentStatement()
		}
		return p.parseExpressionStatement()
	case TokenIf:
		return p.parseIfStatement()
	case TokenWhile:
		return p.parseWhileStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseAssignmentStatement() Statement {
	stmt := &LetStatement{Name: p.curToken.Literal}
	p.nextToken() // '='
	p.nextToken() // 다음 토큰으로 이동
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekToken.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() Statement {
	stmt := &ExpressionStatement{Expression: p.parseExpression(LOWEST)}
	if p.peekToken.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != TokenSemicolon && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() Expression {
	value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		return nil
	}
	return &IntegerLiteral{Value: value}
}

func (p *Parser) parsePrefixExpression() Expression {
	expression := &PrefixExpression{
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if p.curToken.Type != TokenRParen {
		return nil
	}
	return exp
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &InfixExpression{
		Left:     left,
		Operator: p.curToken.Literal,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseIfStatement() Statement {
	expr := &IfExpression{}
	p.nextToken()
	expr.Condition = p.parseExpression(LOWEST)
	if p.curToken.Type != TokenLBrace {
		return nil
	}
	expr.Consequence = p.parseBlockStatement()
	if p.peekToken.Type == TokenElse {
		p.nextToken()
		p.nextToken()
		expr.Alternative = p.parseBlockStatement()
	}
	stmt := &ExpressionStatement{Expression: expr}
	return stmt
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{}
	p.nextToken()
	for p.curToken.Type != TokenRBrace && p.curToken.Type != TokenEOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseWhileStatement() Statement {
	stmt := &WhileStatement{}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if p.curToken.Type != TokenLBrace {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

type Environment struct {
	store map[string]int64
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]int64)}
}

func (e *Environment) Get(name string) (int64, bool) {
	val, ok := e.store[name]
	return val, ok
}

func (e *Environment) Set(name string, val int64) int64 {
	e.store[name] = val
	return val
}

func Eval(node Node, env *Environment) int64 {
	switch node := node.(type) {
	case *Program:
		var result int64
		for _, stmt := range node.Statements {
			result = Eval(stmt, env)
		}
		return result
	case *LetStatement:
		val := Eval(node.Value, env)
		env.Set(node.Name, val)
		return val
	case *ExpressionStatement:
		return Eval(node.Expression, env)
	case *IntegerLiteral:
		return node.Value
	case *Identifier:
		if val, ok := env.Get(node.Value); ok {
			return val
		}
		return 0
	case *PrefixExpression:
		right := Eval(node.Right, env)
		switch node.Operator {
		case "-":
			return -right
		default:
			return 0
		}
	case *InfixExpression:
		left := Eval(node.Left, env)
		right := Eval(node.Right, env)
		switch node.Operator {
		case "+":
			return left + right
		case "-":
			return left - right
		case "*":
			return left * right
		case "/":
			return left / right
		case "==":
			if left == right {
				return 1
			} else {
				return 0
			}
		case "!=":
			if left != right {
				return 1
			} else {
				return 0
			}
		case "<":
			if left < right {
				return 1
			} else {
				return 0
			}
		case ">":
			if left > right {
				return 1
			} else {
				return 0
			}
		default:
			return 0
		}
	case *IfExpression:
		condition := Eval(node.Condition, env)
		if condition != 0 {
			return Eval(node.Consequence, env)
		} else if node.Alternative != nil {
			return Eval(node.Alternative, env)
		}
		return 0
	case *BlockStatement:
		var result int64
		for _, stmt := range node.Statements {
			result = Eval(stmt, env)
		}
		return result
	case *WhileStatement:
		var result int64
		for {
			condition := Eval(node.Condition, env)
			if condition == 0 {
				break
			}
			result = Eval(node.Body, env)
		}
		return result
	default:
		return 0
	}
}

func main() {
	env := NewEnvironment()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("간단한 인터프리터입니다. 종료하려면 'exit'를 입력하세요.")
	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if strings.TrimSpace(line) == "exit" {
			fmt.Println("인터프리터를 종료합니다.")
			break
		}
		lexer := NewLexer(line)
		parser := NewParser(lexer)
		program := parser.ParseProgram()
		result := Eval(program, env)
		fmt.Println(result)
	}
}
