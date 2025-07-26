package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ab36245/go-pkgs/source"
)

type Rules []Rule

func (rs Rules) String() string {
	s := "[\n"
	for _, r := range rs {
		for _, l := range strings.Split(r.String(), "\n") {
			s += fmt.Sprintf("  %s\n", l)
		}
	}
	s += "]"
	return s
}

func Load(bytes []byte) (Rules, error) {
	rules, err := parse(bytes)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func parse(bytes []byte) ([]Rule, error) {
	p := &parser{
		tokens: NewTokens(bytes),
	}
	return p.parse()
}

type parser struct {
	tokens *source.Buffer[Token]
}

func (p *parser) parse() ([]Rule, error) {
	rules := []Rule{}
	for !p.atEnd() {
		msg, err := p.parseRule()
		if err != nil {
			return nil, err
		}
		rules = append(rules, msg)
	}
	return rules, nil
}

func (p *parser) parseRule() (Rule, error) {
	rule := Rule{}

	if !p.consumeName("if") {
		return rule, p.expected("if")
	}
	filter, err := p.parseOr()
	if err != nil {
		return rule, err
	}

	if !p.consumeName("then") {
		return rule, p.expected("then")
	}
	actions, err := p.parseActions()
	if err != nil {
		return rule, err
	}

	rule.Filter = filter
	rule.Actions = actions

	return rule, nil
}

func (p *parser) parseOr() (Filter, error) {
	var list []Filter
	for {
		filter, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		list = append(list, filter)
		if !p.consume(OrToken) && !p.consumeName("or") {
			break
		}
	}
	if len(list) == 1 {
		return list[0], nil
	}
	return OrFilter{list}, nil
}

func (p *parser) parseAnd() (Filter, error) {
	var list []Filter
	for {
		filter, err := p.parseField()
		if err != nil {
			return nil, err
		}
		list = append(list, filter)
		if !p.consume(AndToken) && !p.consumeName("and") {
			break
		}
	}
	if len(list) == 1 {
		return list[0], nil
	}
	return AndFilter{list}, nil
}

func (p *parser) parseField() (Filter, error) {
	if p.consumeChar('(') {
		filter, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if !p.consumeChar(')') {
			return nil, p.expected(")")
		}
		return filter, nil
	}

	if !p.at(NameToken) {
		return nil, p.expected("field name")
	}
	field := p.value()
	p.next()

	not := false
	if p.consumeName("not") || p.consume(NotToken) {
		not = true
	}

	operator, err := p.parseOperator(not)
	if err != nil {
		return nil, err
	}

	return FieldFilter{field, operator}, nil
}

func (p *parser) parseOperator(not bool) (Operator, error) {
	switch {
	case p.consumeName("contains"):
		return p.parseContainsOperator(not)

	case p.consumeName("equals"):
		return p.parseEqualsOperator(not)
	case p.consume(EqualsToken):
		return p.parseEqualsOperator(false)
	case p.consume(NotEqualsToken):
		return p.parseEqualsOperator(true)

	case p.consumeName("matches"):
		return p.parseMatchesOperator(not)
	case p.consume(MatchesToken):
		return p.parseMatchesOperator(false)
	case p.consume(NotMatchesToken):
		return p.parseMatchesOperator(true)

	default:
		return nil, p.expected("operator")
	}
}

func (p *parser) parseContainsOperator(not bool) (Operator, error) {
	strings, err := p.parseStringList()
	if err != nil {
		return nil, err
	}
	if not {
		return NotContainsOperator{strings}, nil
	}
	return ContainsOperator{strings}, nil
}

func (p *parser) parseEqualsOperator(not bool) (Operator, error) {
	strings, err := p.parseStringList()
	if err != nil {
		return nil, err
	}
	if not {
		return NotEqualsOperator{strings}, nil
	}
	return EqualsOperator{strings}, nil
}

func (p *parser) parseMatchesOperator(not bool) (Operator, error) {
	regexps, err := p.parseRegexpList()
	if err != nil {
		return nil, err
	}
	if not {
		return NotMatchesOperator{regexps}, nil
	}
	return MatchesOperator{regexps}, nil
}

func (p *parser) parseStringList() ([]string, error) {
	var text []string
	if p.at(StringToken) {
		text = append(text, p.value())
		p.next()
	} else if p.consumeChar('[') {
		for !p.atEnd() {
			if !p.at(StringToken) {
				return nil, p.expected("string")
			}
			text = append(text, p.value())
			p.next()
			p.consumeChar(',')
			if p.consumeChar(']') {
				break
			}
		}
	} else {
		return nil, p.expected("string or list")
	}
	return text, nil
}

func (p *parser) parseRegexpList() ([]*regexp.Regexp, error) {
	strings, err := p.parseStringList()
	if err != nil {
		return nil, err
	}
	var regexps []*regexp.Regexp
	for _, s := range strings {
		re, err := regexp.Compile(s)
		if err != nil {
			return nil, p.expected("regexp")
		}
		regexps = append(regexps, re)
	}
	return regexps, nil
}

func (p *parser) parseActions() ([]Action, error) {
	var list []Action
	for {
		action, err := p.parseAction()
		if err != nil {
			return nil, err
		}
		list = append(list, action)
		if !p.consumeChar(',') {
			break
		}
	}
	return list, nil
}

func (p *parser) parseAction() (Action, error) {
	switch {
	case p.consumeName("add"):
		return p.parseAddAction()
	case p.consumeName("moveto"):
		return p.parseMoveToAction()
	default:
		return nil, p.expected("action")
	}
}

func (p *parser) parseAddAction() (Action, error) {
	if !p.at(StringToken) {
		return nil, p.expected("label")
	}
	label := p.value()
	p.next()
	return AddAction{label}, nil
}

func (p *parser) parseMoveToAction() (Action, error) {
	if !p.at(StringToken) {
		return nil, p.expected("label")
	}
	label := p.value()
	p.next()
	return MoveToAction{label}, nil
}

func (p *parser) expected(want any) error {
	span := p.tokens.Current().Span
	text := fmt.Sprintf("expected %v", want)
	return source.SpanError{Span: span, Text: text}
}

func (p *parser) at(kind TokenKind) bool {
	return p.tokens.Current().Is(kind)
}

func (p *parser) atEnd() bool {
	return p.tokens.Current().IsEnd()
}

func (p *parser) atChar(char rune) bool {
	return p.tokens.Current().IsChar(char)
}

func (p *parser) atName(name string) bool {
	return p.tokens.Current().IsName(name)
}

func (p *parser) _consume(f func(Token) bool) bool {
	if !f(p.tokens.Current()) {
		return false
	}
	p.next()
	return true
}

func (p *parser) consume(kind TokenKind) bool {
	return p._consume(func(t Token) bool {
		return t.Is(kind)
	})
}

func (p *parser) consumeChar(char rune) bool {
	return p._consume(func(t Token) bool {
		return t.IsChar(char)
	})
}

func (p *parser) consumeName(name string) bool {
	return p._consume(func(t Token) bool {
		return t.IsName(name)
	})
}

func (p *parser) next() {
	p.tokens.Next()
}

func (p *parser) value() string {
	return p.tokens.Current().Value()
}
