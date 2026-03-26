package lexer

import (
	"Me/tokens"
)

// Scanner struct
type Scanner struct {
	source  string // The source string from the input file
	start   int    // Starting index
	current int    // Current index
	line    uint   // Tracks the line
}

// Initialize the scanner
func init_scanner(source string) *Scanner {
	return &Scanner{
		source:  source,
		start:   0,
		current: 0,
		line:    1,
	}
}

// Moves the pointer forward and returns the current char
func (s *Scanner) advance() byte {
	if s.current >= len(s.source) {
		return 0
	}
	c := s.source[s.current]
	s.current += 1
	return c
}

// Returns true if we are at the end of the file
func (s *Scanner) is_at_end() bool {
	if s.current >= len(s.source) {
		return true
	}
	return false
}

// Checks the next character
func (s *Scanner) peek() byte {
	if s.is_at_end() {
		return 0
	}
	return s.source[s.current]
}

// Checks the second next character
func (s *Scanner) peek_next() byte {
	if s.current+1 >= len(s.source) {
		return 0
	}
	return s.source[s.current+1]
}

// Skips whitespace
func (s *Scanner) skip_whitespace() {

	for {
		switch s.peek() {
		case ' ', '\t', '\r':
			s.advance()
		case '\n':
			s.line += 1
			s.advance()
		case '/':
			if s.peek_next() == '/' {
				s.advance()
				s.advance()
				for {
					if !s.is_at_end() && s.peek() != '\n' {
						s.advance()
					} else {
						break
					}
				}
			} else {
				return
			}
		default:
			return
		}
	}
}

// Create new tokens
func (s *Scanner) create_token(token_type tokens.TokenType) tokens.Token {
	lexeme := s.source[s.start:s.current]
	return tokens.Token{
		Type:   token_type,
		Lexeme: lexeme,
		Line:   s.line,
	}
}

// Creates a error token for error reporting
func (s *Scanner) error_token(message string) tokens.Token {
	lexeme := message
	return tokens.Token{
		Type:   tokens.ERROR_TOKEN,
		Lexeme: lexeme,
		Line:   s.line,
	}
}

// Helper function to evaluate conditional advances - '!=' , '=='
func (s *Scanner) match(expected byte) bool {
	if s.is_at_end() {
		return false
	}
	if s.source[s.current] != expected {
		return false
	}
	s.current += 1
	return true
}

