package rules

import (
	"fmt"
	"regexp"
	"strings"
)

type Operator interface {
	Match(any) bool
	String() string
}

type ContainsOperator struct {
	Strings []string
}

func (o ContainsOperator) Match(v any) bool {
	match := func(vs ...string) bool {
		for _, v := range vs {
			v := strings.ToLower(v)
			for _, ov := range o.Strings {
				ov := strings.ToLower(ov)
				if strings.Contains(v, ov) {
					return true
				}
			}
		}
		return false
	}
	switch v := v.(type) {
	case string:
		return match(v)
	case []string:
		return match(v...)
	default:
		return false
	}
}

func (o ContainsOperator) String() string {
	return "contains " + do(o.Strings)
}

type NotContainsOperator struct {
	Strings []string
}

func (o NotContainsOperator) Match(v any) bool {
	match := func(vs ...string) bool {
		for _, v := range vs {
			v := strings.ToLower(v)
			for _, ov := range o.Strings {
				ov := strings.ToLower(ov)
				if strings.Contains(v, ov) {
					return false
				}
			}
		}
		return true
	}
	switch v := v.(type) {
	case string:
		return match(v)
	case []string:
		return match(v...)
	default:
		return false
	}
}

func (o NotContainsOperator) String() string {
	return "not contains " + do(o.Strings)
}

type EqualsOperator struct {
	Strings []string
}

func (o EqualsOperator) Match(v any) bool {
	match := func(vs ...string) bool {
		for _, v := range vs {
			v := strings.ToLower(v)
			for _, ov := range o.Strings {
				ov := strings.ToLower(ov)
				if v == ov {
					return true
				}
			}
		}
		return false
	}
	switch v := v.(type) {
	case string:
		return match(v)
	case []string:
		return match(v...)
	default:
		return false
	}
}

func (o EqualsOperator) String() string {
	return "equals " + do(o.Strings)
}

type NotEqualsOperator struct {
	Strings []string
}

func (o NotEqualsOperator) Match(v any) bool {
	match := func(vs ...string) bool {
		for _, v := range vs {
			v := strings.ToLower(v)
			for _, ov := range o.Strings {
				ov := strings.ToLower(ov)
				if v == ov {
					return false
				}
			}
		}
		return true
	}
	switch v := v.(type) {
	case string:
		return match(v)
	case []string:
		return match(v...)
	default:
		return false
	}
}

func (o NotEqualsOperator) String() string {
	return "not equals " + do(o.Strings)
}

type MatchesOperator struct {
	Regexps []*regexp.Regexp
}

func (o MatchesOperator) Match(v any) bool {
	match := func(vs ...string) bool {
		for _, v := range vs {
			v := []byte(v)
			for _, ov := range o.Regexps {
				if ov.Match(v) {
					return true
				}
			}
		}
		return false
	}
	switch v := v.(type) {
	case string:
		return match(v)
	case []string:
		return match(v...)
	default:
		return false
	}
}

func (o MatchesOperator) String() string {
	return "matches " + do(o.Regexps)
}

type NotMatchesOperator struct {
	Regexps []*regexp.Regexp
}

func (o NotMatchesOperator) Match(v any) bool {
	match := func(vs ...string) bool {
		for _, v := range vs {
			v := []byte(v)
			for _, ov := range o.Regexps {
				if ov.Match(v) {
					return false
				}
			}
		}
		return true
	}
	switch v := v.(type) {
	case string:
		return match(v)
	case []string:
		return match(v...)
	default:
		return false
	}
}

func (o NotMatchesOperator) String() string {
	return "not matches " + do(o.Regexps)
}

func do[T any](values []T) string {
	if len(values) == 0 {
		return "[]"
	}
	s := "[\n"
	for _, value := range values {
		s += fmt.Sprintf("  %v\n", value)
	}
	s += "]"
	return s
}
