package rules

import "github.com/ab36245/go-pkgs/source"

func NewTokens(bytes []byte) *source.Buffer[Token] {
	return source.NewBuffer[Token](bytes, tokensFinished, tokensProducer)
}

func tokensFinished(t Token) bool {
	return t.IsEnd()
}

func tokensProducer(bytes []byte, output chan<- Token) {
	chars := source.NewChars(bytes)
	char := chars.Current()

	var token Token
	next := func() source.Char {
		token.Span = token.Span.Extend(char.Span)
		return chars.Next()
	}

	for {
		token = Token{
			Span:  char.Span,
			Kind:  InvalidToken,
			value: "",
		}

		switch {
		case char.IsSpace():
			token.Kind = InvalidToken
			char = next()

		case char.Is('/') && chars.Peek(1).Is('/'):
			token.Kind = InvalidToken
			// skip rest of line
			for !char.IsEnd() {
				char = next()
				if char.Is('\n') {
					char = next()
					break
				}
			}

		case char.Is('#'):
			token.Kind = InvalidToken
			// skip rest of line
			for !char.IsEnd() {
				char = next()
				if char.Is('\n') {
					char = next()
					break
				}
			}

		case char.IsEnd():
			token.Kind = EndToken

		case char.IsDigit():
			token.Kind = NumberToken
			for char.IsDigit() {
				token.value += string(char.Rune)
				char = next()
			}

		case char.IsLetter():
			token.Kind = NameToken
			for char.IsLetter() || char.IsDigit() || char.Is('_') || char.Is('-') {
				token.value += string(char.Rune)
				char = next()
			}

		case char.Is('"') || char.Is('\''):
			token.Kind = StringToken
			end := char.Rune
			char = next()
			for !char.IsEnd() {
				if char.Is(end) {
					char = next()
					break
				}
				if char.Is('\\') && chars.Peek(1).Is(end) {
					char = next()
				}
				token.value += string(char.Rune)
				char = next()
			}

		case char.Is('!'):
			char = next()
			if char.Is('=') {
				char = next()
				token.Kind = NotEqualsToken
			} else if char.Is('~') {
				char = next()
				token.Kind = NotMatchesToken
			} else {
				token.Kind = NotToken
			}

		case char.Is('='):
			char = chars.Next()
			if char.Is('=') {
				char = next()
				token.Kind = EqualsToken
			} else if char.Is('~') {
				char = next()
				token.Kind = MatchesToken
			} else {
				token.Kind = EqualsToken
			}

		case char.Is('&') && chars.Peek(1).Is('&'):
			token.Kind = AndToken
			next()
			char = next()

		case char.Is('|') && chars.Peek(1).Is('|'):
			token.Kind = OrToken
			next()
			char = next()

		case char.Is('~'):
			token.Kind = MatchesToken
			char = chars.Next()
			if char.Is('=') || char.Is('~') {
				char = next()
			}

		default:
			token.Kind = CharToken
			token.value = string(char.Rune)
			char = next()
		}

		if !token.IsInvalid() {
			output <- token
			if token.IsEnd() {
				break
			}
		}
	}
}
