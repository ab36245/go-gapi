package rules

import (
	"fmt"

	"github.com/ab36245/go-pkgs/source"
)

type Token struct {
	source.Span
	Kind  TokenKind
	value string
}

func (t Token) Is(kind TokenKind) bool {
	return t.Kind == kind
}

func (t Token) IsEnd() bool {
	return t.Is(EndToken)
}

func (t Token) IsInvalid() bool {
	return t.Is(InvalidToken)
}

func (t Token) IsChar(r rune) bool {
	return t.Is(CharToken) && t.value == string(r)
}

func (t Token) IsName(name string) bool {
	return t.Is(NameToken) && t.IsValue(name)
}

func (t Token) IsValue(s string) bool {
	return t.value == s
}

func (t Token) Value() string {
	return t.value
}

func (t Token) String() string {
	switch t.Kind {
	case CharToken:
		return t.value
	case EndToken:
		return "<EOF>"
	case NameToken:
		return t.value
	case NumberToken:
		return t.value
	case StringToken:
		return fmt.Sprintf("%q", t.value)
	default:
		return fmt.Sprintf("unknown token (kind %d)", t.Kind)
	}
}

type TokenKind int

const (
	InvalidToken TokenKind = iota
	EndToken

	AndToken

	CharToken

	EqualsToken
	NotEqualsToken

	MatchesToken
	NotMatchesToken

	NameToken

	NotToken

	NumberToken

	OrToken

	StringToken
)
