package rules

import (
	"fmt"
	"strings"
)

type Filter interface {
	Match(Getter) bool
	String() string
}

type FieldFilter struct {
	Field    string
	Operator Operator
}

func (f FieldFilter) Match(getter Getter) bool {
	return f.Operator.Match(getter(f.Field))
}

func (f FieldFilter) String() string {
	return fmt.Sprintf("%s %s", f.Field, f.Operator)
}

type AndFilter struct {
	Filters []Filter
}

func (f AndFilter) Match(getter Getter) bool {
	for _, f := range f.Filters {
		if !f.Match(getter) {
			return false
		}
	}
	return true
}

func (f AndFilter) String() string {
	s := "and [\n"
	for _, f := range f.Filters {
		for _, l := range strings.Split(f.String(), "\n") {
			s += fmt.Sprintf("  %s\n", l)
		}
	}
	s += "]"
	return s
}

type OrFilter struct {
	Filters []Filter
}

func (f OrFilter) Match(getter Getter) bool {
	for _, f := range f.Filters {
		if f.Match(getter) {
			return true
		}
	}
	return false
}

func (f OrFilter) String() string {
	s := "or [\n"
	for _, f := range f.Filters {
		for _, l := range strings.Split(f.String(), "\n") {
			s += fmt.Sprintf("  %s\n", l)
		}
	}
	s += "]"
	return s
}
