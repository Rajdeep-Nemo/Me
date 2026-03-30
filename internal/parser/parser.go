package parser

import (
	"fmt"
	"pluesi/internal/ast"
	"pluesi/internal/token"
)

// Parser struct to hold the tokens and current position
type Parser struct {
	tokens []token.Token // List of tokens to parse
	pos    int           // Current position in the token list
	errors []string      // List of errors encountered during parsing
}

// New creates a new parser instance with the given tokens
func New(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

// To retrieve errors encountered during parsing
func (p *Parser) Errors() []string {
	return p.errors
}

// Helper function to get the current token
func (p *Parser) currentToken() token.Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return token.Token{Type: token.END_OF_FILE, Lexeme: ""}
}

// Helper function to peek at the next token without advancing the position
func (p *Parser) peekToken() token.Token {
	if p.pos+1 < len(p.tokens) {
		return p.tokens[p.pos+1]
	}
	return token.Token{Type: token.END_OF_FILE, Lexeme: ""}
}

// Helper function to check if the current token is of the expected type
func (p *Parser) check(tt token.TokenType) bool {
	return p.currentToken().Type == tt
}

// Helper function to advance to the next token
func (p *Parser) advance() token.Token {
	t := p.currentToken()
	p.pos += 1
	return t
}

// Helper function to add an error message to the parser's error list
func (p *Parser) errorf(format string, args ...any) {
	p.errors = append(p.errors, fmt.Sprintf(format, args...))
}

// Helper function to check if the current token is of the expected type
func (p *Parser) expect(tt token.TokenType) (token.Token, bool) {
	t := p.currentToken()
	if t.Type != tt {
		p.errorf("expected %v but got %q at line %d", tt, t.Lexeme, t.Line)
		return token.Token{}, false
	}
	p.advance()
	return t, true
}

// Loops through the tokens and parse statements until the end of the file is reached
func (p *Parser) parseProgram() *ast.Program {
	program := &ast.Program{}
	for !p.check(token.END_OF_FILE) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program
}

// Checks the current token and decides which statement parsing function to call based on the token type
func (p *Parser) parseStatement() ast.Statement {
	switch p.currentToken().Type {
	case token.LET:
		return p.parseLetStatement()
	case token.CONST:
		return p.parseConstStatement()
	case token.IDENTIFIER:
		if _, ok := assignOperators[p.peekToken().Type]; ok {
			return p.parseAssignmentStatement()
		}
		fallthrough
	default:
		p.advance()
		return nil
	}
}

// Type tokens for type annotations
var typeTokens = map[token.TokenType]string{
	token.I8:     "i8",
	token.I16:    "i16",
	token.I32:    "i32",
	token.I64:    "i64",
	token.U8:     "u8",
	token.U16:    "u16",
	token.U32:    "u32",
	token.U64:    "u64",
	token.F32:    "f32",
	token.F64:    "f64",
	token.CHAR:   "char",
	token.STRING: "string",
	token.BOOL:   "bool",
}

// Parses a type annotation after a colon in a let/const statement, e.g. `: i32`
func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	t := p.currentToken()
	name, ok := typeTokens[t.Type]
	if !ok {
		p.errorf("expected a type but got %q at line %d", t.Lexeme, t.Line)
		return nil
	}
	p.advance()
	return &ast.TypeAnnotation{Token: t, Name: name}
}

// Helper function to consume a semicolon if it's present, since semicolons are optional
func (p *Parser) consumeSemicolon() {
	if p.check(token.SEMICOLON) {
		p.advance()
	}
}

// Parses a let statement, which can optionally include a type annotation.
func (p *Parser) parseLetStatement() *ast.LetStatement {
	// Consume let token
	letToken := p.advance()
	// Expect and consume identifier
	nameToken, ok := p.expect(token.IDENTIFIER)
	if !ok {
		return nil
	}
	name := &ast.Identifier{Token: nameToken, Value: nameToken.Lexeme}
	// if ':' parse type
	if p.check(token.COLON) {
		p.advance()
		typeHint := p.parseTypeAnnotation()
		if p.check(token.EQUAL) {
			p.advance()
			value := p.parseExpression()
			p.consumeSemicolon()
			return &ast.LetStatement{Token: letToken, Name: name, TypeHint: typeHint, Value: value}
		} else {
			p.consumeSemicolon()
			return &ast.LetStatement{Token: letToken, Name: name, TypeHint: typeHint, Value: nil}
		}
	} else if p.check(token.EQUAL) { // if '=' parse expression
		p.advance()
		value := p.parseExpression()
		p.consumeSemicolon()
		return &ast.LetStatement{Token: letToken, Name: name, TypeHint: nil, Value: value}
	} else { // if nothing then error
		p.errorf("uninitialized variable %q must have a type annotation at line %d", name.Value, p.currentToken().Line)
	}
	return nil
}

// Parses a const statement, which must include a type annotation and an initializer.
func (p *Parser) parseConstStatement() *ast.ConstStatement {
	// Consume const token
	constToken := p.advance()
	// Expect and consume identifier
	nameToken, ok := p.expect(token.IDENTIFIER)
	if !ok {
		return nil
	}
	name := &ast.Identifier{Token: nameToken, Value: nameToken.Lexeme}
	// if ':' parse type
	if p.check(token.COLON) {
		p.advance()
		typeHint := p.parseTypeAnnotation()
		if p.check(token.EQUAL) {
			p.advance()
			value := p.parseExpression()
			p.consumeSemicolon()
			return &ast.ConstStatement{Token: constToken, Name: name, TypeHint: typeHint, Value: value}
		} else {
			p.errorf("const %q must be initialized at line %d", name.Value, p.currentToken().Line)
			return nil
		}
	} else if p.check(token.EQUAL) { // if '=' parse expression
		p.advance()
		value := p.parseExpression()
		p.consumeSemicolon()
		return &ast.ConstStatement{Token: constToken, Name: name, TypeHint: nil, Value: value}
	} else { // if nothing then error
		p.errorf("uninitialized variable %q must have a type annotation at line %d", name.Value, p.currentToken().Line)
	}
	return nil
}

// Map of assignment operators for easy lookup
var assignOperators = map[token.TokenType]ast.AssignOperator{
	token.EQUAL:         ast.Assign,
	token.PLUS_EQUAL:    ast.PlusAssign,
	token.MINUS_EQUAL:   ast.MinusAssign,
	token.STAR_EQUAL:    ast.StarAssign,
	token.SLASH_EQUAL:   ast.SlashAssign,
	token.PERCENT_EQUAL: ast.PercentAssign,
}

// Parses an assignment statement, which must be in the form of `<identifier> <assignment operators> <expression>`.
func (p *Parser) parseAssignmentStatement() *ast.AssignStatement {
	nameToken, ok := p.expect(token.IDENTIFIER)
	if !ok {
		return nil
	}
	name := &ast.Identifier{Token: nameToken, Value: nameToken.Lexeme}
	opToken := p.currentToken()
	assignOp, ok := assignOperators[opToken.Type]
	if !ok {
		p.errorf("expected an assignment operator but got %q at line %d", p.currentToken().Lexeme, p.currentToken().Line)
		p.advance()
		return nil
	}
	p.advance()
	value := p.parseExpression()
	p.consumeSemicolon()
	return &ast.AssignStatement{Token: opToken, Target: name, Operator: assignOp, Value: value}
}

// temporary stub for expression parsing, just returns an identifier with the current token's lexeme
func (p *Parser) parseExpression() ast.Expression {
	// Temporary stub — returns current token as identifier
	t := p.currentToken()
	p.advance()
	return &ast.Identifier{Token: t, Value: t.Lexeme}
}
