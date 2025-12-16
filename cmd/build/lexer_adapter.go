package main

import (
	"github.com/vinayprograms/build/internal/lexer"
)

// tokenAdapter wraps lexer.Token to implement the Token interface.
type tokenAdapter struct {
	tok lexer.Token
}

func (t tokenAdapter) TokenType() string     { return t.tok.Type.String() }
func (t tokenAdapter) TokenLiteral() string  { return t.tok.Literal }
func (t tokenAdapter) TokenLocation() string { return t.tok.Location.String() }
func (t tokenAdapter) IsEOF() bool           { return t.tok.Type == lexer.EOF }
func (t tokenAdapter) IsError() bool         { return t.tok.Type == lexer.ERROR }

// lexerAdapter wraps lexer.Lexer to implement the Lexer interface.
type lexerAdapter struct {
	lex *lexer.Lexer
}

func (l *lexerAdapter) NextToken() Token {
	return tokenAdapter{tok: l.lex.NextToken()}
}

// NewLexer creates a new Lexer for the given source.
// This is the LexerFactory implementation using the internal/lexer package.
func NewLexer(file, input string) Lexer {
	return &lexerAdapter{lex: lexer.New(file, input)}
}
